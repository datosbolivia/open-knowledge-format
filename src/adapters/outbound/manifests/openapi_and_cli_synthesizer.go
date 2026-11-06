package manifests

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
	"gopkg.in/yaml.v3"
)

var ojsCodeBlockRegex = regexp.MustCompile("(?s)```ojs\\b.*?```\\s*")

// StripOJSBlocks removes Observable JS (```ojs ... ```) visualization blocks from Markdown descriptions
// so LLM agents receive clean prose and synthesized semantic tables without JS noise.
func StripOJSBlocks(markdown string) string {
	cleaned := ojsCodeBlockRegex.ReplaceAllString(markdown, "")
	return strings.TrimSpace(cleaned)
}

// inferDatasetSlug extracts a clean dataset slug from a node URI or path (e.g. ".../nodes/bcb-semanal/index.md" -> "bcb-semanal").
func inferDatasetSlug(uriOrPath string, fallbackTitle string) string {
	clean := strings.TrimSpace(uriOrPath)
	clean = strings.TrimSuffix(clean, "/")
	if strings.HasSuffix(clean, "/index.md") || strings.HasSuffix(clean, "\\index.md") {
		clean = clean[:len(clean)-9]
	} else if strings.HasSuffix(clean, ".md") || strings.HasSuffix(clean, ".yaml") || strings.HasSuffix(clean, ".yml") || strings.HasSuffix(clean, ".json") {
		if idx := strings.LastIndexAny(clean, "/\\"); idx != -1 {
			base := clean[idx+1:]
			if strings.HasPrefix(base, "datapackage.") || base == "index.md" {
				clean = clean[:idx]
			}
		}
	}
	if idx := strings.LastIndexAny(clean, "/\\:"); idx != -1 {
		slug := strings.TrimSpace(clean[idx+1:])
		slug = strings.TrimSuffix(slug, ".md")
		if slug != "" && slug != "nodes" && slug != "bundles" {
			return slug
		}
	}
	if clean != "" && !strings.ContainsAny(clean, "/\\") {
		return strings.TrimSuffix(clean, ".md")
	}
	return domain.Slugify(fallbackTitle)
}

