package usecases

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/datosbolivia/open-knowledge-format/src/domain"
	"github.com/datosbolivia/open-knowledge-format/src/ports/outbound"
)

// PublishDatasetUseCase orchestrates publishing an ODKF dataset to multiple platforms.
type PublishDatasetUseCase struct {
	reader     outbound.UnifiedMetadataReaderPort
	publishers []outbound.PublisherTargetPort
}

// NewPublishDatasetUseCase creates a new instance of PublishDatasetUseCase.
func NewPublishDatasetUseCase(reader outbound.UnifiedMetadataReaderPort, publishers ...outbound.PublisherTargetPort) *PublishDatasetUseCase {
	return &PublishDatasetUseCase{
		reader:     reader,
		publishers: publishers,
	}
}

// RegisterPublisher appends a new publisher adapter.
func (uc *PublishDatasetUseCase) RegisterPublisher(p outbound.PublisherTargetPort) {
	uc.publishers = append(uc.publishers, p)
}

// Publish executes publishing across the requested targets.
func (uc *PublishDatasetUseCase) Publish(
	ctx context.Context,
	datasetPath string,
	targets []string,
	options map[string]interface{},
) ([]domain.PublicationTargetResult, error) {
	if len(targets) == 0 {
		targets = []string{"local"}
	}

	var manifest *domain.PackageManifest
	var err error
	if uc.reader != nil {
		manifest, err = uc.reader.ReadPackage(ctx, datasetPath)
	}
	if err != nil || manifest == nil {
		// Fallback minimal manifest
		manifest = &domain.PackageManifest{
			Name:  datasetPath,
			Title: datasetPath,
		}
	}

	var results []domain.PublicationTargetResult
	for _, target := range targets {
		matched := false
		for _, pub := range uc.publishers {
			if pub.CanPublish(target) {
				matched = true
				res, pubErr := pub.Publish(ctx, manifest, datasetPath, options)
				if pubErr != nil {
					res.Success = false
					res.Error = pubErr.Error()
				}
				results = append(results, res)
				break
			}
		}
		if !matched {
			results = append(results, domain.PublicationTargetResult{
				Target:  target,
				Success: false,
				Error:   fmt.Sprintf("no publisher adapter registered for target '%s'", target),
			})
		}
	}

	return results, nil
}

// SemanticAlignmentUseCase provides domain-agnostic concept correspondence resolution
// via the OKF semantic chain: Dataset ValueLabel -> Concept -> SKOS (External Ontologies & Cross-Dataset Concepts),
// as well as deterministic SQL CASE and cross-dataset JOIN expression generation.
type SemanticAlignmentUseCase struct{}

// NewSemanticAlignmentUseCase creates a new instance of SemanticAlignmentUseCase.
func NewSemanticAlignmentUseCase() *SemanticAlignmentUseCase {
	return &SemanticAlignmentUseCase{}
}

// normalizeLexicalToken normalizes whitespace and case for matching ValueLabels and skos:altLabel entries.
func normalizeLexicalToken(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(s)), " "))
}

