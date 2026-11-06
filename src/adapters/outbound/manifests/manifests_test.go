package manifests

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
)

func TestDataPackageJSONParser(t *testing.T) {
	jsonContent := `{
		"name": "estadisticas-agetic",
		"title": "Estadísticas AGETIC",
		"resources": [
			{
				"name": "Estadisticas Ciudadania Digital",
				"path": "data/ciudadania.csv",
				"format": "csv",
				"bytes": 1024
			},
			{
				"name": "interoperabilidad",
				"path": "https://remote.com/data.parquet"
			}
		]
	}`

	parser := NewDataPackageJSONParser()
	if !parser.CanParse("datapackage.json", []byte(jsonContent)) {
		t.Fatal("expected parser to handle datapackage.json")
	}

	manifest, err := parser.ParseManifest(context.Background(), []byte(jsonContent), "https://example.com/nodes/agetic/datapackage.json")
	if err != nil {
		t.Fatalf("failed to parse json manifest: %v", err)
	}

	if manifest.Name != "estadisticas-agetic" {
		t.Errorf("name mismatch: %s", manifest.Name)
	}
	if len(manifest.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(manifest.Resources))
	}

	res0 := manifest.Resources[0]
	if res0.Name != "Estadisticas Ciudadania Digital" || res0.Format != "csv" || res0.BytesCount != 1024 {
		t.Errorf("res0 mismatch: %+v", res0)
	}
	if res0.Path != "https://example.com/nodes/agetic/data/ciudadania.csv" {
		t.Errorf("res0 path resolution mismatch: %s", res0.Path)
	}

	res1 := manifest.Resources[1]
	if res1.Format != "parquet" {
		t.Errorf("expected format inferred as parquet, got: %s", res1.Format)
	}
	if res1.Path != "https://remote.com/data.parquet" {
		t.Errorf("res1 path mismatch: %s", res1.Path)
	}
}

func TestDataPackageYAMLParser(t *testing.T) {
	yamlContent := `name: p2p-bob-exchange
title: Mercado Cambiario P2P BOB / USDT
resources:
- name: advertiser
  path: https://www.kaggle.com/datasets/andreschirinos/p2p-bob-exchange/advertiser.parquet
  format: parquet
  bytes_count: 2048
- name: Compilación de datos de calidad del aire de Bolivia
  path: data/calidad_aire.csv
`

	parser := NewDataPackageYAMLParser()
	if !parser.CanParse("datapackage.yaml", []byte(yamlContent)) {
		t.Fatal("expected parser to handle datapackage.yaml")
	}
	if !parser.CanParse("datapackage.yml", []byte(yamlContent)) {
		t.Fatal("expected parser to handle datapackage.yml")
	}

	manifest, err := parser.ParseManifest(context.Background(), []byte(yamlContent), "/local/nodes/p2p/datapackage.yaml")
	if err != nil {
		t.Fatalf("failed to parse yaml manifest: %v", err)
	}

	if manifest.Name != "p2p-bob-exchange" {
		t.Errorf("name mismatch: %s", manifest.Name)
	}
	if len(manifest.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(manifest.Resources))
	}

	r0, found := manifest.FindResource("advertiser")
	if !found || r0.Format != "parquet" || r0.BytesCount != 2048 {
		t.Errorf("r0 find mismatch: %+v", r0)
	}

	r1, found := manifest.FindResource("compilaci_n_de_datos_de_calidad_del_aire_de_bolivia")
	if !found || r1.Format != "csv" {
		t.Errorf("r1 find by slug mismatch: %+v", r1)
	}
}

