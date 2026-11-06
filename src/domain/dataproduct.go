package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyTitle       = errors.New("data product title cannot be empty")
	ErrInvalidType      = errors.New("data product type cannot be empty")
	ErrMissingDimension = errors.New("data product must specify at least one dimension")
)

// GeneratedSpec models the creation provenance per OKF v0.2 §5.1 & W3C PROV-O.
type GeneratedSpec struct {
	By     string   `json:"by,omitempty" yaml:"by,omitempty"`         // Actor: "Name <email>", URL, or agent:<id>
	At     string   `json:"at,omitempty" yaml:"at,omitempty"`         // ISO 8601 UTC timestamp
	From   []string `json:"from,omitempty" yaml:"from,omitempty"`     // Input URIs or relative concept paths
	Recipe string   `json:"recipe,omitempty" yaml:"recipe,omitempty"` // Pipeline script, prompt, or recipe URL
}

// VerificationRecord models an explicit verification assertion per OKF v0.2 §5.2.
type VerificationRecord struct {
	By     string `json:"by,omitempty" yaml:"by,omitempty"`         // Verifier actor
	At     string `json:"at,omitempty" yaml:"at,omitempty"`         // ISO 8601 UTC timestamp
	Method string `json:"method,omitempty" yaml:"method,omitempty"` // automated, human, cryptographic, peer-review
	Notes  string `json:"notes,omitempty" yaml:"notes,omitempty"`   // Verification observations
}

// FreshnessSpec models concept currency, cadence, and expiration per OKF v0.2 §5.4.
type FreshnessSpec struct {
	Cadence            string `json:"cadence,omitempty" yaml:"cadence,omitempty"`                         // ISO 8601 duration (e.g. P1D, P1W, P1M)
	Expires            string `json:"expires,omitempty" yaml:"expires,omitempty"`                         // ISO 8601 timestamp
	StalenessIndicator string `json:"staleness_indicator,omitempty" yaml:"staleness_indicator,omitempty"` // Query or signal to check if stale
}

// EpistemicActor represents an epistemic agent or verifier in the provenance chain (emitter or attester).
type EpistemicActor struct {
	By        string                 `json:"by" yaml:"by"`                             // human:<id>, agent:<id>, process:<id>
	At        time.Time              `json:"at" yaml:"at"`                             // ISO 8601 timestamp
	Role      string                 `json:"role,omitempty" yaml:"role,omitempty"`     // raw-data-producer, validator, auditor
	Verdict   string                 `json:"verdict,omitempty" yaml:"verdict,omitempty"` // verified, refuted, pending
	TrustTier string                 `json:"trust_tier,omitempty" yaml:"trust_tier,omitempty"` // unverified, machine-confirmed, human-reviewed
	Signature *Signature             `json:"signature,omitempty" yaml:"signature,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Lineage tracks provenance, version, and timestamps according to PE3 and ODKF epistemic stack.
type Lineage struct {
	Sources         []string               `json:"sources" yaml:"sources"`
	Version         string                 `json:"version" yaml:"version"`
	UpdatedAt       time.Time              `json:"updated_at" yaml:"updated_at"`
	Emitter         *EpistemicActor        `json:"emitter,omitempty" yaml:"emitter,omitempty"`
	Attester        *EpistemicActor        `json:"attester,omitempty" yaml:"attester,omitempty"`
	Ref             string                 `json:"ref,omitempty" yaml:"ref,omitempty"` // By-Reference lineage external file
	Transformations []LineageTransformation `json:"transformations,omitempty" yaml:"transformations,omitempty"`
}

// ResourceField represents a column/field definition inside a tabular Resource schema.
type ResourceField struct {
	Name        string            `json:"name" yaml:"name"`
	Type        string            `json:"type,omitempty" yaml:"type,omitempty"`
	Title       string            `json:"title,omitempty" yaml:"title,omitempty"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	RDFType     string            `json:"rdfType,omitempty" yaml:"rdfType,omitempty"`
	Concept     string            `json:"concept,omitempty" yaml:"concept,omitempty"`
	Unit        string            `json:"unit,omitempty" yaml:"unit,omitempty"`
	Categories  []string          `json:"categories,omitempty" yaml:"categories,omitempty"`
	Custom      map[string]string `json:"custom,omitempty" yaml:"custom,omitempty"`
}

