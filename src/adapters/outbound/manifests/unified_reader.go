package manifests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
	"github.com/datosbolivia/open-knowledge-format/src/ports/outbound"
)

// UnifiedMetadataReader coordinates heterogeneous manifest parsers (JSON, YAML, Markdown, LLMs.txt)
// and implements both UnifiedMetadataReaderPort and TriadResolverPort.
type UnifiedMetadataReader struct {
	parsers    []outbound.ManifestParserPort
	httpClient *http.Client
	catalogURL string
}

// NewUnifiedMetadataReader initializes the reader with standard parsers.
func NewUnifiedMetadataReader(catalogURL string, timeout time.Duration) *UnifiedMetadataReader {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	r := &UnifiedMetadataReader{
		httpClient: &http.Client{Timeout: timeout},
		catalogURL: catalogURL,
	}

	// Register built-in standard parsers
	r.RegisterParser(NewDataPackageJSONParser())
	r.RegisterParser(NewDataPackageYAMLParser())
	r.RegisterParser(NewOKFMarkdownParser())
	r.RegisterParser(NewLLMSTxtParser())

	return r
}

// RegisterParser dynamically adds or extends support for manifest formats (DCAT, RO-Crate, etc.).
func (r *UnifiedMetadataReader) RegisterParser(parser outbound.ManifestParserPort) {
	r.parsers = append(r.parsers, parser)
}

// ReadPackage reads and parses a manifest file from a local path or HTTP URL.
func (r *UnifiedMetadataReader) ReadPackage(ctx context.Context, uriOrPath string) (*domain.PackageManifest, error) {
	data, err := r.readData(ctx, uriOrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest from %s: %w", uriOrPath, err)
	}

	for _, p := range r.parsers {
		if p.CanParse(uriOrPath, data) {
			return p.ParseManifest(ctx, data, uriOrPath)
		}
	}

	return nil, fmt.Errorf("no manifest parser found capable of parsing %s", uriOrPath)
}