// SynthesizeContractMarkdown builds a unified Markdown section ("## Contrato de Datos y Esquema Semántico (Sintetizado)")
// combining tabular DataPackage schemas, OpenAPI 3.x service specifications, and OWL-CLI tool definitions.
func SynthesizeContractMarkdown(dp *domain.DataProduct, rawOpenAPIs map[string][]byte) string {
	if dp == nil {
		return ""
	}

	var sections []string
	dsSlug := inferDatasetSlug(dp.ID, dp.Manifest.Title)

	// Map semantic_fields from frontmatter by field name for fallback enrichment
	semFieldMap := make(map[string]domain.SemanticFieldMapping)
	for _, sf := range dp.Manifest.SemanticFields {
		semFieldMap[strings.ToLower(strings.TrimSpace(sf.FieldName))] = sf
	}

	// 1. Tabular DataPackage Resources & Fields
	if len(dp.Resources) > 0 {
		var sb strings.Builder
		for i, res := range dp.Resources {
			if i > 0 {
				sb.WriteString("\n\n")
			}
			triad := fmt.Sprintf("%s:%s", dsSlug, res.Name)
			fmtLabel := res.Format
			if fmtLabel == "" {
				fmtLabel = filepath.Ext(res.Path)
				fmtLabel = strings.TrimPrefix(fmtLabel, ".")
			}
			sb.WriteString(fmt.Sprintf("### Recurso: `%s` (`%s`)\n", triad, fmtLabel))
			sb.WriteString(fmt.Sprintf("- **Tríada SQL**: `\"%s\"`\n", triad))
			if res.Path != "" {
				sb.WriteString(fmt.Sprintf("- **Ruta Física**: `%s`\n", res.Path))
			}
			sb.WriteString(fmt.Sprintf("- **Consulta Rápida**: `SELECT * FROM \"%s\" LIMIT 10`\n", triad))

			if len(res.Fields) > 0 {
				sb.WriteString("\n| Columna | Tipo | Etiqueta / Descripción | Concepto Semántico (SKOS / OWL) | Valores / Categorías |\n")
				sb.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")
				for _, f := range res.Fields {
					colType := f.Type
					if colType == "" {
						colType = "string"
					}
					labelDesc := strings.TrimSpace(f.Title)
					if f.Description != "" && f.Description != f.Title {
						if labelDesc != "" {
							labelDesc = labelDesc + " — " + strings.TrimSpace(f.Description)
						} else {
							labelDesc = strings.TrimSpace(f.Description)
						}
					}
					if labelDesc == "" {
						labelDesc = "—"
					}

					concept := strings.TrimSpace(f.Concept)
					if concept == "" {
						concept = strings.TrimSpace(f.RDFType)
					}
					if sf, ok := semFieldMap[strings.ToLower(strings.TrimSpace(f.Name))]; ok {
						if concept == "" && sf.ConceptRef != "" {
							concept = sf.ConceptRef
						}
					}
					if concept == "" {
						concept = "—"
					}

					cats := "—"
					if len(f.Categories) > 0 {
						maxCats := f.Categories
						suffix := ""
						if len(maxCats) > 8 {
							maxCats = maxCats[:8]
							suffix = ", ..."
						}
						quoted := make([]string, 0, len(maxCats))
						for _, c := range maxCats {
							quoted = append(quoted, fmt.Sprintf("`%s`", c))
						}
						cats = strings.Join(quoted, ", ") + suffix
					}
					sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s | %s |\n",
						escapeMarkdownPipe(f.Name),
						escapeMarkdownPipe(colType),
						escapeMarkdownPipe(labelDesc),
						escapeMarkdownPipe(concept),
						escapeMarkdownPipe(cats),
					))
				}
			}
		}
		sections = append(sections, strings.TrimSpace(sb.String()))
	}

	// 2. OpenAPI Service Contracts
	if len(rawOpenAPIs) > 0 {
		keys := make([]string, 0, len(rawOpenAPIs))
		for k := range rawOpenAPIs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if md := synthesizeOpenAPIMarkdown(k, rawOpenAPIs[k]); md != "" {
				sections = append(sections, md)
			}
		}
	}

	// 3. OWL-CLI / OKF-CLI Local Tool Contracts
	for _, c := range dp.Manifest.Contracts {
		cType := strings.ToLower(strings.TrimSpace(c.Type))
		if cType == "owl-cli" || cType == "okf-cli" || c.Binary != "" {
			if md := synthesizeCLIMarkdown(c); md != "" {
				sections = append(sections, md)
			}
		}
	}

	if len(sections) == 0 {
		return ""
	}

	return "## Contrato de Datos y Esquema Semántico (Sintetizado)\n\n" + strings.Join(sections, "\n\n")
}

func escapeMarkdownPipe(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.ReplaceAll(s, "|", "\\|")
}

type openAPIDoc struct {
	OpenAPI string `yaml:"openapi" json:"openapi"`
	Info    struct {
		Title       string `yaml:"title" json:"title"`
		Version     string `yaml:"version" json:"version"`
		Description string `yaml:"description" json:"description"`
	} `yaml:"info" json:"info"`
	Servers []struct {
		URL         string `yaml:"url" json:"url"`
		Description string `yaml:"description" json:"description"`
	} `yaml:"servers" json:"servers"`
	Paths map[string]map[string]openAPIOperation `yaml:"paths" json:"paths"`
}

type openAPIOperation struct {
	Summary     string                   `yaml:"summary" json:"summary"`
	Description string                   `yaml:"description" json:"description"`
	OperationID string                   `yaml:"operationId" json:"operationId"`
	Parameters  []openAPIParameter       `yaml:"parameters" json:"parameters"`
	Responses   map[string]openAPIResp   `yaml:"responses" json:"responses"`
}

type openAPIParameter struct {
	Name        string                 `yaml:"name" json:"name"`
	In          string                 `yaml:"in" json:"in"`
	Required    bool                   `yaml:"required" json:"required"`
	Description string                 `yaml:"description" json:"description"`
	Concept     string                 `yaml:"x-okf-concept" json:"x-okf-concept"`
	Schema      map[string]interface{} `yaml:"schema" json:"schema"`
}