// CLIInputParam models an input argument/flag for an OWL/OKF-described CLI tool (`owl-cli` / `okf-cli`).
type CLIInputParam struct {
	Name        string `json:"name" yaml:"name"`
	Flag        string `json:"flag,omitempty" yaml:"flag,omitempty"`
	Position    int    `json:"position,omitempty" yaml:"position,omitempty"`
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	Concept     string `json:"concept,omitempty" yaml:"concept,omitempty"`
	Required    bool   `json:"required,omitempty" yaml:"required,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// CLIOutputSpec models the semantic output of a CLI tool (`owl-cli` / `okf-cli`).
type CLIOutputSpec struct {
	Format      string `json:"format,omitempty" yaml:"format,omitempty"`
	Concept     string `json:"concept,omitempty" yaml:"concept,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
}

// Contract defines a syntactic, service, or CLI contract adhering to OKF/ODKF v0.2+.
// Supports inline specification, by-reference resolution (path/ref), and OWL-CLI tool definitions.
type Contract struct {
	Type       string                 `json:"type" yaml:"type"`
	Path       string                 `json:"path,omitempty" yaml:"path,omitempty"`
	Ref        string                 `json:"ref,omitempty" yaml:"ref,omitempty"`               // By-Reference resolution (local path or remote URL)
	Binary     string                 `json:"binary,omitempty" yaml:"binary,omitempty"`         // For type: owl-cli / okf-cli
	Subcommand string                 `json:"subcommand,omitempty" yaml:"subcommand,omitempty"` // For type: owl-cli / okf-cli
	Inputs     []CLIInputParam        `json:"inputs,omitempty" yaml:"inputs,omitempty"`         // For type: owl-cli / okf-cli
	Outputs    *CLIOutputSpec         `json:"outputs,omitempty" yaml:"outputs,omitempty"`       // For type: owl-cli / okf-cli
	Inline     map[string]interface{} `json:"inline,omitempty" yaml:"inline,omitempty"`         // Inline contract specification
}

// Resource represents an individual tabular or binary asset within a Data Product.
type Resource struct {
	Name        string          `json:"name" yaml:"name"`
	Path        string          `json:"path" yaml:"path"`
	Format      string          `json:"format" yaml:"format"`
	Description string          `json:"description,omitempty" yaml:"description,omitempty"`
	HashSHA256  string          `json:"hash_sha256,omitempty" yaml:"hash_sha256,omitempty"`
	BytesCount  int64           `json:"bytes_count,omitempty" yaml:"bytes_count,omitempty"`
	Fields      []ResourceField `json:"fields,omitempty" yaml:"fields,omitempty"`
}

// SpatialCoverage encapsulates geospatial extent and resolution per ISO 19115 / DCAT v3.
type SpatialCoverage struct {
	Country     string    `json:"country,omitempty" yaml:"country,omitempty"`
	Regions     []string  `json:"regions,omitempty" yaml:"regions,omitempty"`
	BBox        []float64 `json:"bbox,omitempty" yaml:"bbox,omitempty"` // [minX, minY, maxX, maxY]
	Granularity string    `json:"granularity,omitempty" yaml:"granularity,omitempty"`
}

// TemporalCoverage encapsulates temporal extent and frequency per ISO 8601 / DCAT.
type TemporalCoverage struct {
	Start     string `json:"start,omitempty" yaml:"start,omitempty"`
	End       string `json:"end,omitempty" yaml:"end,omitempty"`
	Frequency string `json:"frequency,omitempty" yaml:"frequency,omitempty"`
	Timezone  string `json:"timezone,omitempty" yaml:"timezone,omitempty"`
}

