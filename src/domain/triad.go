package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrInvalidTriadFormat = errors.New("invalid triad format: expected '[catalog:]dataset:resource', 'okf://catalog/dataset/resource', or '<dataset>.<resource>@<catalog>'")
	tableRefSQLRegex      = regexp.MustCompile(`(?i)\b(?:FROM|JOIN)\s+(?:ONLY\s+)?(?:["']([^"']+)["']|([a-zA-Z0-9_\-\.:/@#]+))`)
	quotedColonRegex      = regexp.MustCompile(`["']([^"':\s]+:[^"']+)["']`)
	quotedSlashRegex      = regexp.MustCompile(`["']([^"'\s]+/[^"'\s]+)["']`)
	quotedAtRegex         = regexp.MustCompile(`["']([a-zA-Z0-9_\-]+\.[a-zA-Z0-9_\-]+@[a-zA-Z0-9_\-\.#/]+)["']`)
)

var (
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

// NormalizeTriadSegment cleans and standardizes a triad component to lowercase with hyphens replacing whitespace.
func NormalizeTriadSegment(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ""
	}
	replaced := whitespaceRegex.ReplaceAllString(trimmed, "-")
	return strings.ToLower(replaced)
}

// ResourceTriad represents the canonical address '[catalog:]dataset:resource[#fragment]' in the federated DataMesh.
type ResourceTriad struct {
	Catalog  string `json:"catalog,omitempty"`
	Dataset  string `json:"dataset"`
	Resource string `json:"resource"`
	Fragment string `json:"fragment,omitempty"`
}

// String returns the canonical triad string representation in normalized format.
func (t ResourceTriad) String() string {
	cat := NormalizeTriadSegment(t.Catalog)
	ds := NormalizeTriadSegment(t.Dataset)
	res := NormalizeTriadSegment(t.Resource)
	base := fmt.Sprintf("%s:%s", ds, res)
	if cat != "" {
		base = fmt.Sprintf("%s:%s:%s", cat, ds, res)
	}
	if strings.TrimSpace(t.Fragment) != "" {
		return fmt.Sprintf("%s#%s", base, strings.TrimPrefix(strings.TrimSpace(t.Fragment), "#"))
	}
	return base
}

// FullURI returns the URI representation with datamesh:// scheme in normalized format.
func (t ResourceTriad) FullURI() string {
	cat := NormalizeTriadSegment(t.Catalog)
	ds := NormalizeTriadSegment(t.Dataset)
	res := NormalizeTriadSegment(t.Resource)
	base := fmt.Sprintf("datamesh://%s/%s", ds, res)
	if cat != "" {
		base = fmt.Sprintf("datamesh://%s/%s/%s", cat, ds, res)
	}
	if strings.TrimSpace(t.Fragment) != "" {
		return fmt.Sprintf("%s#%s", base, strings.TrimPrefix(strings.TrimSpace(t.Fragment), "#"))
	}
	return base
}

// OKFURI returns the canonical okf:// URI representation per SPEC.md §14.4 & §15.2.1.
func (t ResourceTriad) OKFURI() string {
	cat := NormalizeTriadSegment(t.Catalog)
	ds := NormalizeTriadSegment(t.Dataset)
	res := NormalizeTriadSegment(t.Resource)
	base := fmt.Sprintf("okf://%s/%s", ds, res)
	if cat != "" {
		base = fmt.Sprintf("okf://%s/%s/%s", cat, ds, res)
	}
	if strings.TrimSpace(t.Fragment) != "" {
		return fmt.Sprintf("%s#%s", base, strings.TrimPrefix(strings.TrimSpace(t.Fragment), "#"))
	}
	return base
}

// SovereignDNS returns the reverse-lookup Sovereign DNS representation '<dataset>.<resource>@<catalog>[#fragment]'.
func (t ResourceTriad) SovereignDNS() string {
	cat := NormalizeTriadSegment(t.Catalog)
	ds := NormalizeTriadSegment(t.Dataset)
	res := NormalizeTriadSegment(t.Resource)
	base := fmt.Sprintf("%s.%s", ds, res)
	if cat != "" {
		base = fmt.Sprintf("%s.%s@%s", ds, res, cat)
	}
	if strings.TrimSpace(t.Fragment) != "" {
		return fmt.Sprintf("%s#%s", base, strings.TrimPrefix(strings.TrimSpace(t.Fragment), "#"))
	}
	return base
}

