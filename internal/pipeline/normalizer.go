package pipeline

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// NormalizeName strips accents, lowercases, and removes non-letter characters.
// Hyphens are converted to spaces so "Thuram-Ulien" → "thuram ulien" (not "thuramulien").
// "N'Golo Kanté" → "ngolo kante", "James Rodríguez" → "james rodriguez"
func NormalizeName(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	result = strings.ToLower(result)

	var b strings.Builder
	for _, r := range result {
		if unicode.IsLetter(r) || r == ' ' {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// nationTranslations maps EA Sports nation names (normalized) to FIFA nation names (normalized).
// ~15 entries covering the most common World Cup mismatches.
var nationTranslations = map[string]string{
	"usa":                          "united states",
	"south korea":                  "korea republic",
	"ivory coast":                  "cote divoire",
	"czech republic":               "czechia",
	"iran":                         "ir iran",
	"china":                        "china pr",
	"democratic republic of congo": "dr congo",
	"drc":                          "dr congo",
	"bosnia  herzegovina":          "bosnia and herzegovina",
	"kyrgyz republic":              "kyrgyzstan",
	"trinidad  tobago":             "trinidad and tobago",
	"cape verde":                   "cabo verde",
	"syria":                        "syrian arab republic",
	"hong kong":                    "china hong kong",
	"north korea":                  "korea dpr",
}

// NormalizeNation normalizes a nation name for comparison, applying EA Sports → FIFA translations.
func NormalizeNation(s string) string {
	n := NormalizeName(s)
	if mapped, ok := nationTranslations[n]; ok {
		return mapped
	}
	return n
}
