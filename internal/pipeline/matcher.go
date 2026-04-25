package pipeline

import "strings"

// FC26Player is a record parsed from the FC 26 dataset CSV.
type FC26Player struct {
	ID            string
	Name          string // short_name: may be abbreviated ("J. Bellingham")
	LongName      string // long_name: full legal name ("Jude Victor William Bellingham")
	Nation        string
	Club          string
	Position      string // primary club position
	Overall       int
	Potential     int
	Age           int
	HeightCM      int
	WeightKG      int
	NationPosition     string // position played for the national team (may differ from club)
	NationJerseyNumber int    // jersey number worn for the national team
	PreferredFoot string // "Right" or "Left"
	WeakFoot      int    // 1–5 stars
	SkillMoves    int    // 1–5 stars
	Pace          int
	Shooting      int
	Passing       int
	Dribbling     int
	Defending     int
	Physical      int
	GKDiving      int
	GKHandling    int
	GKKicking     int
	GKReflexes    int
	GKSpeed       int
	GKPositioning int
	PlayStyles    []string
	DOB           string // ISO date string "YYYY-MM-DD"
}

// StickerRecord is a player sticker row from the album's SQLite database.
type StickerRecord struct {
	ID         string
	PlayerName string
	Nation     string // from national_team column
	Position   string
}

// MatchResult is the output of the matching algorithm for one sticker.
type MatchResult struct {
	StickerID   string
	PlayerIndex int    // index into players slice; -1 if unmatched
	Confidence  string // "exact", "fuzzy", "manual", "unmatched"
}

// Match runs the 3-pass algorithm over all sticker records:
//  1. Exact normalized (name, nation) — singleton matches only
//  2. Last-name + nation — singleton matches only
//  3. Residue → "unmatched"
//
// overrides maps stickerID → player slice index for manually resolved cases.
// Overridden stickers skip all three passes.
func Match(stickers []StickerRecord, players []FC26Player, overrides map[string]int) []MatchResult {
	results := make([]MatchResult, 0, len(stickers))

	// exact index: one or more name keys per player → []playerIndex.
	// Multiple keys per player handle the mismatch between FC26 name formats and
	// Panini sticker names:
	//   short_name "J. Bellingham" vs sticker "Jude Bellingham"
	//   long_name  "Kylian Mbappé Lottin" vs sticker "Kylian Mbappé"
	exactIdx := make(map[string][]int, len(players)*3)
	for i, p := range players {
		nation := NormalizeNation(p.Nation)
		seenKeys := make(map[string]bool, 4) // deduplicate keys within the same player
		addKey := func(name string) {
			if name == "" {
				return
			}
			k := name + "|" + nation
			if seenKeys[k] {
				return
			}
			seenKeys[k] = true
			exactIdx[k] = append(exactIdx[k], i)
		}

		normShort := NormalizeName(p.Name)
		normLong := NormalizeName(p.LongName)
		addKey(normShort) // "j bellingham|england"
		addKey(normLong)  // "jude victor william bellingham|england"

		// Additional keys derived from long_name to bridge the gap between FC26 name formats
		// and Panini sticker names (abbreviated short names, hyphenated surnames, etc.).
		longParts := strings.Fields(normLong)
		n := len(longParts)
		if n >= 3 {
			addKey(longParts[0] + " " + longParts[n-1])   // first+last  "jude bellingham"
			addKey(longParts[0] + " " + longParts[1])     // first+two   "kylian mbappe"
			addKey(longParts[n-2] + " " + longParts[n-1]) // penult+last "ousmane dembele"
		}
		if n >= 4 {
			addKey(longParts[0] + " " + longParts[2])                            // first+third "luis diaz", "alvaro morata"
			addKey(longParts[0] + " " + longParts[1] + " " + longParts[2])       // first-three "miguel angel borja"
		}
	}

	// last-name index: "lastName|normNation" → []playerIndex (uses long_name last word)
	lastNameIdx := make(map[string][]int, len(players))
	for i, p := range players {
		nation := NormalizeNation(p.Nation)
		parts := strings.Fields(NormalizeName(p.LongName))
		if len(parts) > 0 {
			key := parts[len(parts)-1] + "|" + nation
			lastNameIdx[key] = append(lastNameIdx[key], i)
		}
	}

	for _, s := range stickers {
		// Manual override: skip algorithm entirely.
		if idx, ok := overrides[s.ID]; ok {
			results = append(results, MatchResult{StickerID: s.ID, PlayerIndex: idx, Confidence: "manual"})
			continue
		}

		normName := NormalizeName(s.PlayerName)
		normNation := NormalizeNation(s.Nation)

		// Pass 1: exact normalized match, singleton only.
		if hits := exactIdx[normName+"|"+normNation]; len(hits) == 1 {
			results = append(results, MatchResult{StickerID: s.ID, PlayerIndex: hits[0], Confidence: "exact"})
			continue
		}

		// Pass 2: sticker's last word vs long_name's last word, singleton only.
		if parts := strings.Fields(normName); len(parts) > 0 {
			lastName := parts[len(parts)-1]
			if hits := lastNameIdx[lastName+"|"+normNation]; len(hits) == 1 {
				results = append(results, MatchResult{StickerID: s.ID, PlayerIndex: hits[0], Confidence: "fuzzy"})
				continue
			}
		}

		results = append(results, MatchResult{StickerID: s.ID, PlayerIndex: -1, Confidence: "unmatched"})
	}

	return results
}

// ComputeRarity maps an overall rating to a rarity tier.
// Thresholds: legendary ≥89, gold 82–88, silver 75–81, bronze <75.
func ComputeRarity(overall int) string {
	switch {
	case overall >= 89:
		return "legendary"
	case overall >= 82:
		return "gold"
	case overall >= 75:
		return "silver"
	default:
		return "bronze"
	}
}