// QualityProfile encapsulates completeness and quality verification rules.
type QualityProfile struct {
	Status       string             `json:"status,omitempty" yaml:"status,omitempty"`
	Completeness float64            `json:"completeness,omitempty" yaml:"completeness,omitempty"`
	RowCount     int64              `json:"row_count,omitempty" yaml:"row_count,omitempty"`
	Checks       []QualityCheckRule `json:"checks,omitempty" yaml:"checks,omitempty"`
}

// QualityCheckRule defines a declarative assertion on data values.
type QualityCheckRule struct {
	Rule        string                 `json:"rule" yaml:"rule"`
	TargetField string                 `json:"field,omitempty" yaml:"field,omitempty"`
	Params      map[string]interface{} `json:"params,omitempty" yaml:"params,omitempty"`
}

// ValueLabelConceptEntry maps a raw cell value or label in a dataset column to a Concept
// and its SKOS ontological bridges (external ontologies or concepts in other datasets).
type ValueLabelConceptEntry struct {
	RawValue      string   `json:"raw_value" yaml:"raw_value"`                                     // Raw code or label in dataset (e.g. "Gobierno Autónomo Municipal De La Paz", "020101", "LPZ")
	ConceptID     string   `json:"concept_id" yaml:"concept_id"`                                   // Local or canonical concept URI/path (e.g. "concepts/la_paz.md" or "concept:municipios:la_paz")
	PrefLabel     string   `json:"pref_label,omitempty" yaml:"pref_label,omitempty"`               // skos:prefLabel
	AltLabels     []string `json:"alt_labels,omitempty" yaml:"alt_labels,omitempty"`               // skos:altLabel / lexical variants
	Notation      string   `json:"notation,omitempty" yaml:"notation,omitempty"`                   // skos:notation (canonical code, e.g. "020101")
	ExactMatch    []string `json:"exact_match,omitempty" yaml:"exact_match,omitempty"`             // skos:exactMatch (external ontologies like wd:Q1491 or concepts in other datasets)
	CloseMatch    []string `json:"close_match,omitempty" yaml:"close_match,omitempty"`             // skos:closeMatch
	Broader       []string `json:"broader,omitempty" yaml:"broader,omitempty"`                     // skos:broader
	Related       []string `json:"related,omitempty" yaml:"related,omitempty"`                     // skos:related
}

// ConceptNodeSKOS represents a standalone OKF concept's SKOS semantic graph links.
// Enables resolving correspondences across two datasets that share an external SKOS anchor
// or cross-reference each other's concept documents.
type ConceptNodeSKOS struct {
	ConceptID  string   `json:"concept_id" yaml:"concept_id"`                       // Concept path or URI (e.g. "bolivia:censo:concepts/la_paz.md")
	PrefLabel  string   `json:"pref_label,omitempty" yaml:"pref_label,omitempty"`   // skos:prefLabel
	AltLabels  []string `json:"alt_labels,omitempty" yaml:"alt_labels,omitempty"`   // skos:altLabel
	Notation   string   `json:"notation,omitempty" yaml:"notation,omitempty"`       // skos:notation
	SKOS       []string `json:"skos,omitempty" yaml:"skos,omitempty"`               // General SKOS / Wikidata / Schema.org URIs
	ExactMatch []string `json:"exact_match,omitempty" yaml:"exact_match,omitempty"` // skos:exactMatch (external URIs or other dataset concept URIs)
	CloseMatch []string `json:"close_match,omitempty" yaml:"close_match,omitempty"` // skos:closeMatch
	Broader    []string `json:"broader,omitempty" yaml:"broader,omitempty"`         // skos:broader
}