// buildEffectiveConceptNodes merges inline ValueLabelConceptEntry definitions with the ConceptNodeSKOS catalog.
func buildEffectiveConceptNodes(
	mapping domain.SemanticFieldMapping,
	conceptCatalog map[string]domain.ConceptNodeSKOS,
) ([]domain.ValueLabelConceptEntry, map[string]domain.ConceptNodeSKOS) {
	entries := make([]domain.ValueLabelConceptEntry, 0, len(mapping.ValueLabels)+len(mapping.ValueMapping))
	seenRaw := make(map[string]bool)

	for _, vl := range mapping.ValueLabels {
		entries = append(entries, vl)
		seenRaw[vl.RawValue] = true
	}

	// Include shorthand ValueMapping entries in deterministic order
	rawKeys := make([]string, 0, len(mapping.ValueMapping))
	for k := range mapping.ValueMapping {
		if !seenRaw[k] {
			rawKeys = append(rawKeys, k)
		}
	}
	sort.Strings(rawKeys)
	for _, k := range rawKeys {
		entries = append(entries, domain.ValueLabelConceptEntry{
			RawValue:  k,
			ConceptID: mapping.ValueMapping[k],
		})
	}

	nodes := make(map[string]domain.ConceptNodeSKOS)
	for k, v := range conceptCatalog {
		nodes[k] = v
	}

	// Enrich or synthesize ConceptNodeSKOS from inline ValueLabelConceptEntry fields
	for _, vl := range entries {
		cid := strings.TrimSpace(vl.ConceptID)
		if cid == "" {
			continue
		}
		existing, ok := nodes[cid]
		if !ok {
			existing = domain.ConceptNodeSKOS{ConceptID: cid}
		}
		if existing.PrefLabel == "" && vl.PrefLabel != "" {
			existing.PrefLabel = vl.PrefLabel
		}
		if existing.Notation == "" && vl.Notation != "" {
			existing.Notation = vl.Notation
		}
		existing.AltLabels = appendUniqueStrings(existing.AltLabels, vl.AltLabels...)
		existing.ExactMatch = appendUniqueStrings(existing.ExactMatch, vl.ExactMatch...)
		existing.CloseMatch = appendUniqueStrings(existing.CloseMatch, vl.CloseMatch...)
		existing.Broader = appendUniqueStrings(existing.Broader, vl.Broader...)
		nodes[cid] = existing
	}

	return entries, nodes
}

func appendUniqueStrings(base []string, items ...string) []string {
	seen := make(map[string]bool, len(base))
	for _, b := range base {
		seen[strings.TrimSpace(b)] = true
	}
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" && !seen[trimmed] {
			base = append(base, trimmed)
			seen[trimmed] = true
		}
	}
	return base
}

// ResolveValueLabelToConcept resolves a raw dataset value or lexical label into its linked ConceptID
// and ConceptNodeSKOS using the dataset's ValueLabel mappings and the SKOS concept catalog.
func (uc *SemanticAlignmentUseCase) ResolveValueLabelToConcept(
	rawInput string,
	mapping domain.SemanticFieldMapping,
	conceptCatalog map[string]domain.ConceptNodeSKOS,
) (string, domain.ConceptNodeSKOS, bool) {
	normInput := normalizeLexicalToken(rawInput)
	if normInput == "" {
		return "", domain.ConceptNodeSKOS{}, false
	}

	entries, nodes := buildEffectiveConceptNodes(mapping, conceptCatalog)
	for _, vl := range entries {
		cid := strings.TrimSpace(vl.ConceptID)
		node := nodes[cid]

		if normalizeLexicalToken(vl.RawValue) == normInput ||
			(vl.Notation != "" && normalizeLexicalToken(vl.Notation) == normInput) ||
			(vl.PrefLabel != "" && normalizeLexicalToken(vl.PrefLabel) == normInput) ||
			(node.Notation != "" && normalizeLexicalToken(node.Notation) == normInput) ||
			(node.PrefLabel != "" && normalizeLexicalToken(node.PrefLabel) == normInput) {
			return cid, node, true
		}
		for _, alt := range vl.AltLabels {
			if normalizeLexicalToken(alt) == normInput {
				return cid, node, true
			}
		}
		for _, alt := range node.AltLabels {
			if normalizeLexicalToken(alt) == normInput {
				return cid, node, true
			}
		}
	}
	return "", domain.ConceptNodeSKOS{}, false
}

