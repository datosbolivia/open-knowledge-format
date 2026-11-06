package manifests

import (
	"context"
	"fmt"
	"strings"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
	"gopkg.in/yaml.v3"
)

// DataPackageYAMLParser parses Frictionless DataPackage manifests formatted in YAML (.yaml and .yml).
type DataPackageYAMLParser struct{}

func NewDataPackageYAMLParser() *DataPackageYAMLParser {
	return &DataPackageYAMLParser{}
}

func (p *DataPackageYAMLParser) Name() string {
	return "datapackage-yaml"
}

func (p *DataPackageYAMLParser) CanParse(filenameOrURL string, data []byte) bool {
	lower := strings.ToLower(filenameOrURL)
	if strings.HasSuffix(lower, "datapackage.yaml") || strings.HasSuffix(lower, "datapackage.yml") {
		return true
	}
	content := string(data)
	return strings.Contains(content, "resources:") && (strings.Contains(content, "- name:") || strings.Contains(content, "  name:"))
}

type rawYAMLSchemaField struct {
	Name        string                 `yaml:"name" json:"name"`
	Type        string                 `yaml:"type" json:"type"`
	Title       string                 `yaml:"title" json:"title"`
	Description string                 `yaml:"description" json:"description"`
	RDFType     string                 `yaml:"rdfType" json:"rdfType"`
	Concept     string                 `yaml:"concept" json:"concept"`
	ConceptRef  string                 `yaml:"concept_ref" json:"concept_ref"`
	OKFConcept  string                 `yaml:"okf:concept" json:"okf:concept"`
	Unit        string                 `yaml:"unit" json:"unit"`
	Categories  interface{}            `yaml:"categories" json:"categories"`
	ValueMap    map[string]string      `yaml:"value_mapping" json:"value_mapping"`
	Custom      map[string]interface{} `yaml:"custom" json:"custom"`
}

type rawYAMLResource struct {
	Name        string `yaml:"name" json:"name"`
	Path        string `yaml:"path" json:"path"`
	URL         string `yaml:"url" json:"url"`
	Format      string `yaml:"format" json:"format"`
	Description string `yaml:"description" json:"description"`
	Hash        string `yaml:"hash" json:"hash"`
	HashSHA256  string `yaml:"hash_sha256" json:"hash_sha256"`
	Bytes       int64  `yaml:"bytes" json:"bytes"`
	BytesCount  int64  `yaml:"bytes_count" json:"bytes_count"`
	Schema      struct {
		Fields []rawYAMLSchemaField `yaml:"fields" json:"fields"`
	} `yaml:"schema" json:"schema"`
	InternalResources []rawJSONInternalSubRes `yaml:"internal_resources" json:"internal_resources"`
	Subresources      []rawJSONInternalSubRes `yaml:"subresources" json:"subresources"`
}

type rawYAMLDataPackage struct {
	Name        string            `yaml:"name" json:"name"`
	Title       string            `yaml:"title" json:"title"`
	Description string            `yaml:"description" json:"description"`
	Version     string            `yaml:"version" json:"version"`
	Resources   []rawYAMLResource `yaml:"resources" json:"resources"`
}

