package domain

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateConceptContent_ActorFormat_Rule09(t *testing.T) {
	// Valid actor
	validContent := `---
type: dataset
title: Test Dataset
skos: ["http://www.w3.org/ns/dcat#Dataset"]
contracts:
  - type: datapackage
    path: https://example.com/datapackage.json
generated:
  by: "Andres Chirinos <andres@datosbolivia.org>"
  at: "2026-10-22T12:00:00Z"
---
# Content
`
	rep := ValidateConceptContent(validContent, "valid.md")
	for _, issue := range rep.Issues {
		if issue.Code == "RULE-09-ACTOR-FORMAT" {
			t.Errorf("expected no actor format issue, got: %v", issue.Message)
		}
	}

	// Invalid actor format
	invalidContent := `---
type: dataset
title: Test Dataset
skos: ["http://www.w3.org/ns/dcat#Dataset"]
contracts:
  - type: datapackage
    path: https://example.com/datapackage.json
generated:
  by: "InvalidActorFormatWithoutEmailOrURI"
---
# Content
`
	repInvalid := ValidateConceptContent(invalidContent, "invalid.md")
	found := false
	for _, issue := range repInvalid.Issues {
		if issue.Code == "RULE-09-ACTOR-FORMAT" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected RULE-09-ACTOR-FORMAT warning for invalid actor, none found")
	}
}

func TestValidateConceptContent_StatusEnum_Rule10(t *testing.T) {
	// Superseded without superseded_by must fail with ERROR
	supersededWithoutTarget := `---
type: dataset
title: Superseded Dataset
skos: ["http://www.w3.org/ns/dcat#Dataset"]
contracts:
  - type: datapackage
    path: https://example.com/datapackage.json
status: superseded
---
# Content
`
	rep := ValidateConceptContent(supersededWithoutTarget, "superseded.md")
	found := false
	for _, issue := range rep.Issues {
		if issue.Code == "RULE-10-SUPERSEDED-TARGET" && issue.Severity == SeverityError {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected RULE-10-SUPERSEDED-TARGET error for superseded without superseded_by")
	}

	// Superseded WITH superseded_by must pass
	supersededWithTarget := `---
type: dataset
title: Superseded Dataset
skos: ["http://www.w3.org/ns/dcat#Dataset"]
contracts:
  - type: datapackage
    path: https://example.com/datapackage.json
status: superseded
superseded_by: "../v2/concept.md"
---
# Content
`
	rep2 := ValidateConceptContent(supersededWithTarget, "superseded2.md")
	for _, issue := range rep2.Issues {
		if issue.Code == "RULE-10-SUPERSEDED-TARGET" {
			t.Errorf("expected no RULE-10-SUPERSEDED-TARGET error, got: %s", issue.Message)
		}
	}
}

func TestValidateConceptContent_FreshnessAndDates_Rule11(t *testing.T) {
	// Valid ISO 8601 duration and RFC3339 dates
	validFreshness := `---
type: dataset
title: Fresh Dataset
skos: ["http://www.w3.org/ns/dcat#Dataset"]
contracts:
  - type: datapackage
    path: https://example.com/datapackage.json
freshness:
  cadence: P1W
  expires: "2026-12-31T23:59:59Z"
---
# Content
`
	rep := ValidateConceptContent(validFreshness, "fresh.md")
	for _, issue := range rep.Issues {
		if issue.Code == "RULE-11-CADENCE-SYNTAX" || issue.Code == "RULE-11-EXPIRES-SYNTAX" {
			t.Errorf("unexpected date issue: %s", issue.Message)
		}
	}

	// Invalid cadence
	invalidFreshness := `---
type: dataset
title: Fresh Dataset
skos: ["http://www.w3.org/ns/dcat#Dataset"]
contracts:
  - type: datapackage
    path: https://example.com/datapackage.json
freshness:
  cadence: "every 2 weeks"
---
# Content
`
	repInvalid := ValidateConceptContent(invalidFreshness, "invalid_fresh.md")
	found := false
	for _, issue := range repInvalid.Issues {
		if issue.Code == "RULE-11-CADENCE-SYNTAX" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected RULE-11-CADENCE-SYNTAX warning for 'every 2 weeks'")
	}
}

func TestValidateBundle_CycleDetection_Rule12(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "okf_bundle_cycle_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Concept A references Concept B
	conceptA := `---
type: concept
title: Concept A
---
Links to [Concept B](b.md)
`
	// Concept B references Concept A (Circular!)
	conceptB := `---
type: concept
title: Concept B
---
Links to [Concept A](a.md)
`
	if err := os.WriteFile(filepath.Join(tmpDir, "a.md"), []byte(conceptA), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.md"), []byte(conceptB), 0644); err != nil {
		t.Fatal(err)
	}

	report, err := ValidateBundle(tmpDir)
	if err != nil {
		t.Fatalf("ValidateBundle failed: %v", err)
	}

	foundCycle := false
	for _, issue := range report.Issues {
		if issue.Code == "RULE-12-CIRCULAR-REFERENCE" && issue.Severity == SeverityWarning {
			foundCycle = true
			break
		}
	}
	if !foundCycle {
		t.Errorf("expected RULE-12-CIRCULAR-REFERENCE warning for navigational circular link, none found")
	}

	// Now test strict DAG cycle (supersedes cycle) -> must be SeverityError
	dagA := `---
type: concept
title: Concept A
supersedes: b.md
---
# Concept A
`
	dagB := `---
type: concept
title: Concept B
supersedes: a.md
---
# Concept B
`
	_ = os.WriteFile(filepath.Join(tmpDir, "a.md"), []byte(dagA), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "b.md"), []byte(dagB), 0644)

	dagReport, err := ValidateBundle(tmpDir)
	if err != nil {
		t.Fatalf("ValidateBundle DAG failed: %v", err)
	}
	foundDAGError := false
	for _, issue := range dagReport.Issues {
		if issue.Code == "RULE-12-DAG-CYCLE" && issue.Severity == SeverityError {
			foundDAGError = true
			break
		}
	}
	if !foundDAGError {
		t.Errorf("expected RULE-12-DAG-CYCLE error for circular supersedes DAG, none found")
	}
}