func TestOKFMarkdownParser(t *testing.T) {
	mdContent := `---
type: dataset
title: Censo Nacional de Poblacion y Vivienda
dimensions:
- departamento
- provincia
contracts:
- type: datapackage
  path: ./datapackage.json
lineage:
  version: 2.1.0
---

# Documentacion Censo

Detalles sobre el censo de Bolivia.
`

	parser := NewOKFMarkdownParser()
	if !parser.CanParse("index.md", []byte(mdContent)) {
		t.Fatal("expected parser to handle index.md")
	}

	manifest, err := parser.ParseManifest(context.Background(), []byte(mdContent), "")
	if err != nil {
		t.Fatalf("failed to parse markdown manifest: %v", err)
	}

	if manifest.Title != "Censo Nacional de Poblacion y Vivienda" {
		t.Errorf("title mismatch: %s", manifest.Title)
	}
	if manifest.Version != "2.1.0" {
		t.Errorf("version mismatch: %s", manifest.Version)
	}
}

func TestUnifiedMetadataReader(t *testing.T) {
	tmpDir := t.TempDir()
	nodeDir := filepath.Join(tmpDir, "air_quality")
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		t.Fatal(err)
	}

	indexContent := `---
type: dataset
title: Calidad del Aire
dimensions:
- fecha
- lugar
---
# Calidad del Aire en Bolivia
`
	if err := os.WriteFile(filepath.Join(nodeDir, "index.md"), []byte(indexContent), 0644); err != nil {
		t.Fatal(err)
	}

	pkgYaml := `name: air_quality
resources:
- name: mediciones
  path: https://remote.io/mediciones.csv
  format: csv
`
	if err := os.WriteFile(filepath.Join(nodeDir, "datapackage.yaml"), []byte(pkgYaml), 0644); err != nil {
		t.Fatal(err)
	}

	reader := NewUnifiedMetadataReader("", 5*time.Second)

	// Test ReadNode
	dp, err := reader.ReadNode(context.Background(), nodeDir)
	if err != nil {
		t.Fatalf("failed to read node: %v", err)
	}

	if dp.Manifest.Title != "Calidad del Aire" {
		t.Errorf("title mismatch: %s", dp.Manifest.Title)
	}
	if len(dp.Resources) != 1 || dp.Resources[0].Name != "mediciones" {
		t.Errorf("resources mismatch: %+v", dp.Resources)
	}

	// Test ResolveTriad
	triad := domain.ResourceTriad{
		Dataset:  "air_quality",
		Resource: "mediciones",
	}

	resolved, err := reader.ResolveTriad(context.Background(), triad)
	if err != nil {
		t.Fatalf("failed to resolve triad: %v", err)
	}
	if resolved.Triad.Dataset != "air_quality" || resolved.Triad.Resource != "mediciones" {
		t.Errorf("resolved triad mismatch: %+v", resolved.Triad)
	}
}

func TestOKFMarkdownParserWithByReferenceContract(t *testing.T) {
	mdContent := `---
type: dataset
title: Atlas Electoral y Elecciones
nature: structured
status: verified
supersedes: odkf://org.gob.bolivia/oep/atlas-preliminar:v1.0
dimensions:
- año
- departamento
contracts:
- type: datapackage
  ref: ./datapackage.yaml
lineage:
  version: 2.0.0
  ref: ./lineage.yaml
---
# Atlas Electoral

Metodología científica validada.
`
	manifest, body, err := ParseOKFFrontmatter([]byte(mdContent))
	if err != nil {
		t.Fatalf("unexpected error parsing frontmatter: %v", err)
	}

	if manifest.Title != "Atlas Electoral y Elecciones" {
		t.Errorf("title mismatch: %s", manifest.Title)
	}
	if manifest.Nature != "structured" {
		t.Errorf("nature mismatch: %s", manifest.Nature)
	}
	if manifest.Status != "verified" {
		t.Errorf("status mismatch: %s", manifest.Status)
	}
	if manifest.Supersedes != "odkf://org.gob.bolivia/oep/atlas-preliminar:v1.0" {
		t.Errorf("supersedes mismatch: %s", manifest.Supersedes)
	}
	if len(manifest.Contracts) != 1 || manifest.Contracts[0].Ref != "./datapackage.yaml" {
		t.Errorf("contracts ref mismatch: %+v", manifest.Contracts)
	}
	if manifest.Lineage.Ref != "./lineage.yaml" || manifest.Lineage.Version != "2.0.0" {
		t.Errorf("lineage ref mismatch: %+v", manifest.Lineage)
	}
	if !strings.Contains(body, "Metodología científica validada") {
		t.Errorf("body content mismatch: %s", body)
	}
}