// ReadNode resolves a node directory, specific .md concept document, or bare slug/triad
// and returns a unified DataProduct with resources and synthesized Markdown contract tables.
func (r *UnifiedMetadataReader) ReadNode(ctx context.Context, uriOrPath string) (*domain.DataProduct, error) {
	resolvedTarget := r.resolveSlugOrTriadTarget(ctx, strings.TrimSpace(uriOrPath))
	isURL := strings.HasPrefix(resolvedTarget, "http://") || strings.HasPrefix(resolvedTarget, "https://")
	baseDir := resolvedTarget
	var indexPath string
	isSpecificMarkdownDoc := false

	if isURL {
		lower := strings.ToLower(resolvedTarget)
		if strings.HasSuffix(lower, ".md") && !strings.HasSuffix(lower, "/index.md") {
			isSpecificMarkdownDoc = true
			indexPath = resolvedTarget
		}
		if strings.HasSuffix(baseDir, ".md") || strings.HasSuffix(baseDir, ".json") || strings.HasSuffix(baseDir, ".yaml") || strings.HasSuffix(baseDir, ".yml") {
			if idx := strings.LastIndex(baseDir, "/"); idx != -1 {
				baseDir = baseDir[:idx]
			}
		}
		baseDir = strings.TrimSuffix(baseDir, "/")
		if !isSpecificMarkdownDoc {
			indexPath = baseDir + "/index.md"
		}
	} else {
		lower := strings.ToLower(resolvedTarget)
		if strings.HasSuffix(lower, ".md") && filepath.Base(lower) != "index.md" {
			isSpecificMarkdownDoc = true
			indexPath = resolvedTarget
		}
		if strings.HasSuffix(baseDir, ".md") || strings.HasSuffix(baseDir, ".json") || strings.HasSuffix(baseDir, ".yaml") || strings.HasSuffix(baseDir, ".yml") {
			baseDir = filepath.Dir(baseDir)
		}
		if !isSpecificMarkdownDoc {
			indexPath = filepath.Join(baseDir, "index.md")
		}
	}

	// 1. Read target .md (or index.md) for frontmatter and descriptions
	var manifest domain.Manifest
	var description string
	var rawContent string

	if data, err := r.readData(ctx, indexPath); err == nil {
		rawContent = string(data)
		m, desc, err := ParseOKFFrontmatter(data)
		if err == nil {
			manifest = m
			description = desc
		}
	} else if strings.HasSuffix(resolvedTarget, ".md") && resolvedTarget != indexPath {
		if data, err := r.readData(ctx, resolvedTarget); err == nil {
			rawContent = string(data)
			m, desc, err := ParseOKFFrontmatter(data)
			if err == nil {
				manifest = m
				description = desc
			}
		}
	}

	// 2. Discover package manifest and OpenAPI contracts guided by frontmatter or standard files
	var resources []domain.Resource
	var candidateFiles []string
	rawOpenAPIs := make(map[string][]byte)

	resolveContractRef := func(ref string) string {
		if isURL {
			if !strings.HasPrefix(ref, "http://") && !strings.HasPrefix(ref, "https://") {
				return fmt.Sprintf("%s/%s", baseDir, strings.TrimPrefix(ref, "./"))
			}
			return ref
		}
		if !filepath.IsAbs(ref) && !strings.HasPrefix(ref, "http://") && !strings.HasPrefix(ref, "https://") {
			return filepath.Join(baseDir, strings.TrimPrefix(ref, "./"))
		}
		return ref
	}

	for _, c := range manifest.Contracts {
		targetP := strings.TrimSpace(c.Path)
		if targetP == "" {
			targetP = strings.TrimSpace(c.Ref)
		}
		if targetP != "" {
			fullP := resolveContractRef(targetP)
			cType := strings.ToLower(strings.TrimSpace(c.Type))
			lowerP := strings.ToLower(fullP)
			if cType == "openapi" || strings.Contains(lowerP, "openapi.") || strings.Contains(lowerP, "swagger.") {
				if oaBytes, err := r.readData(ctx, fullP); err == nil {
					rawOpenAPIs[fullP] = oaBytes
				}
			} else {
				candidateFiles = append(candidateFiles, fullP)
			}
		}
	}

	if !isSpecificMarkdownDoc {
		if isURL {
			candidateFiles = append(candidateFiles,
				fmt.Sprintf("%s/datapackage.json", baseDir),
				fmt.Sprintf("%s/datapackage.yaml", baseDir),
				fmt.Sprintf("%s/datapackage.yml", baseDir),
			)
		} else {
			candidateFiles = append(candidateFiles,
				filepath.Join(baseDir, "datapackage.json"),
				filepath.Join(baseDir, "datapackage.yaml"),
				filepath.Join(baseDir, "datapackage.yml"),
			)
		}

		for _, targetP := range candidateFiles {
			if pkg, err := r.ReadPackage(ctx, targetP); err == nil {
				resources = pkg.Resources
				if manifest.Title == "" && pkg.Title != "" {
					manifest.Title = pkg.Title
				}
				break
			}
		}
	}

	if rawContent == "" && len(resources) == 0 && len(rawOpenAPIs) == 0 {
		return nil, fmt.Errorf("could not read node content or manifest from %s", uriOrPath)
	}

	if manifest.Title == "" {
		if isSpecificMarkdownDoc {
			base := filepath.Base(resolvedTarget)
			manifest.Title = strings.TrimSuffix(base, filepath.Ext(base))
		} else if isURL {
			parts := strings.Split(baseDir, "/")
			manifest.Title = parts[len(parts)-1]
		} else {
			manifest.Title = filepath.Base(baseDir)
		}
	}
	if manifest.Type == "" {
		if isSpecificMarkdownDoc {
			manifest.Type = "concept"
		} else {
			manifest.Type = "dataset"
		}
	}
	if len(manifest.Dimensions) == 0 {
		manifest.Dimensions = []string{"core"}
	}

	cleanDesc := StripOJSBlocks(description)
	dp, err := domain.NewDataProduct(resolvedTarget, manifest, cleanDesc, resources)
	if err != nil {
		return nil, err
	}

	if synthMD := SynthesizeContractMarkdown(dp, rawOpenAPIs); synthMD != "" && !strings.Contains(dp.Description, "## Contrato de Datos y Esquema Semántico") {
		if strings.TrimSpace(dp.Description) != "" {
			dp.Description = strings.TrimSpace(dp.Description) + "\n\n" + synthMD
		} else {
			dp.Description = synthMD
		}
	}

	dp.RawContent = rawContent
	return dp, nil
}