// ResolveConceptCorrespondences resolves semantic correspondences between Dataset A's ValueLabels
// and Dataset B's ValueLabels through their linked Concepts and SKOS ontological bridges:
// ValueLabel A -> Concept A <-> [Shared External Ontology URI or Direct Cross-Dataset Concept Link] <-> Concept B <- ValueLabel B.
func (uc *SemanticAlignmentUseCase) ResolveConceptCorrespondences(
	sourceMapping domain.SemanticFieldMapping,
	targetMapping domain.SemanticFieldMapping,
	conceptCatalog map[string]domain.ConceptNodeSKOS,
) []domain.ConceptCorrespondence {
	srcEntries, srcNodes := buildEffectiveConceptNodes(sourceMapping, conceptCatalog)
	tgtEntries, tgtNodes := buildEffectiveConceptNodes(targetMapping, srcNodes)
	// Ensure srcNodes also includes any inline definitions from tgtNodes
	for k, v := range tgtNodes {
		srcNodes[k] = v
	}

	var correspondences []domain.ConceptCorrespondence
	for _, srcVL := range srcEntries {
		srcCID := strings.TrimSpace(srcVL.ConceptID)
		if srcCID == "" {
			continue
		}
		srcNode := srcNodes[srcCID]

		for _, tgtVL := range tgtEntries {
			tgtCID := strings.TrimSpace(tgtVL.ConceptID)
			if tgtCID == "" {
				continue
			}
			tgtNode := srcNodes[tgtCID]

			if sharedURI, predicate, ok := matchConceptNodesSKOS(srcNode, tgtNode); ok {
				correspondences = append(correspondences, domain.ConceptCorrespondence{
					SourceValue:     srcVL.RawValue,
					SourceConceptID: srcCID,
					TargetValue:     tgtVL.RawValue,
					TargetConceptID: tgtCID,
					SharedAnchorURI: sharedURI,
					MatchPredicate:  predicate,
				})
			}
		}
	}

	sort.Slice(correspondences, func(i, j int) bool {
		if correspondences[i].SourceValue != correspondences[j].SourceValue {
			return correspondences[i].SourceValue < correspondences[j].SourceValue
		}
		return correspondences[i].TargetValue < correspondences[j].TargetValue
	})

	return correspondences
}

// matchConceptNodesSKOS checks if two ConceptNodeSKOS instances correspond via:
// 1. Identical ConceptID or direct cross-dataset concept reference in exact_match / skos / close_match.
// 2. Shared external ontology URI in skos:exactMatch.
// 3. Shared external ontology URI in skos (general ontological anchor).
// 4. Shared external ontology URI in skos:closeMatch.
func matchConceptNodesSKOS(a, b domain.ConceptNodeSKOS) (sharedAnchor string, predicate string, matched bool) {
	if a.ConceptID != "" && a.ConceptID == b.ConceptID {
		return a.ConceptID, "direct_concept_ref", true
	}

	// Direct cross-dataset concept link (Concept A references Concept B or vice versa)
	if containsString(a.ExactMatch, b.ConceptID) || containsString(a.SKOS, b.ConceptID) {
		return b.ConceptID, "direct_concept_ref", true
	}
	if containsString(b.ExactMatch, a.ConceptID) || containsString(b.SKOS, a.ConceptID) {
		return a.ConceptID, "direct_concept_ref", true
	}
	if containsString(a.CloseMatch, b.ConceptID) {
		return b.ConceptID, "skos:closeMatch", true
	}
	if containsString(b.CloseMatch, a.ConceptID) {
		return a.ConceptID, "skos:closeMatch", true
	}

	// Shared external SKOS / ontology URI in exact_match
	if shared := firstIntersection(a.ExactMatch, b.ExactMatch); shared != "" {
		return shared, "skos:exactMatch", true
	}
	if shared := firstIntersection(a.ExactMatch, b.SKOS); shared != "" {
		return shared, "skos:exactMatch", true
	}
	if shared := firstIntersection(a.SKOS, b.ExactMatch); shared != "" {
		return shared, "skos:exactMatch", true
	}

	// Shared general SKOS anchor
	if shared := firstIntersection(a.SKOS, b.SKOS); shared != "" {
		return shared, "skos_anchor", true
	}

	// Shared close_match (non-transitive per §14.3.1: at most 1 hop via close_match, never close_match <-> close_match)
	exactA := appendUniqueStrings(appendUniqueStrings(nil, a.ExactMatch...), a.SKOS...)
	exactB := appendUniqueStrings(appendUniqueStrings(nil, b.ExactMatch...), b.SKOS...)
	if shared := firstIntersection(a.CloseMatch, exactB); shared != "" {
		return shared, "skos:closeMatch", true
	}
	if shared := firstIntersection(exactA, b.CloseMatch); shared != "" {
		return shared, "skos:closeMatch", true
	}

	return "", "", false
}