func TestDataPackageYAMLParserWithSchemaFields(t *testing.T) {
	yamlContent := `name: bcb-semanal
title: Estadísticas Semanales del BCB
resources:
  - name: bcb_reservas
    path: https://datosbolivia.github.io/raw/nodes/bcb-semanal/data/bcb_reservas.parquet
    format: parquet
    schema:
      fields:
        - name: fecha
          type: date
          title: Fecha de corte
          description: Fecha semanal del reporte
          concept: ./concepts/fecha_corte.md
        - name: tipo
          type: string
          title: Tipo de activo de reserva
          categories:
            - Divisas
            - Oro
            - DEG
`
	parser := NewDataPackageYAMLParser()
	pkg, err := parser.ParseManifest(context.Background(), []byte(yamlContent), "")
	if err != nil {
		t.Fatalf("unexpected error parsing YAML with schema.fields: %v", err)
	}
	if len(pkg.Resources) != 1 {
		t.Fatalf("expected exactly 1 resource (no fake column resources), got %d: %+v", len(pkg.Resources), pkg.Resources)
	}
	res := pkg.Resources[0]
	if res.Name != "bcb_reservas" || res.Format != "parquet" {
		t.Errorf("resource metadata mismatch: %+v", res)
	}
	if len(res.Fields) != 2 {
		t.Fatalf("expected 2 schema fields, got %d", len(res.Fields))
	}
	if res.Fields[0].Name != "fecha" || res.Fields[0].Concept != "./concepts/fecha_corte.md" {
		t.Errorf("field 0 mismatch: %+v", res.Fields[0])
	}
	if res.Fields[1].Name != "tipo" || len(res.Fields[1].Categories) != 3 {
		t.Errorf("field 1 categories mismatch: %+v", res.Fields[1])
	}
}

func TestLLMSTxtParserWithMultiLinkLineAndAccentSearch(t *testing.T) {
	llmsContent := `# Catálogo Datos Bolivia
> Índice soberano de datos abiertos

- [Créditos Bancarios ASFI](/raw/nodes/asfi-creditos/index.md): Cartera de créditos por municipio. (Dominio: Finanzas. Recursos: cartera_municipal)
- [Índice de Calidad del Aire (ICA)](./concepts/air_quality_index.md) - [Q2364111](https://www.wikidata.org/wiki/Q2364111)
`
	parser := NewLLMSTxtParser()
	cat, err := parser.ParseCatalog([]byte(llmsContent), "https://datosbolivia.github.io/raw/nodes/air_quality/index.md")
	if err != nil {
		t.Fatalf("failed to parse llms.txt: %v", err)
	}
	if len(cat.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(cat.Entries))
	}
	icaEntry := cat.Entries[1]
	if icaEntry.URI != "./concepts/air_quality_index.md" {
		t.Errorf("multi-link regex corrupted URI: got %q, want %q", icaEntry.URI, "./concepts/air_quality_index.md")
	}
	if icaEntry.ResolvedURL != "https://datosbolivia.github.io/raw/nodes/air_quality/concepts/air_quality_index.md" {
		t.Errorf("resolved URL mismatch: got %q", icaEntry.ResolvedURL)
	}

	// Accent-insensitive & multi-token search
	matches := cat.FindByKeyword("creditos")
	if len(matches) != 1 || matches[0].Title != "Créditos Bancarios ASFI" {
		t.Errorf("accent-insensitive search for 'creditos' failed: %+v", matches)
	}
	multiMatches := cat.FindByKeyword("creditos municipio")
	if len(multiMatches) != 1 {
		t.Errorf("multi-token search 'creditos municipio' failed: %+v", multiMatches)
	}
}