// ParseTriad parses a string in 'catalogo:dataset:resource', 'dataset:resource',
// 'okf://[catalogo/]dataset/resource', 'datamesh://[catalogo/]dataset/resource',
// or Sovereign DNS '<dataset>.<resource>@<catalog>' format into a normalized ResourceTriad.
// Also preserves optional RFC 3986 fragment identifiers ('#subpath') for composite archives.
func ParseTriad(raw string) (*ResourceTriad, error) {
	trimmed := strings.Trim(strings.TrimSpace(raw), "\"'`")
	if trimmed == "" {
		return nil, ErrInvalidTriadFormat
	}

	fragment := ""
	if hashIdx := strings.Index(trimmed, "#"); hashIdx != -1 {
		fragment = strings.TrimSpace(trimmed[hashIdx+1:])
		trimmed = strings.TrimSpace(trimmed[:hashIdx])
	}

	// 1. okf://, datamesh://, or odkf:// URI scheme
	if strings.HasPrefix(trimmed, "okf://") || strings.HasPrefix(trimmed, "datamesh://") || strings.HasPrefix(trimmed, "odkf://") {
		clean := trimmed
		switch {
		case strings.HasPrefix(clean, "okf://"):
			clean = strings.TrimPrefix(clean, "okf://")
		case strings.HasPrefix(clean, "datamesh://"):
			clean = strings.TrimPrefix(clean, "datamesh://")
		default:
			clean = strings.TrimPrefix(clean, "odkf://")
		}
		// Strip optional version tag after colon or query (e.g. :v1.0 or ?version=...)
		clean = strings.Split(clean, "?")[0]

		parts := strings.Split(clean, "/")
		if len(parts) >= 3 {
			// Format okf://catalog/dataset/resource or datamesh://catalog/dataset/resource
			cat := NormalizeTriadSegment(parts[len(parts)-3])
			ds := NormalizeTriadSegment(parts[len(parts)-2])
			resRaw := strings.TrimSpace(parts[len(parts)-1])
			if colonIdx := strings.Index(resRaw, ":"); colonIdx > 0 {
				resRaw = resRaw[:colonIdx]
			}
			res := NormalizeTriadSegment(resRaw)
			if ds == "" || res == "" {
				return nil, ErrInvalidTriadFormat
			}
			return &ResourceTriad{
				Catalog:  cat,
				Dataset:  ds,
				Resource: res,
				Fragment: fragment,
			}, nil
		} else if len(parts) == 2 {
			ds := NormalizeTriadSegment(parts[0])
			resRaw := strings.TrimSpace(parts[1])
			if colonIdx := strings.Index(resRaw, ":"); colonIdx > 0 {
				resRaw = resRaw[:colonIdx]
			}
			res := NormalizeTriadSegment(resRaw)
			if ds == "" || res == "" {
				return nil, ErrInvalidTriadFormat
			}
			return &ResourceTriad{
				Dataset:  ds,
				Resource: res,
				Fragment: fragment,
			}, nil
		}
		return nil, ErrInvalidTriadFormat
	}

	// 2. Sovereign DNS format: '<dataset>.<resource>@<catalog>'
	if atIdx := strings.Index(trimmed, "@"); atIdx > 0 && !strings.Contains(trimmed, "://") && !strings.Contains(trimmed, "/") && !strings.Contains(trimmed, ":") {
		left := strings.TrimSpace(trimmed[:atIdx])
		catRaw := strings.TrimSpace(trimmed[atIdx+1:])
		dotParts := strings.SplitN(left, ".", 2)
		if len(dotParts) == 2 {
			ds := NormalizeTriadSegment(dotParts[0])
			res := NormalizeTriadSegment(dotParts[1])
			cat := NormalizeTriadSegment(catRaw)
			if ds != "" && res != "" && cat != "" {
				return &ResourceTriad{
					Catalog:  cat,
					Dataset:  ds,
					Resource: res,
					Fragment: fragment,
				}, nil
			}
		}
		return nil, ErrInvalidTriadFormat
	}

	// 3. HTTP/HTTPS or file URLs (e.g. https://.../datasets/cartera-creditos/creditos.csv)
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") || strings.HasPrefix(trimmed, "file://") {
		cleanURL := strings.Split(trimmed, "?")[0]
		cleanURL = strings.TrimRight(cleanURL, "/")
		parts := strings.Split(cleanURL, "/")
		if len(parts) >= 2 {
			ds := NormalizeTriadSegment(parts[len(parts)-2])
			resRaw := strings.TrimSpace(parts[len(parts)-1])
			res := resRaw
			if dotIdx := strings.LastIndex(resRaw, "."); dotIdx > 0 {
				res = resRaw[:dotIdx]
			}
			resNorm := NormalizeTriadSegment(res)
			if ds != "" && resNorm != "" {
				return &ResourceTriad{
					Dataset:  ds,
					Resource: resNorm,
					Fragment: fragment,
				}, nil
			}
		}
	}

	// 4. Colon-separated format ('cat:ds:res' or 'ds:res')
	if strings.Contains(trimmed, ":") {
		parts := strings.Split(trimmed, ":")
		if len(parts) == 3 {
			cat := NormalizeTriadSegment(parts[0])
			ds := NormalizeTriadSegment(parts[1])
			res := NormalizeTriadSegment(parts[2])
			if ds == "" || res == "" {
				return nil, ErrInvalidTriadFormat
			}
			return &ResourceTriad{
				Catalog:  cat,
				Dataset:  ds,
				Resource: res,
				Fragment: fragment,
			}, nil
		} else if len(parts) == 2 {
			ds := NormalizeTriadSegment(parts[0])
			res := NormalizeTriadSegment(parts[1])
			if ds == "" || res == "" {
				return nil, ErrInvalidTriadFormat
			}
			return &ResourceTriad{
				Dataset:  ds,
				Resource: res,
				Fragment: fragment,
			}, nil
		}
	}

	// 5. Slash-separated format ('cat/ds/res' or 'ds/res')
	if strings.Contains(trimmed, "/") {
		var parts []string
		for _, p := range strings.Split(trimmed, "/") {
			if s := strings.TrimSpace(p); s != "" {
				parts = append(parts, s)
			}
		}
		if len(parts) == 3 {
			ds := NormalizeTriadSegment(parts[1])
			res := NormalizeTriadSegment(parts[2])
			if ds == "" || res == "" {
				return nil, ErrInvalidTriadFormat
			}
			return &ResourceTriad{
				Catalog:  NormalizeTriadSegment(parts[0]),
				Dataset:  ds,
				Resource: res,
				Fragment: fragment,
			}, nil
		} else if len(parts) == 2 {
			ds := NormalizeTriadSegment(parts[0])
			res := NormalizeTriadSegment(parts[1])
			if ds == "" || res == "" {
				return nil, ErrInvalidTriadFormat
			}
			return &ResourceTriad{
				Dataset:  ds,
				Resource: res,
				Fragment: fragment,
			}, nil
		}
	}

	return nil, ErrInvalidTriadFormat
}