func TestParseTriad_SovereignDNS_OKF_Fragments(t *testing.T) {
	cases := []struct {
		raw      string
		wantCat  string
		wantDs   string
		wantRes  string
		wantFrag string
	}{
		{
			raw:      "elecciones_2025.votos_municipales@bolivia",
			wantCat:  "bolivia",
			wantDs:   "elecciones_2025",
			wantRes:  "votos_municipales",
			wantFrag: "",
		},
		{
			raw:      "censo_2024_cartografia.cartografia_completa_zip@ine_bolivia#scz_manzanas.parquet",
			wantCat:  "ine_bolivia",
			wantDs:   "censo_2024_cartografia",
			wantRes:  "cartografia_completa_zip",
			wantFrag: "scz_manzanas.parquet",
		},
		{
			raw:      "okf://udabol_investigacion/epidemiologia/dengue_2024#sub.csv",
			wantCat:  "udabol_investigacion",
			wantDs:   "epidemiologia",
			wantRes:  "dengue_2024",
			wantFrag: "sub.csv",
		},
		{
			raw:      "bcb-semanal.bcb_reservas#divisas.csv",
			wantCat:  "",
			wantDs:   "bcb-semanal",
			wantRes:  "bcb_reservas",
			wantFrag: "divisas.csv",
		},
	}

	for _, tc := range cases {
		tr, err := ParseTriad(tc.raw)
		if err != nil {
			t.Fatalf("ParseTriad(%q) unexpected error: %v", tc.raw, err)
		}
		if tr.Catalog != tc.wantCat || tr.Dataset != tc.wantDs || tr.Resource != tc.wantRes || tr.Fragment != tc.wantFrag {
			t.Errorf("ParseTriad(%q) = %+v, want cat=%s ds=%s res=%s frag=%s",
				tc.raw, tr, tc.wantCat, tc.wantDs, tc.wantRes, tc.wantFrag)
		}
	}

	// §14.4 Dot-notation disambiguation: filenames with recognized extensions must NOT parse as two-part triads
	for _, fn := range []string{"data.csv", "table.parquet", "index.md", "archive.zip", "config.yaml", "state.sqlite"} {
		if _, err := ParseTriad(fn); err == nil {
			t.Errorf("ParseTriad(%q) succeeded, expected rejection as physical filename per §14.4", fn)
		}
	}
}

func TestDataProductValidate_OpenTypeTaxonomy(t *testing.T) {
	openTypes := []string{"dcat:Dataset", "dcat:DataService", "Attested Computation", "BigQuery Table", "CustomDomainType"}
	for _, typ := range openTypes {
		_, err := NewDataProduct("id-1", Manifest{Title: "Sample", Type: typ}, "desc", nil)
		if err != nil {
			t.Errorf("NewDataProduct rejected valid open type %q: %v", typ, err)
		}
	}
}

func TestValidateBundle_AcyclicDAG_PassesRule12(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "okf_bundle_dag_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Concept A -> Concept B -> Concept C (DAG, No cycles)
	conceptA := `---
type: concept
title: Concept A
---
Links to [Concept B](b.md)
`
	conceptB := `---
type: concept
title: Concept B
---
Links to [Concept C](c.md)
`
	conceptC := `---
type: concept
title: Concept C
---
Terminal concept without outgoing links.
`
	_ = os.WriteFile(filepath.Join(tmpDir, "a.md"), []byte(conceptA), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "b.md"), []byte(conceptB), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "c.md"), []byte(conceptC), 0644)

	report, err := ValidateBundle(tmpDir)
	if err != nil {
		t.Fatalf("ValidateBundle failed: %v", err)
	}

	for _, issue := range report.Issues {
		if issue.Code == "RULE-12-CIRCULAR-REFERENCE" {
			t.Errorf("unexpected circular dependency error in DAG: %s", issue.Message)
		}
	}
}

func TestValidateConceptContent_MetaSpecificationPoliciesAndContracts(t *testing.T) {
	// Meta-specification dataset & service linking multiple external contracts (SDMX, AsyncAPI, OpenAPI)
	// and multiple external policy specifications (W3C ODRL, OPA/Rego, W3C DPV)
	metaSpecContent := `---
type: dcat:DataService
title: Federated Sovereign Service
skos:
  - "https://www.w3.org/TR/vocab-dcat-3/#Class:Data_Service"
contracts:
  - type: openapi_3
    ref: "https://example.org/specs/api.yaml"
  - type: sdmx
    ref: "https://example.org/specs/dsd.xml"
  - type: custom_industry_spec
    inline:
      version: "1.0"
policy:
  type: odrl
  ref: "https://www.w3.org/ns/odrl/2/policy.json"
policies:
  - type: opa
    ref: "https://example.org/policies/authz.rego"
  - type: dpv
    profile: zero_microdata
    rules:
      - id: k_anon_5
        aggregate_only: true
        min_group_size: 5
---
# Federated Service
`
	rep := ValidateConceptContent(metaSpecContent, "service.md")
	if !rep.Valid {
		t.Fatalf("expected meta-specification concept to be valid (0 errors), got issues: %+v", rep.Issues)
	}
}