func TestUnifiedMetadataReaderSynthesizesMarkdownContracts(t *testing.T) {
	tmpDir := t.TempDir()
	nodeDir := filepath.Join(tmpDir, "geo-municipio")
	conceptsDir := filepath.Join(nodeDir, "concepts")
	if err := os.MkdirAll(conceptsDir, 0755); err != nil {
		t.Fatal(err)
	}

	indexMD := `---
type: dataset
title: Geo Municipio Hub
dimensions:
  - municipio
contracts:
  - type: datapackage
    path: ./datapackage.yaml
  - type: openapi
    path: ./openapi.yaml
  - type: owl-cli
    binary: okf-geo
    subcommand: enrich-municipio
    inputs:
      - name: codigo_ine
        flag: --cod-ine
        type: string
        concept: wd:Q15284
        required: true
        description: Código INE de 6 dígitos
    outputs:
      format: json
      concept: wd:Q186588
      description: Polígono GeoJSON del municipio
---
# Geo Municipio Hub

Descripción general del nodo.

` + "```ojs\nviewof x = Inputs.select(['a', 'b'])\n```" + `
`
	if err := os.WriteFile(filepath.Join(nodeDir, "index.md"), []byte(indexMD), 0644); err != nil {
		t.Fatal(err)
	}

	dpYAML := `name: geo-municipio
resources:
  - name: municipios_2024
    path: data/municipios.parquet
    format: parquet
    schema:
      fields:
        - name: cod_ine
          type: string
          title: Código INE Municipal
          concept: wd:Q15284
`
	if err := os.WriteFile(filepath.Join(nodeDir, "datapackage.yaml"), []byte(dpYAML), 0644); err != nil {
		t.Fatal(err)
	}

	openAPIYAML := `openapi: 3.1.0
info:
  title: API de Riesgo Municipal
  version: 1.0.0
servers:
  - url: https://api.datosbolivia.org
paths:
  /v1/riesgo/municipio:
    get:
      summary: Consulta riesgo climático por código INE
      parameters:
        - name: cod_ine
          in: query
          required: true
          x-okf-concept: wd:Q15284
          schema:
            type: string
          description: Código INE de 6 dígitos
      responses:
        "200":
          description: Respuesta exitosa
          content:
            application/json:
              schema:
                type: object
                properties:
                  indice_riesgo:
                    type: number
                    x-okf-concept: wd:Q7944
                    description: Índice de vulnerabilidad sísmica/climática
`
	if err := os.WriteFile(filepath.Join(nodeDir, "openapi.yaml"), []byte(openAPIYAML), 0644); err != nil {
		t.Fatal(err)
	}

	conceptMD := `---
type: concept
title: Código INE Municipal
skos:
  - https://www.wikidata.org/wiki/Q15284
---
# Código INE Municipal

Identificador oficial de 6 dígitos del municipio.
`
	conceptFile := filepath.Join(conceptsDir, "cod_ine.md")
	if err := os.WriteFile(conceptFile, []byte(conceptMD), 0644); err != nil {
		t.Fatal(err)
	}

	reader := NewUnifiedMetadataReader("", 5*time.Second)
	dp, err := reader.ReadNode(context.Background(), nodeDir)
	if err != nil {
		t.Fatalf("ReadNode failed: %v", err)
	}

	if strings.Contains(dp.Description, "```ojs") {
		t.Errorf("expected OJS code block to be stripped from Description, got: %s", dp.Description)
	}
	if !strings.Contains(dp.Description, "## Contrato de Datos y Esquema Semántico (Sintetizado)") {
		t.Errorf("expected synthesized contract header in Description, got: %s", dp.Description)
	}
	if !strings.Contains(dp.Description, "`cod_ine`") || !strings.Contains(dp.Description, "`wd:Q15284`") && !strings.Contains(dp.Description, "wd:Q15284") {
		t.Errorf("expected synthesized column and concept in Description, got: %s", dp.Description)
	}
	if !strings.Contains(dp.Description, "GET /v1/riesgo/municipio") {
		t.Errorf("expected OpenAPI endpoint synthesized in Description, got: %s", dp.Description)
	}
	if !strings.Contains(dp.Description, "okf-geo enrich-municipio") {
		t.Errorf("expected OWL-CLI tool synthesized in Description, got: %s", dp.Description)
	}

	// Verify reading a specific concept .md file does NOT overwrite it with parent index.md
	conceptDP, err := reader.ReadNode(context.Background(), conceptFile)
	if err != nil {
		t.Fatalf("ReadNode on specific concept file failed: %v", err)
	}
	if conceptDP.Manifest.Title != "Código INE Municipal" {
		t.Errorf("expected concept title 'Código INE Municipal', got %q", conceptDP.Manifest.Title)
	}
}