func convertRawFields(rawFields []rawYAMLSchemaField) []domain.ResourceField {
	if len(rawFields) == 0 {
		return nil
	}
	out := make([]domain.ResourceField, 0, len(rawFields))
	for _, rf := range rawFields {
		if strings.TrimSpace(rf.Name) == "" {
			continue
		}
		concept := strings.TrimSpace(rf.Concept)
		if concept == "" {
			concept = strings.TrimSpace(rf.ConceptRef)
		}
		if concept == "" {
			concept = strings.TrimSpace(rf.OKFConcept)
		}
		var customMap map[string]string
		if len(rf.Custom) > 0 {
			customMap = make(map[string]string, len(rf.Custom))
			for k, v := range rf.Custom {
				strVal := fmt.Sprintf("%v", v)
				customMap[k] = strVal
				if concept == "" && (k == "okf:concept" || k == "concept") {
					concept = strVal
				}
			}
		}

		var cats []string
		switch cv := rf.Categories.(type) {
		case []interface{}:
			for _, item := range cv {
				switch iv := item.(type) {
				case string:
					cats = append(cats, iv)
				case map[string]interface{}:
					if val, ok := iv["value"]; ok {
						cats = append(cats, fmt.Sprintf("%v", val))
					} else if lbl, ok := iv["label"]; ok {
						cats = append(cats, fmt.Sprintf("%v", lbl))
					}
				}
			}
		case map[string]interface{}:
			for k := range cv {
				cats = append(cats, k)
			}
		}
		if len(cats) == 0 && len(rf.ValueMap) > 0 {
			for k := range rf.ValueMap {
				cats = append(cats, k)
			}
		}

		out = append(out, domain.ResourceField{
			Name:        strings.TrimSpace(rf.Name),
			Type:        strings.TrimSpace(rf.Type),
			Title:       strings.TrimSpace(rf.Title),
			Description: strings.TrimSpace(rf.Description),
			RDFType:     strings.TrimSpace(rf.RDFType),
			Concept:     concept,
			Unit:        strings.TrimSpace(rf.Unit),
			Categories:  cats,
			Custom:      customMap,
		})
	}
	return out
}

func (p *DataPackageYAMLParser) ParseManifest(ctx context.Context, data []byte, baseURL string) (*domain.PackageManifest, error) {
	var raw rawYAMLDataPackage
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	manifest := &domain.PackageManifest{
		Name:        strings.TrimSpace(raw.Name),
		Title:       strings.TrimSpace(raw.Title),
		Description: strings.TrimSpace(raw.Description),
		Version:     strings.TrimSpace(raw.Version),
		Format:      domain.PackageFormatDataPackageYAML,
		Resources:   make([]domain.Resource, 0, len(raw.Resources)),
	}

	for _, r := range raw.Resources {
		if strings.TrimSpace(r.Name) == "" {
			continue
		}
		rawPath := strings.TrimSpace(r.Path)
		if rawPath == "" {
			rawPath = strings.TrimSpace(r.URL)
		}
		fmtStr := strings.ToLower(strings.TrimSpace(r.Format))
		if fmtStr == "" {
			fmtStr = inferFormatFromPath(rawPath)
		}
		resolvedPath := resolveResourcePath(baseURL, rawPath)

		bytesCount := r.BytesCount
		if bytesCount == 0 {
			bytesCount = r.Bytes
		}
		sha := strings.TrimSpace(r.HashSHA256)
		if sha == "" {
			sha = strings.TrimSpace(r.Hash)
		}

		manifest.Resources = append(manifest.Resources, domain.Resource{
			Name:        strings.TrimSpace(r.Name),
			Path:        resolvedPath,
			Format:      fmtStr,
			Description: strings.TrimSpace(r.Description),
			HashSHA256:  sha,
			BytesCount:  bytesCount,
			Fields:      convertRawFields(r.Schema.Fields),
		})

		subList := append(r.InternalResources, r.Subresources...)
		for _, sub := range subList {
			subName := strings.TrimSpace(sub.Name)
			if subName == "" {
				subName = strings.TrimSpace(sub.ID)
			}
			if subName == "" {
				continue
			}
			frag := strings.TrimPrefix(strings.TrimSpace(sub.Subpath), "#")
			if frag == "" {
				frag = strings.TrimPrefix(strings.TrimSpace(sub.Fragment), "#")
			}
			if frag == "" && strings.Contains(sub.URIFragment, "#") {
				parts := strings.SplitN(sub.URIFragment, "#", 2)
				frag = strings.TrimSpace(parts[1])
			}
			subPath := resolvedPath
			if frag != "" {
				subPath = resolvedPath + "#" + frag
			}
			subFmt := strings.ToLower(strings.TrimSpace(sub.Format))
			if subFmt == "" {
				subFmt = inferFormatFromPath(frag)
			}
			manifest.Resources = append(manifest.Resources, domain.Resource{
				Name:   subName,
				Path:   subPath,
				Format: subFmt,
			})
		}
	}

	return manifest, nil
}