func containsString(slice []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}
	for _, item := range slice {
		if strings.TrimSpace(item) == target {
			return true
		}
	}
	return false
}

func firstIntersection(a, b []string) string {
	setB := make(map[string]bool, len(b))
	for _, item := range b {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			setB[trimmed] = true
		}
	}
	for _, item := range a {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" && setB[trimmed] {
			return trimmed
		}
	}
	return ""
}

// GenerateCaseExpression builds a deterministic SQL CASE expression mapping raw ValueLabels to their ConceptIDs.
func (uc *SemanticAlignmentUseCase) GenerateCaseExpression(mapping domain.SemanticFieldMapping, tablePrefix string) string {
	col := mapping.FieldName
	if tablePrefix != "" {
		col = fmt.Sprintf("%s.%s", tablePrefix, mapping.FieldName)
	}

	valueToConcept := make(map[string]string)
	for k, v := range mapping.ValueMapping {
		valueToConcept[k] = v
	}
	for _, vl := range mapping.ValueLabels {
		if vl.RawValue != "" && vl.ConceptID != "" {
			valueToConcept[vl.RawValue] = vl.ConceptID
		}
		for _, alt := range vl.AltLabels {
			if alt != "" && vl.ConceptID != "" {
				valueToConcept[alt] = vl.ConceptID
			}
		}
	}

	if len(valueToConcept) == 0 {
		return col
	}

	rawValues := make([]string, 0, len(valueToConcept))
	for rawVal := range valueToConcept {
		rawValues = append(rawValues, rawVal)
	}
	sort.Strings(rawValues)

	whenClauses := make([]string, 0, len(rawValues))
	for _, rawVal := range rawValues {
		conceptID := valueToConcept[rawVal]
		escapedVal := strings.ReplaceAll(rawVal, "'", "''")
		escapedConcept := strings.ReplaceAll(conceptID, "'", "''")
		whenClauses = append(whenClauses, fmt.Sprintf("WHEN %s = '%s' THEN '%s'", col, escapedVal, escapedConcept))
	}

	return fmt.Sprintf("(CASE %s ELSE %s END)", strings.Join(whenClauses, " "), col)
}

// GenerateCrossDatasetJoinSQL builds a deterministic ANSI SQL join predicate that bridges Dataset A and Dataset B
// using their resolved ValueLabel -> Concept -> SKOS correspondences.
func (uc *SemanticAlignmentUseCase) GenerateCrossDatasetJoinSQL(
	sourceMapping domain.SemanticFieldMapping,
	sourceAlias string,
	targetMapping domain.SemanticFieldMapping,
	targetAlias string,
	conceptCatalog map[string]domain.ConceptNodeSKOS,
) string {
	correspondences := uc.ResolveConceptCorrespondences(sourceMapping, targetMapping, conceptCatalog)
	srcProjected := domain.SemanticFieldMapping{
		FieldName:    sourceMapping.FieldName,
		ValueMapping: make(map[string]string),
	}
	tgtProjected := domain.SemanticFieldMapping{
		FieldName:    targetMapping.FieldName,
		ValueMapping: make(map[string]string),
	}

	for _, corr := range correspondences {
		srcProjected.ValueMapping[corr.SourceValue] = corr.SharedAnchorURI
		tgtProjected.ValueMapping[corr.TargetValue] = corr.SharedAnchorURI
	}

	leftSQL := uc.GenerateCaseExpression(srcProjected, sourceAlias)
	rightSQL := uc.GenerateCaseExpression(tgtProjected, targetAlias)
	return fmt.Sprintf("%s = %s", leftSQL, rightSQL)
}
