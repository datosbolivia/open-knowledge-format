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
		if issue.Code == "RULE-12-CIRCULAR-REFERENCE" {
			foundCycle = true
			break
		}
	}
	if !foundCycle {
		t.Errorf("expected RULE-12-CIRCULAR-REFERENCE error for circular dependency, none found")
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