type openAPIResp struct {
	Description string                            `yaml:"description" json:"description"`
	Content     map[string]openAPIMediaTypeObject `yaml:"content" json:"content"`
}

type openAPIMediaTypeObject struct {
	Schema openAPISchemaObject `yaml:"schema" json:"schema"`
}

type openAPISchemaObject struct {
	Type       string                              `yaml:"type" json:"type"`
	Concept    string                              `yaml:"x-okf-concept" json:"x-okf-concept"`
	Properties map[string]openAPISchemaPropertyObj `yaml:"properties" json:"properties"`
}

type openAPISchemaPropertyObj struct {
	Type        string `yaml:"type" json:"type"`
	Description string `yaml:"description" json:"description"`
	Concept     string `yaml:"x-okf-concept" json:"x-okf-concept"`
}

func synthesizeOpenAPIMarkdown(sourcePath string, raw []byte) string {
	var doc openAPIDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return ""
	}
	if len(doc.Paths) == 0 && doc.Info.Title == "" {
		return ""
	}

	var sb strings.Builder
	title := strings.TrimSpace(doc.Info.Title)
	if title == "" {
		title = filepath.Base(sourcePath)
	}
	sb.WriteString(fmt.Sprintf("### Servicio Remoto (OpenAPI): %s\n", title))
	if len(doc.Servers) > 0 && doc.Servers[0].URL != "" {
		sb.WriteString(fmt.Sprintf("- **Servidor Base**: `%s`\n", doc.Servers[0].URL))
	}

	pathKeys := make([]string, 0, len(doc.Paths))
	for p := range doc.Paths {
		pathKeys = append(pathKeys, p)
	}
	sort.Strings(pathKeys)

	for _, pathStr := range pathKeys {
		methods := doc.Paths[pathStr]
		methodKeys := make([]string, 0, len(methods))
		for m := range methods {
			methodKeys = append(methodKeys, m)
		}
		sort.Strings(methodKeys)

		for _, m := range methodKeys {
			op := methods[m]
			upperMethod := strings.ToUpper(m)
			summary := strings.TrimSpace(op.Summary)
			if summary == "" {
				summary = strings.TrimSpace(op.Description)
			}
			if summary != "" {
				sb.WriteString(fmt.Sprintf("\n#### Endpoint: `%s %s` — %s\n", upperMethod, pathStr, summary))
			} else {
				sb.WriteString(fmt.Sprintf("\n#### Endpoint: `%s %s`\n", upperMethod, pathStr))
			}

			if len(op.Parameters) > 0 {
				sb.WriteString("\n| Parámetro (Input) | Ubicación | Tipo | Requerido | Concepto Semántico (`x-okf-concept`) | Descripción |\n")
				sb.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- |\n")
				for _, param := range op.Parameters {
					pType := "string"
					if t, ok := param.Schema["type"].(string); ok && t != "" {
						pType = t
					}
					reqStr := "No"
					if param.Required {
						reqStr = "Sí"
					}
					concept := strings.TrimSpace(param.Concept)
					if concept == "" {
						if c, ok := param.Schema["x-okf-concept"].(string); ok && c != "" {
							concept = c
						}
					}
					if concept == "" {
						concept = "—"
					}
					desc := strings.TrimSpace(param.Description)
					if desc == "" {
						desc = "—"
					}
					sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %s | %s | %s |\n",
						escapeMarkdownPipe(param.Name),
						escapeMarkdownPipe(param.In),
						escapeMarkdownPipe(pType),
						reqStr,
						escapeMarkdownPipe(concept),
						escapeMarkdownPipe(desc),
					))
				}
			}

			// Extract response properties for 200 / 201 / default
			for status, resp := range op.Responses {
				for _, media := range resp.Content {
					if len(media.Schema.Properties) > 0 {
						sb.WriteString(fmt.Sprintf("\n** Campos de Respuesta (`HTTP %s`):**\n\n", status))
						sb.WriteString("| Propiedad (Output) | Tipo | Concepto Semántico (`x-okf-concept`) | Descripción |\n")
						sb.WriteString("| :--- | :--- | :--- | :--- |\n")
						propNames := make([]string, 0, len(media.Schema.Properties))
						for pn := range media.Schema.Properties {
							propNames = append(propNames, pn)
						}
						sort.Strings(propNames)
						for _, pn := range propNames {
							prop := media.Schema.Properties[pn]
							pType := prop.Type
							if pType == "" {
								pType = "any"
							}
							concept := strings.TrimSpace(prop.Concept)
							if concept == "" {
								concept = "—"
							}
							desc := strings.TrimSpace(prop.Description)
							if desc == "" {
								desc = "—"
							}
							sb.WriteString(fmt.Sprintf("| `%s` | `%s` | %s | %s |\n",
								escapeMarkdownPipe(pn),
								escapeMarkdownPipe(pType),
								escapeMarkdownPipe(concept),
								escapeMarkdownPipe(desc),
							))
						}
					}
				}
			}
		}
	}

	return strings.TrimSpace(sb.String())
}

