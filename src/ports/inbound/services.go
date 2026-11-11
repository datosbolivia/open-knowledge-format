package inbound

import (
	"context"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
)

// CatalogServicePort defines inbound operations for discovering and querying sovereign data catalogs.
type CatalogServicePort interface {
	DiscoverAll(ctx context.Context) (*domain.Catalog, error)
	Discover(ctx context.Context, catalogURL string) (*domain.Catalog, error)
	Search(ctx context.Context, catalogURL, keyword string) ([]domain.CatalogEntry, error)
}

// DataProductServicePort defines inbound operations for resolving and validating Data Products.
type DataProductServicePort interface {
	Resolve(ctx context.Context, uriOrURL string) (*domain.DataProduct, error)
	Validate(ctx context.Context, dp *domain.DataProduct) error
}

// QueryServicePort defines inbound operations for executing queries against data resources.
type QueryServicePort interface {
	Query(ctx context.Context, req domain.QueryRequest) (*domain.QueryResult, error)
	ExecuteSQL(ctx context.Context, sqlQuery string) (*domain.QueryResult, error)
}

// ConfigServicePort defines inbound operations for obtaining active configuration settings.
type ConfigServicePort interface {
	GetConfig() domain.Config
}

// PublicationServicePort defines inbound operations for publishing datasets and ODKF bundles across platforms.
type PublicationServicePort interface {
	Publish(ctx context.Context, datasetPath string, targets []string, options map[string]interface{}) ([]domain.PublicationTargetResult, error)
}

// SemanticServicePort defines inbound operations for semantic concept alignment and mapping.
type SemanticServicePort interface {
	AlignSemantics(ctx context.Context, manifest *domain.PackageManifest) ([]domain.SemanticFieldMapping, error)
	GenerateCaseExpression(ctx context.Context, mapping domain.SemanticFieldMapping, tablePrefix string) string
}

// DiscoveryServicePort defines inbound operations for well-known catalog discovery (RFC 8615).
type DiscoveryServicePort interface {
	DiscoverWellKnown(ctx context.Context, targetURL string) (*domain.WellKnownDiscovery, error)
}

// LineageServicePort defines inbound operations for dataset provenance and derivation DAG analysis.
type LineageServicePort interface {
	BuildGraph(ctx context.Context, searchDirs []string, catalogURL string) (*domain.LineageGraph, error)
	GetLineage(ctx context.Context, datasetIDOrPath string, searchDirs []string, catalogURL string) (map[string]interface{}, error)
}

// InteroperabilityServicePort defines inbound operations for as-is ingestion of heterogeneous
// metadata into canonical ODKF DataProducts (ADR 0013) and export from ODKF to open
// standards (Frictionless DataPackage, W3C DCAT).
type InteroperabilityServicePort interface {
	Ingest(ctx context.Context, target string, authType string) (*domain.DataProduct, error)
	ExportFrictionlessDataPackage(ctx context.Context, dp *domain.DataProduct) ([]byte, error)
	ExportCKANDataPackage(ctx context.Context, dp *domain.DataProduct) ([]byte, error)
	ExportDCAT(ctx context.Context, dp *domain.DataProduct) ([]byte, error)
}