// resolveSlugOrTriadTarget maps a bare slug ("bcb-semanal") or relative node subpath
// ("air_quality/concepts/air_quality_index.md" or "air_quality:concepts/air_quality_index.md")
// to a local bundle directory or remote catalog URL when the path does not exist directly on disk.
func (r *UnifiedMetadataReader) resolveSlugOrTriadTarget(ctx context.Context, uriOrPath string) string {
	if uriOrPath == "" || strings.HasPrefix(uriOrPath, "http://") || strings.HasPrefix(uriOrPath, "https://") || strings.HasPrefix(uriOrPath, "file://") {
		return uriOrPath
	}
	if _, err := os.Stat(uriOrPath); err == nil {
		return uriOrPath
	}

	// Normalize colon separators in non-URL targets (e.g. "air_quality:concepts/air_quality_index.md" -> "air_quality/concepts/air_quality_index.md")
	normalized := strings.ReplaceAll(uriOrPath, ":", "/")
	localPrefixes := []string{
		"bundles",
		filepath.Join("..", "bundles"),
		filepath.Join("examples", "bundles"),
		filepath.Join("knowledge", "nodes"),
		"nodes",
		filepath.Join("..", "catalogo-datamesh", "knowledge", "nodes"),
	}
	for _, pref := range localPrefixes {
		cand := filepath.Join(pref, normalized)
		if _, err := os.Stat(cand); err == nil {
			return cand
		}
	}

	if r.catalogURL != "" && (strings.HasPrefix(r.catalogURL, "http://") || strings.HasPrefix(r.catalogURL, "https://")) {
		origin := r.catalogURL
		if idx := strings.LastIndex(origin, "/llms"); idx != -1 {
			origin = origin[:idx]
		} else {
			origin = strings.TrimSuffix(origin, "/")
		}

		// Check if normalized has a subpath (e.g. "air_quality/concepts/air_quality_index.md")
		parts := strings.SplitN(normalized, "/", 2)
		slug := parts[0]
		subpath := ""
		if len(parts) == 2 {
			subpath = parts[1]
		}

		if cat, err := r.ReadCatalog(ctx, r.catalogURL); err == nil {
			cleanSlug := strings.ToLower(slug)
			slugified := domain.Slugify(slug)
			for _, entry := range cat.Entries {
				u := strings.ToLower(entry.URI)
				rUrl := strings.ToLower(entry.ResolvedURL)
				t := strings.ToLower(entry.Title)
				if strings.Contains(u, "/"+cleanSlug+"/") || strings.HasSuffix(u, "/"+cleanSlug) ||
					strings.Contains(rUrl, "/"+cleanSlug+"/") || t == cleanSlug || domain.Slugify(t) == slugified {
					baseResolved := entry.ResolvedURL
					if strings.HasSuffix(baseResolved, "/index.md") {
						baseResolved = strings.TrimSuffix(baseResolved, "/index.md")
					}
					if subpath != "" {
						return strings.TrimSuffix(baseResolved, "/") + "/" + subpath
					}
					return entry.ResolvedURL
				}
			}
		}

		// Direct fallback pattern on catalog origin
		if strings.HasSuffix(normalized, ".md") {
			return fmt.Sprintf("%s/raw/nodes/%s", origin, normalized)
		}
		return fmt.Sprintf("%s/raw/nodes/%s/index.md", origin, normalized)
	}

	return uriOrPath
}

// ReadCatalog fetches and parses a federated catalog (llms.txt).
func (r *UnifiedMetadataReader) ReadCatalog(ctx context.Context, uriOrPath string) (*domain.Catalog, error) {
	target := uriOrPath
	if target == "" {
		target = r.catalogURL
	}
	data, err := r.readData(ctx, target)
	if err != nil {
		return nil, err
	}

	parser := NewLLMSTxtParser()
	return parser.ParseCatalog(data, target)
}