// ConceptCorrespondence represents a resolved semantic bridge between a value label in Dataset A
// and a matching value label in Dataset B via their linked Concepts and SKOS anchors.
type ConceptCorrespondence struct {
	SourceValue     string `json:"source_value"`               // ValueLabel in Dataset A
	SourceConceptID string `json:"source_concept_id"`          // Concept linked to SourceValue
	TargetValue     string `json:"target_value"`               // ValueLabel in Dataset B
	TargetConceptID string `json:"target_concept_id"`          // Concept linked to TargetValue
	SharedAnchorURI string `json:"shared_anchor_uri"`          // Shared SKOS/Wikidata URI or direct cross-dataset concept reference
	MatchPredicate  string `json:"match_predicate"`            // "skos:exactMatch", "skos:closeMatch", "direct_concept_ref", "skos_anchor"
}

// SemanticFieldMapping maps a raw tabular column to an ODKF concept document,
// physical unit (QUDT/UCUM), aggregation behavior, and per-ValueLabel concept mappings.
type SemanticFieldMapping struct {
	FieldName           string                   `json:"field_name" yaml:"field_name"`
	ConceptRef          string                   `json:"concept_ref,omitempty" yaml:"concept_ref,omitempty"`
	Unit                string                   `json:"unit,omitempty" yaml:"unit,omitempty"`                                 // UCUM / QUDT unit code (e.g. "ug/m3", "BOB", "MWh")
	AggregationBehavior string                   `json:"aggregation_behavior,omitempty" yaml:"aggregation_behavior,omitempty"` // additive, semi_additive, non_additive
	ValueMapping        map[string]string        `json:"value_mapping,omitempty" yaml:"value_mapping,omitempty"`               // Shorthand: raw_value -> concept_id
	ValueLabels         []ValueLabelConceptEntry `json:"value_labels,omitempty" yaml:"value_labels,omitempty"`                 // Rich ValueLabel -> Concept -> SKOS definitions
}

