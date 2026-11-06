package domain

import (
	"strings"
)

// CatalogEntry represents a reference to a federated OKF Data Product in llms.txt.
type CatalogEntry struct {
	Title         string   `json:"title"`
	URI           string   `json:"uri"`
	ResolvedURL   string   `json:"resolved_url"`
	Description   string   `json:"description"`
	Domain        string   `json:"domain"`
	Resources     []string `json:"resources"`
	CatalogSource string   `json:"catalog_source,omitempty"`
}

// Catalog represents a collection of Data Products discovered from one or more sovereign catalogs.
type Catalog struct {
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	SourceURL      string         `json:"source_url"`
	SourceCatalogs []string       `json:"source_catalogs,omitempty"`
	Entries        []CatalogEntry `json:"entries"`
}

var searchDiacriticsReplacer = strings.NewReplacer(
	"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u",
	"à", "a", "è", "e", "ì", "i", "ò", "o", "ù", "u",
	"ä", "a", "ë", "e", "ï", "i", "ö", "o", "ü", "u",
	"â", "a", "ê", "e", "î", "i", "ô", "o", "û", "u",
	"ñ", "n", "ç", "c",
)

func normalizeSearchText(s string) string {
	return searchDiacriticsReplacer.Replace(strings.ToLower(strings.TrimSpace(s)))
}

// FindByKeyword searches catalog entries matching all tokens (accent-insensitive) in title, description, domain, URI, or resources.
func (c *Catalog) FindByKeyword(keyword string) []CatalogEntry {
	if strings.TrimSpace(keyword) == "" {
		return c.Entries
	}
	tokens := strings.Fields(normalizeSearchText(keyword))
	if len(tokens) == 0 {
		return c.Entries
	}
	var matches []CatalogEntry

	for _, entry := range c.Entries {
		haystack := normalizeSearchText(strings.Join([]string{
			entry.Title,
			entry.Description,
			entry.Domain,
			entry.URI,
			entry.ResolvedURL,
			strings.Join(entry.Resources, " "),
		}, " "))
		allMatched := true
		for _, tok := range tokens {
			if !strings.Contains(haystack, tok) {
				allMatched = false
				break
			}
		}
		if allMatched {
			matches = append(matches, entry)
		}
	}
	return matches
}

// FindByDomain returns catalog entries belonging to an exact or partial domain category.
func (c *Catalog) FindByDomain(domainCategory string) []CatalogEntry {
	dc := normalizeSearchText(domainCategory)
	var matches []CatalogEntry

	for _, entry := range c.Entries {
		if strings.Contains(normalizeSearchText(entry.Domain), dc) {
			matches = append(matches, entry)
		}
	}
	return matches
}
