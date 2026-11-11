package usecases

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
	"github.com/datosbolivia/open-knowledge-format/src/ports/inbound"
	"github.com/datosbolivia/open-knowledge-format/src/ports/outbound"
)

// AsIsIngestUseCase orchestrates the dual discovery path defined in ADR 0013:
// conformant well-known first, then an as-is heuristic probe (llms.txt,
// datapackage.json/yaml/yml, index.md OKF), always converting
// the source into a canonical ODKF DataProduct; it also exposes ODKF export
// converters toward Frictionless DataPackage and W3C DCAT.
type AsIsIngestUseCase struct {
	wellKnown     outbound.WellKnownResolverPort
	reader        outbound.UnifiedMetadataReaderPort
	authProviders map[string]outbound.AuthProviderPort
}

var _ inbound.InteroperabilityServicePort = (*AsIsIngestUseCase)(nil)
var _ inbound.DiscoveryServicePort = (*AsIsIngestUseCase)(nil)

// NewAsIsIngestUseCase constructs AsIsIngestUseCase; all ports are optional so the
// probe chain degrades gracefully when a port is not configured.
func NewAsIsIngestUseCase(
	wellKnown outbound.WellKnownResolverPort,
	reader outbound.UnifiedMetadataReaderPort,
	authProviders map[string]outbound.AuthProviderPort,
) *AsIsIngestUseCase {
	if authProviders == nil {
		authProviders = make(map[string]outbound.AuthProviderPort)
	}
	return &AsIsIngestUseCase{
		wellKnown:     wellKnown,
		reader:        reader,
		authProviders: authProviders,
	}
}

// DiscoverWellKnown probes for /.well-known/datamesh.json to inspect required auth and catalog specs.
func (uc *AsIsIngestUseCase) DiscoverWellKnown(ctx context.Context, targetURL string) (*domain.WellKnownDiscovery, error) {
	if uc.wellKnown == nil {
		return nil, errors.New("well-known resolver port is not configured")
	}
	return uc.wellKnown.FetchWellKnown(ctx, targetURL)
}

// Ingest resolves any target endpoint into a canonical ODKF DataProduct (ADR 0013).
// If every probe fails the last error is returned with per-stage context.
func (uc *AsIsIngestUseCase) Ingest(ctx context.Context, target, authType string) (*domain.DataProduct, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, errors.New("ingest target cannot be empty")
	}

	errs := []string{}

	if uc.wellKnown != nil {
		if wk, err := uc.wellKnown.FetchWellKnown(ctx, target); err == nil && wk != nil {
			if dp, err := uc.ingestConformant(ctx, target, wk, authType); err == nil {
				return dp, nil
			} else {
				errs = append(errs, fmt.Sprintf("conformant: %v", err))
			}
		}
	}

	if uc.reader != nil {
		if dp, err := uc.reader.ReadNode(ctx, target); err == nil {
			return dp, nil
		} else {
			errs = append(errs, fmt.Sprintf("as-is reader: %v", err))
		}
	}

	if len(errs) == 0 {
		return nil, errors.New("no ingestion port is configured for target " + target)
	}
	return nil, fmt.Errorf("cannot ingest %q as ODKF: %s", target, strings.Join(errs, "; "))
}

// ingestConformant dispatches on the well-known catalog type (CU-14) to the idiomatic port.
func (uc *AsIsIngestUseCase) ingestConformant(ctx context.Context, target string, wk *domain.WellKnownDiscovery, authType string) (*domain.DataProduct, error) {
	switch strings.ToLower(wk.Catalog.Type) {
	case "okf", "git", "dcat":
		if uc.reader == nil {
			return nil, errors.New("reader port required for catalog type " + wk.Catalog.Type)
		}
		return uc.reader.ReadNode(ctx, target)
	case "ckan":
		return nil, errors.New("ckan is an external software service; use a CKAN service connector or script (connectors/ckan/) rather than core ingestion")
	default:
		return nil, fmt.Errorf("unsupported conformant catalog type %q", wk.Catalog.Type)
	}
}

// ExportFrictionlessDataPackage serializes an ODKF DataProduct as standard Frictionless datapackage JSON.
func (uc *AsIsIngestUseCase) ExportFrictionlessDataPackage(ctx context.Context, dp *domain.DataProduct) ([]byte, error) {
	if dp == nil {
		return nil, errors.New("data product is nil; cannot export")
	}
	return dp.ToCKANDataPackageJSON()
}

// ExportCKANDataPackage serializes an ODKF DataProduct as CKAN/Frictionless datapackage JSON.
func (uc *AsIsIngestUseCase) ExportCKANDataPackage(ctx context.Context, dp *domain.DataProduct) ([]byte, error) {
	return uc.ExportFrictionlessDataPackage(ctx, dp)
}

// ExportDCAT serializes an ODKF DataProduct as a W3C DCAT-AP dataset JSON-LD document.
func (uc *AsIsIngestUseCase) ExportDCAT(ctx context.Context, dp *domain.DataProduct) ([]byte, error) {
	if dp == nil {
		return nil, errors.New("data product is nil; cannot export")
	}
	return dp.ToDCATJSON()
}

// dataProductFromManifest elevates a normalized PackageManifest (información, ADR 0014
// layer 2) into a canonical ODKF DataProduct (conocimiento, layer 3), applying domain
// invariants (título, tipo, dimensiones) at the boundary.
func dataProductFromManifest(pm *domain.PackageManifest) (*domain.DataProduct, error) {
	dims, mtype := []string{"value"}, "dataset"
	if raw, ok := pm.Metadata["dimensions"]; ok {
		switch d := raw.(type) {
		case []interface{}:
			var dd []string
			for _, v := range d {
				if s, ok := v.(string); ok && s != "" {
					dd = append(dd, s)
				}
			}
			if len(dd) > 0 {
				dims = dd
			}
		case []string:
			if len(d) > 0 {
				dims = d
			}
		case string:
			if d != "" {
				dims = []string{d}
			}
		}
	}
	if t, ok := pm.Metadata["type"].(string); ok && t != "" {
		mtype = t
	}

	id := pm.Name
	if id == "" {
		id = domain.Slugify(pm.Title)
	}
	manifest := domain.Manifest{
		Type:       mtype,
		Title:      pm.Title,
		Dimensions: dims,
		Lineage: domain.Lineage{
			Version: pm.Version,
			Sources: []string{},
		},
	}
	return domain.NewDataProduct(id, manifest, pm.Description, pm.Resources)
}