func TestExternalOWLCLIContractDoesNotEclipseDataPackageAndOpenAPIArrayResponse(t *testing.T) {
	tmpDir := t.TempDir()
	nodeDir := filepath.Join(tmpDir, "hub-node")
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Put owl-cli BEFORE datapackage in contracts list to verify it does not eclipse datapackage.yaml
	indexMD := `---
type: dataset
title: Hub Con CLI Externo
dimensions:
  - municipio
contracts:
  - type: owl-cli
    path: ./owl-cli.yaml
  - type: datapackage
    path: ./datapackage.yaml
  - type: openapi
    path: ./openapi.yaml
---
# Hub Con CLI Externo
`
	if err := os.WriteFile(filepath.Join(nodeDir, "index.md"), []byte(indexMD), 0644); err != nil {
		t.Fatal(err)
	}

	cliYAML := `type: owl-cli
binary: datamesh-geo
subcommand: calc-risk
arguments:
  - name: cod_ine
    flag: --cod-ine
    type: string
    concept: wd:Q15284
    required: true
output:
  format: json
  concept: wd:Q7944
  description: Puntaje de riesgo
`
	if err := os.WriteFile(filepath.Join(nodeDir, "owl-cli.yaml"), []byte(cliYAML), 0644); err != nil {
		t.Fatal(err)
	}

	dpYAML := `name: hub-node
resources:
  - name: tabla_base
    path: data/base.parquet
    format: parquet
`
	if err := os.WriteFile(filepath.Join(nodeDir, "datapackage.yaml"), []byte(dpYAML), 0644); err != nil {
		t.Fatal(err)
	}

	openAPIYAML := `openapi: 3.1.0
info:
  title: API Array Concept
  version: 1.0.0
paths:
  /v1/items:
    get:
      summary: Lista items
      responses:
        "200":
          description: Lista
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  x-okf-concept: wd:Q186588
`
	if err := os.WriteFile(filepath.Join(nodeDir, "openapi.yaml"), []byte(openAPIYAML), 0644); err != nil {
		t.Fatal(err)
	}

	reader := NewUnifiedMetadataReader("", 5*time.Second)
	dp, err := reader.ReadNode(context.Background(), nodeDir)
	if err != nil {
		t.Fatalf("ReadNode failed: %v", err)
	}
	if len(dp.Resources) != 1 || dp.Resources[0].Name != "tabla_base" {
		t.Fatalf("expected datapackage resource 'tabla_base' not to be eclipsed by owl-cli.yaml, got: %+v", dp.Resources)
	}
	if !strings.Contains(dp.Description, "datamesh-geo calc-risk") || !strings.Contains(dp.Description, "--cod-ine") {
		t.Errorf("expected external owl-cli.yaml to be hydrated in Description, got: %s", dp.Description)
	}
	if !strings.Contains(dp.Description, "wd:Q186588") {
		t.Errorf("expected OpenAPI array items x-okf-concept in Description, got: %s", dp.Description)
	}
}