func synthesizeCLIMarkdown(c domain.Contract) string {
	var sb strings.Builder
	cmdName := strings.TrimSpace(c.Binary)
	if c.Subcommand != "" {
		cmdName = strings.TrimSpace(cmdName + " " + c.Subcommand)
	}
	if cmdName == "" {
		cmdName = "cli-tool"
	}

	cType := c.Type
	if cType == "" {
		cType = "owl-cli"
	}
	sb.WriteString(fmt.Sprintf("### Herramienta CLI Local (`%s`): `%s`\n", cType, cmdName))

	// Build invocation example
	var cmdParts []string
	cmdParts = append(cmdParts, cmdName)
	for _, inp := range c.Inputs {
		if inp.Flag != "" {
			cmdParts = append(cmdParts, fmt.Sprintf("%s <%s>", inp.Flag, inp.Name))
		} else {
			cmdParts = append(cmdParts, fmt.Sprintf("<%s>", inp.Name))
		}
	}
	sb.WriteString(fmt.Sprintf("- **Plantilla de Ejecución**: `%s`\n", strings.Join(cmdParts, " ")))

	if len(c.Inputs) > 0 {
		sb.WriteString("\n| Parámetro (Input) | Flag / Posición | Tipo | Requerido | Concepto Semántico (SKOS / OWL) | Descripción |\n")
		sb.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- |\n")
		for _, inp := range c.Inputs {
			flagPos := inp.Flag
			if flagPos == "" && inp.Position > 0 {
				flagPos = fmt.Sprintf("pos #%d", inp.Position)
			}
			if flagPos == "" {
				flagPos = "—"
			}
			pType := inp.Type
			if pType == "" {
				pType = "string"
			}
			reqStr := "No"
			if inp.Required {
				reqStr = "Sí"
			}
			concept := strings.TrimSpace(inp.Concept)
			if concept == "" {
				concept = "—"
			}
			desc := strings.TrimSpace(inp.Description)
			if desc == "" {
				desc = "—"
			}
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %s | %s | %s |\n",
				escapeMarkdownPipe(inp.Name),
				escapeMarkdownPipe(flagPos),
				escapeMarkdownPipe(pType),
				reqStr,
				escapeMarkdownPipe(concept),
				escapeMarkdownPipe(desc),
			))
		}
	}

	if c.Outputs != nil && (c.Outputs.Format != "" || c.Outputs.Concept != "" || c.Outputs.Description != "") {
		outFmt := c.Outputs.Format
		if outFmt == "" {
			outFmt = "stdout"
		}
		outConcept := c.Outputs.Concept
		if outConcept == "" {
			outConcept = "—"
		}
		sb.WriteString(fmt.Sprintf("\n- **Salida Semántica (Output)**: Formato `%s` | Concepto: `%s`", outFmt, outConcept))
		if c.Outputs.Description != "" {
			sb.WriteString(fmt.Sprintf(" — %s", c.Outputs.Description))
		}
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String())
}
