package usecases

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
)

type mockPublisher struct {
	targetName string
}

func (m *mockPublisher) Name() string                  { return m.targetName }
func (m *mockPublisher) CanPublish(target string) bool { return target == m.targetName }
func (m *mockPublisher) Publish(ctx context.Context, manifest *domain.PackageManifest, datasetPath string, options map[string]interface{}) (domain.PublicationTargetResult, error) {
	return domain.PublicationTargetResult{
		Target:         m.targetName,
		Success:        true,
		DestinationURI: "file:///mock/destination",
		Message:        "published successfully",
	}, nil
}

func TestPublishDatasetUseCase(t *testing.T) {
	pub := &mockPublisher{targetName: "local"}
	uc := NewPublishDatasetUseCase(nil, pub)

	results, err := uc.Publish(context.Background(), "my-dataset", []string{"local", "kaggle"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if !results[0].Success || results[0].Target != "local" {
		t.Errorf("expected local target success, got %+v", results[0])
	}

	if results[1].Success || results[1].Target != "kaggle" {
		t.Errorf("expected kaggle target failure (no adapter), got %+v", results[1])
	}
}

func TestSemanticAlignmentUseCase(t *testing.T) {
	uc := NewSemanticAlignmentUseCase()
	mapping := domain.SemanticFieldMapping{
		FieldName:  "cod_dep",
		ConceptRef: "concepts/departamentos.md",
		ValueMapping: map[string]string{
			"LPZ": "concept:departamentos:la_paz",
			"SCZ": "concept:departamentos:santa_cruz",
		},
	}

	want := "(CASE WHEN t.cod_dep = 'LPZ' THEN 'concept:departamentos:la_paz' WHEN t.cod_dep = 'SCZ' THEN 'concept:departamentos:santa_cruz' ELSE t.cod_dep END)"

	// Deterministic output: repeated calls must be identical (raw values sorted).
	for i := 0; i < 20; i++ {
		if expr := uc.GenerateCaseExpression(mapping, "t"); expr != want {
			t.Fatalf("non-deterministic CASE expression on call %d: %s", i, expr)
		}
	}

	// Test generic ValueLabel -> Concept -> SKOS correspondence across two heterogeneous datasets
	// Dataset A uses institutional names ("Gobierno Autónomo Municipal De La Paz", "Santa Cruz")
	// Dataset B uses numeric codes ("020101", "070101")
	// Each dataset maps its ValueLabels to its own Concepts, which in turn link via SKOS (external Wikidata URI or direct cross-dataset concept reference).
	datasetAMapping := domain.SemanticFieldMapping{
		FieldName: "entidad_municipal",
		ValueLabels: []domain.ValueLabelConceptEntry{
			{
				RawValue:  "Gobierno Autónomo Municipal De La Paz",
				ConceptID: "presupuesto_2025:concepts/gam_la_paz.md",
				AltLabels: []string{"GAMLP", "Alcaldía de La Paz"},
			},
			{
				RawValue:  "Gobierno Autónomo Municipal de Santa Cruz",
				ConceptID: "presupuesto_2025:concepts/gam_scz.md",
			},
		},
	}

	datasetBMapping := domain.SemanticFieldMapping{
		FieldName: "cod_ine",
		ValueLabels: []domain.ValueLabelConceptEntry{
			{
				RawValue:  "020101",
				ConceptID: "censo_2024:concepts/mun_020101.md",
				PrefLabel: "Nuestra Señora de La Paz",
			},
			{
				RawValue:  "070101",
				ConceptID: "censo_2024:concepts/mun_070101.md",
				PrefLabel: "Santa Cruz de la Sierra",
			},
		},
	}

	conceptCatalog := map[string]domain.ConceptNodeSKOS{
		"presupuesto_2025:concepts/gam_la_paz.md": {
			ConceptID:  "presupuesto_2025:concepts/gam_la_paz.md",
			PrefLabel:  "Gobierno Autónomo Municipal De La Paz",
			ExactMatch: []string{"http://www.wikidata.org/entity/Q1491"}, // Shared external ontology URI
		},
		"censo_2024:concepts/mun_020101.md": {
			ConceptID:  "censo_2024:concepts/mun_020101.md",
			Notation:   "020101",
			ExactMatch: []string{"http://www.wikidata.org/entity/Q1491"}, // Shared external ontology URI
		},
		"presupuesto_2025:concepts/gam_scz.md": {
			ConceptID:  "presupuesto_2025:concepts/gam_scz.md",
			ExactMatch: []string{"censo_2024:concepts/mun_070101.md"}, // Direct cross-dataset concept link
		},
		"censo_2024:concepts/mun_070101.md": {
			ConceptID: "censo_2024:concepts/mun_070101.md",
			Notation:  "070101",
			SKOS:      []string{"http://www.wikidata.org/entity/Q170688"},
		},
	}

	// 1. Resolve ValueLabel (including alt_labels) to Concept
	cid, _, ok := uc.ResolveValueLabelToConcept("GAMLP", datasetAMapping, conceptCatalog)
	if !ok || cid != "presupuesto_2025:concepts/gam_la_paz.md" {
		t.Fatalf("expected GAMLP to resolve to presupuesto_2025:concepts/gam_la_paz.md, got %q (ok=%v)", cid, ok)
	}

	// 2. Resolve cross-dataset correspondences via SKOS & cross-concept references
	corrs := uc.ResolveConceptCorrespondences(datasetAMapping, datasetBMapping, conceptCatalog)
	if len(corrs) != 2 {
		t.Fatalf("expected 2 resolved correspondences, got %d: %+v", len(corrs), corrs)
	}

	if corrs[0].SourceValue != "Gobierno Autónomo Municipal De La Paz" ||
		corrs[0].TargetValue != "020101" ||
		corrs[0].SharedAnchorURI != "http://www.wikidata.org/entity/Q1491" ||
		corrs[0].MatchPredicate != "skos:exactMatch" {
		t.Errorf("unexpected first correspondence (external SKOS ontology match): %+v", corrs[0])
	}

	if corrs[1].SourceValue != "Gobierno Autónomo Municipal de Santa Cruz" ||
		corrs[1].TargetValue != "070101" ||
		corrs[1].SharedAnchorURI != "censo_2024:concepts/mun_070101.md" ||
		corrs[1].MatchPredicate != "direct_concept_ref" {
		t.Errorf("unexpected second correspondence (direct cross-dataset concept match): %+v", corrs[1])
	}

	// 3. Deterministic cross-dataset SQL JOIN condition
	joinSQL := uc.GenerateCrossDatasetJoinSQL(datasetAMapping, "a", datasetBMapping, "b", conceptCatalog)
	expectedJoin := "(CASE WHEN a.entidad_municipal = 'Gobierno Autónomo Municipal De La Paz' THEN 'http://www.wikidata.org/entity/Q1491' WHEN a.entidad_municipal = 'Gobierno Autónomo Municipal de Santa Cruz' THEN 'censo_2024:concepts/mun_070101.md' ELSE a.entidad_municipal END) = (CASE WHEN b.cod_ine = '020101' THEN 'http://www.wikidata.org/entity/Q1491' WHEN b.cod_ine = '070101' THEN 'censo_2024:concepts/mun_070101.md' ELSE b.cod_ine END)"
	if joinSQL != expectedJoin {
		t.Errorf("unexpected cross-dataset join SQL:\ngot:  %s\nwant: %s", joinSQL, expectedJoin)
	}
}

func TestDiffAndAttestUseCases(t *testing.T) {
	verdict, err := AttestComputation("../../bundles/acme_retail/computations/revenue-ytd.md", "", map[string]string{"year": "2026"})
	if err != nil {
		t.Fatalf("AttestComputation failed: %v", err)
	}
	if verdict.Verdict != "verified" || verdict.CanonicalQueryHash == "" {
		t.Errorf("expected verified verdict with hash, got %+v", verdict)
	}

	// Test refuted when SQL is tampered
	tampered, _ := AttestComputation("../../bundles/acme_retail/computations/revenue-ytd.md", "SELECT 999999 AS revenue_usd", map[string]string{"year": "2026"})
	if tampered.Verdict != "refuted" {
		t.Errorf("expected refuted verdict for tampered SQL, got %+v", tampered)
	}
}

func TestGenerateCrossDatasetJoinSQL_ExcludesCloseMatchAndFailsFast(t *testing.T) {
	uc := NewSemanticAlignmentUseCase()
	srcMap := domain.SemanticFieldMapping{
		FieldName: "tipo",
		ValueLabels: []domain.ValueLabelConceptEntry{
			{RawValue: "A", ConceptID: "ds1:concepts/a.md"},
			{RawValue: "C", ConceptID: "ds1:concepts/c.md"},
		},
	}
	tgtMap := domain.SemanticFieldMapping{
		FieldName: "tipo",
		ValueLabels: []domain.ValueLabelConceptEntry{
			{RawValue: "B", ConceptID: "ds2:concepts/b.md"},
			{RawValue: "D", ConceptID: "ds2:concepts/d.md"},
		},
	}
	catalog := map[string]domain.ConceptNodeSKOS{
		"ds1:concepts/a.md": {ConceptID: "ds1:concepts/a.md", CloseMatch: []string{"http://www.wikidata.org/entity/Q999"}},
		"ds2:concepts/b.md": {ConceptID: "ds2:concepts/b.md", ExactMatch: []string{"http://www.wikidata.org/entity/Q999"}},
		"ds1:concepts/c.md": {ConceptID: "ds1:concepts/c.md", ExactMatch: []string{"http://www.wikidata.org/entity/Q100"}},
		"ds2:concepts/d.md": {ConceptID: "ds2:concepts/d.md", ExactMatch: []string{"http://www.wikidata.org/entity/Q100"}},
	}
	joinSQL := uc.GenerateCrossDatasetJoinSQL(srcMap, "a", tgtMap, "b", catalog)
	if strings.Contains(joinSQL, "'A'") || !strings.Contains(joinSQL, "'C'") {
		t.Errorf("expected skos:closeMatch pair (A, B) excluded and exact pair (C, D) kept, got: %s", joinSQL)
	}

	onlyCloseCatalog := map[string]domain.ConceptNodeSKOS{
		"ds1:concepts/a.md": {ConceptID: "ds1:concepts/a.md", CloseMatch: []string{"http://www.wikidata.org/entity/Q999"}},
		"ds2:concepts/b.md": {ConceptID: "ds2:concepts/b.md", ExactMatch: []string{"http://www.wikidata.org/entity/Q999"}},
	}
	emptySQL := uc.GenerateCrossDatasetJoinSQL(
		domain.SemanticFieldMapping{FieldName: "tipo", ValueLabels: []domain.ValueLabelConceptEntry{{RawValue: "A", ConceptID: "ds1:concepts/a.md"}}},
		"a",
		domain.SemanticFieldMapping{FieldName: "tipo", ValueLabels: []domain.ValueLabelConceptEntry{{RawValue: "B", ConceptID: "ds2:concepts/b.md"}}},
		"b",
		onlyCloseCatalog,
	)
	if emptySQL != "" {
		t.Errorf("expected empty join SQL when only closeMatch correspondences exist, got: %s", emptySQL)
	}
}

func TestAttestComputation_BundleContainedRelativePath(t *testing.T) {
	bundleDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(bundleDir, "index.md"), []byte("# Bundle\n"), 0644)
	metricsDir := filepath.Join(bundleDir, "metrics")
	compDir := filepath.Join(bundleDir, "computations")
	_ = os.MkdirAll(metricsDir, 0755)
	_ = os.MkdirAll(compDir, 0755)

	_ = os.WriteFile(filepath.Join(compDir, "revenue.sql"), []byte("SELECT SUM(amount) FROM sales WHERE yr = @year\n"), 0644)
	conceptMD := "---\ntype: Attested Computation\ntitle: Rev\nruntime: bigquery\ncomputation: ../computations/revenue.sql\nparameters:\n  - { name: year, type: integer }\n---\n"
	conceptPath := filepath.Join(metricsDir, "rev.md")
	_ = os.WriteFile(conceptPath, []byte(conceptMD), 0644)

	verdict, err := AttestComputation(conceptPath, "SELECT SUM(amount) FROM sales WHERE yr = 2026", map[string]string{"year": "2026"})
	if err != nil {
		t.Fatalf("AttestComputation failed: %v", err)
	}
	if verdict.Verdict != "verified" {
		t.Errorf("expected verified verdict for bundle-internal ../computations/revenue.sql, got %+v", verdict)
	}
}