// ResolveTriad resolves a canonical 'catalogo:dataset:resource', Sovereign DNS '<dataset>.<resource>@<catalog>',
// or 'okf://catalog/dataset/resource' triad to a physical URL and format, preserving #fragment identifiers.
func (r *UnifiedMetadataReader) ResolveTriad(ctx context.Context, triad domain.ResourceTriad) (*domain.ResolvedResource, error) {
	dataset := triad.Dataset
	resourceName := triad.Resource

	applyFragment := func(resPath string, resFormat string) *domain.ResolvedResource {
		finalURI := resPath
		finalFmt := resFormat
		if strings.TrimSpace(triad.Fragment) != "" && !strings.Contains(finalURI, "#") {
			frag := strings.TrimPrefix(strings.TrimSpace(triad.Fragment), "#")
			finalURI = finalURI + "#" + frag
			if extFmt := inferFormatFromPath(frag); extFmt != "" {
				finalFmt = extFmt
			}
		} else if strings.Contains(finalURI, "#") {
			parts := strings.SplitN(finalURI, "#", 2)
			if extFmt := inferFormatFromPath(parts[1]); extFmt != "" {
				finalFmt = extFmt
			}
		}
		return &domain.ResolvedResource{
			Triad:       triad,
			PhysicalURI: finalURI,
			Format:      finalFmt,
		}
	}

	matchResource := func(resources []domain.Resource, target string) (*domain.Resource, bool) {
		norm := strings.ToLower(strings.TrimSpace(target))
		slug := domain.Slugify(target)
		for i := range resources {
			res := &resources[i]
			rNorm := strings.ToLower(strings.TrimSpace(res.Name))
			if strings.EqualFold(res.Name, target) || domain.Slugify(res.Name) == slug || (len(slug) > 3 && strings.Contains(rNorm, slug)) {
				return res, true
			}
		}
		_ = norm
		return nil, false
	}

	// 1. Search candidate local node and bundle directories via ReadNode (contracts-guided)
	candidateDirs := []string{
		filepath.Join("bundles", dataset),
		filepath.Join("..", "bundles", dataset),
		filepath.Join("knowledge", "nodes", dataset),
		filepath.Join("nodes", dataset),
		filepath.Join("..", "catalogo-datamesh", "knowledge", "nodes", dataset),
	}

	for _, dir := range candidateDirs {
		if dp, err := r.ReadNode(ctx, dir); err == nil && len(dp.Resources) > 0 {
			if matched, ok := matchResource(dp.Resources, resourceName); ok {
				return applyFragment(matched.Path, matched.Format), nil
			}
			// If composite archive has 1 zip resource and triad specifies a #fragment, resolve to that zip + fragment
			if strings.TrimSpace(triad.Fragment) != "" && len(dp.Resources) > 0 {
				for _, res := range dp.Resources {
					if res.Format == "zip" || strings.HasSuffix(strings.ToLower(res.Path), ".zip") {
						return applyFragment(res.Path, res.Format), nil
					}
				}
			}
		}
	}

	// 2. Search catalog entries via HTTP discovery
	if r.catalogURL != "" {
		if cat, err := r.ReadCatalog(ctx, r.catalogURL); err == nil {
			cleanDs := strings.ToLower(dataset)
			slugDs := domain.Slugify(dataset)
			for _, entry := range cat.Entries {
				u := strings.ToLower(entry.URI)
				rUrl := strings.ToLower(entry.ResolvedURL)
				t := strings.ToLower(entry.Title)
				if strings.Contains(u, cleanDs) || strings.Contains(rUrl, cleanDs) || t == cleanDs || domain.Slugify(t) == slugDs {
					if dp, err := r.ReadNode(ctx, entry.ResolvedURL); err == nil && len(dp.Resources) > 0 {
						if matched, ok := matchResource(dp.Resources, resourceName); ok {
							return applyFragment(matched.Path, matched.Format), nil
						}
					}
				}
			}
		}
	}

	// 3. Fallback to resource name directly
	return applyFragment(resourceName, inferFormatFromPath(resourceName)), nil
}

func (r *UnifiedMetadataReader) readData(ctx context.Context, target string) ([]byte, error) {
	if strings.HasPrefix(target, "file://") {
		return os.ReadFile(strings.TrimPrefix(target, "file://"))
	}
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		return os.ReadFile(target)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "datamesh-sdk/0.2 (Go-Core Unified Reader)")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(resp.Body)
}