// ExtractTriadsFromSQL finds all table references formatted as triads or datamesh URIs inside a SQL statement.
func ExtractTriadsFromSQL(sqlQuery string) []ResourceTriad {
	var triads []ResourceTriad
	seen := make(map[string]bool)

	// 1. Scan FROM and JOIN clauses
	matches := tableRefSQLRegex.FindAllStringSubmatch(sqlQuery, -1)
	for _, m := range matches {
		rawRef := m[1]
		if rawRef == "" {
			rawRef = m[2]
		}
		rawRef = strings.TrimSpace(rawRef)
		if rawRef == "" {
			continue
		}

		if triad, err := ParseTriad(rawRef); err == nil {
			key := triad.String()
			if !seen[key] {
				seen[key] = true
				triads = append(triads, *triad)
			}
		}
	}

	// 2. Scan quoted strings with colons
	quotedMatches := quotedColonRegex.FindAllStringSubmatch(sqlQuery, -1)
	for _, m := range quotedMatches {
		if len(m) > 1 {
			rawRef := strings.TrimSpace(m[1])
			if triad, err := ParseTriad(rawRef); err == nil {
				key := triad.String()
				if !seen[key] {
					seen[key] = true
					triads = append(triads, *triad)
				}
			}
		}
	}

	// 3. Scan quoted strings with slashes
	quotedSlashMatches := quotedSlashRegex.FindAllStringSubmatch(sqlQuery, -1)
	for _, m := range quotedSlashMatches {
		if len(m) > 1 {
			rawRef := strings.TrimSpace(m[1])
			if triad, err := ParseTriad(rawRef); err == nil {
				key := triad.String()
				if !seen[key] {
					seen[key] = true
					triads = append(triads, *triad)
				}
			}
		}
	}

	// 4. Scan quoted Sovereign DNS strings (<dataset>.<resource>@<catalog>)
	quotedAtMatches := quotedAtRegex.FindAllStringSubmatch(sqlQuery, -1)
	for _, m := range quotedAtMatches {
		if len(m) > 1 {
			rawRef := strings.TrimSpace(m[1])
			if triad, err := ParseTriad(rawRef); err == nil {
				key := triad.String()
				if !seen[key] {
					seen[key] = true
					triads = append(triads, *triad)
				}
			}
		}
	}

	return triads
}

// ResolvedResource specifies the concrete physical file, URL, and tabular format of a resolved triad.
type ResolvedResource struct {
	Triad       ResourceTriad `json:"triad"`
	PhysicalURI string        `json:"physical_uri"`
	Format      string        `json:"format"` // "parquet", "csv", "json", "tsv"
}