// AttestationVerdict represents the standardized JSON wire contract for Attested Computation verification (§10.2, §12).
type AttestationVerdict struct {
	Verdict            string `json:"verdict" yaml:"verdict"`                                   // "verified", "refuted", or "error"
	CanonicalQueryHash string `json:"canonical_query_hash,omitempty" yaml:"canonical_query_hash,omitempty"` // sha256:<hex64>
	CheckedAt          string `json:"checked_at" yaml:"checked_at"`                             // ISO 8601 UTC timestamp
	Reason             string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

// ConformanceProfile enumerates the W3C QA Framework modular conformance levels for OKF/ODKF.
type ConformanceProfile string

const (
	ProfileLevel0Core      ConformanceProfile = "level-0-core"      // Core Knowledge Bundle (Markdown + YAML, links, provenance, trust)
	ProfileLevel1Contracts ConformanceProfile = "level-1-contracts" // Data Product & Contracts (Frictionless/DCAT, triads, SKOS, zero_microdata, Attested Computation)
	ProfileLevel2Sovereign ConformanceProfile = "level-2-sovereign" // Sovereign Federated / P2P (okf://, DNS @, Ed25519/SSH, HTTP Range #fragment, Kademlia DHT)
)

// PublicationTargetResult contains the execution status for a publication target.
type PublicationTargetResult struct {
	Target         string `json:"target"`
	Success        bool   `json:"success"`
	DestinationURI string `json:"destination_uri,omitempty"`
	Message        string `json:"message,omitempty"`
	Error          string `json:"error,omitempty"`
}

// Signature encapsulates cryptographic authorship proof (SSH / OpenPGP / Ed25519) per PE4 & ADR 0015.
type Signature struct {
	KeyID     string `json:"key_id" yaml:"key_id"`
	Algorithm string `json:"algorithm" yaml:"algorithm"`
	Signature string `json:"signature" yaml:"signature"`
	Verified  bool   `json:"verified" yaml:"verified"`
}

// Manifest represents the frontmatter header specified in OKF / ODKF v0.2.
type Manifest struct {
	Type           string                 `json:"type" yaml:"type"`
	Title          string                 `json:"title" yaml:"title"`
	Description    string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Resource       string                 `json:"resource,omitempty" yaml:"resource,omitempty"`
	Tags           StringList             `json:"tags,omitempty" yaml:"tags,omitempty"`
	Dimensions     StringList             `json:"dimensions" yaml:"dimensions"`
	SKOS           StringList             `json:"skos,omitempty" yaml:"skos,omitempty"`
	ExactMatch     StringList             `json:"exact_match,omitempty" yaml:"exact_match,omitempty"`
	CloseMatch     StringList             `json:"close_match,omitempty" yaml:"close_match,omitempty"`
	Broader        StringList             `json:"broader,omitempty" yaml:"broader,omitempty"`
	Nature         string                 `json:"nature,omitempty" yaml:"nature,omitempty"` // structured, semi-structured, unstructured
	Status         string                 `json:"status,omitempty" yaml:"status,omitempty"` // draft, proposed, active, deprecated, retired, superseded
	Supersedes     string                 `json:"supersedes,omitempty" yaml:"supersedes,omitempty"` // URI of superseded concept (Popperian falsability)
	SupersededBy   string                 `json:"superseded_by,omitempty" yaml:"superseded_by,omitempty"`
	Generated      *GeneratedSpec         `json:"generated,omitempty" yaml:"generated,omitempty"`
	Verified       []VerificationRecord   `json:"verified,omitempty" yaml:"verified,omitempty"`
	Freshness      *FreshnessSpec         `json:"freshness,omitempty" yaml:"freshness,omitempty"`
	Contracts      []Contract             `json:"contracts,omitempty" yaml:"contracts,omitempty"`
	Policy         interface{}            `json:"policy,omitempty" yaml:"policy,omitempty"`
	Policies       []PolicySpec           `json:"policies,omitempty" yaml:"policies,omitempty"`
	SemanticFields []SemanticFieldMapping `json:"semantic_fields,omitempty" yaml:"semantic_fields,omitempty"`
	Spatial        *SpatialCoverage       `json:"spatial,omitempty" yaml:"spatial,omitempty"`
	Temporal       *TemporalCoverage      `json:"temporal,omitempty" yaml:"temporal,omitempty"`
	Quality        *QualityProfile        `json:"quality,omitempty" yaml:"quality,omitempty"`
	Lineage        Lineage                `json:"lineage" yaml:"lineage"`
	Emitter        *EpistemicActor        `json:"emitter,omitempty" yaml:"emitter,omitempty"`
	Attester       *EpistemicActor        `json:"attester,omitempty" yaml:"attester,omitempty"`
	Signatures     []Signature            `json:"signatures,omitempty" yaml:"signatures,omitempty"`
}

// DataProduct is the Aggregate Root representing an Open Knowledge Format node.
type DataProduct struct {
	ID          string     `json:"id"`
	Manifest    Manifest   `json:"manifest"`
	Description string     `json:"description"`
	Resources   []Resource `json:"resources"`
	RawContent  string     `json:"raw_content,omitempty"`
}

// NewDataProduct instantiates and validates a DataProduct Aggregate Root.
func NewDataProduct(id string, manifest Manifest, description string, resources []Resource) (*DataProduct, error) {
	dp := &DataProduct{
		ID:          id,
		Manifest:    manifest,
		Description: description,
		Resources:   resources,
	}

	if err := dp.Validate(); err != nil {
		return nil, err
	}
	return dp, nil
}

// Validate checks domain invariants for OKF / ODKF v0.2 conformance.
// Per SPEC.md §4.1 and §11, type values are open and unregistered; consumers MUST NOT reject unknown types.
func (dp *DataProduct) Validate() error {
	if strings.TrimSpace(dp.Manifest.Title) == "" {
		return ErrEmptyTitle
	}

	if strings.TrimSpace(dp.Manifest.Type) == "" {
		return ErrInvalidType
	}

	return nil
}
