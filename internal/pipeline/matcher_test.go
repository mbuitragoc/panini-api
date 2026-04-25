package pipeline

import "testing"

func TestNormalizeName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Kylian Mbappé", "kylian mbappe"},
		{"N'Golo Kanté", "ngolo kante"},
		{"Pedri", "pedri"},
		{"James Rodríguez", "james rodriguez"},
		{"Dávinson Sánchez", "davinson sanchez"},
		{"Aurélien Tchouaméni", "aurelien tchouameni"},
	}
	for _, c := range cases {
		if got := NormalizeName(c.in); got != c.want {
			t.Errorf("NormalizeName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeNation(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Colombia", "colombia"},
		{"South Korea", "korea republic"},
		{"USA", "united states"},
		{"France", "france"},
		{"Ivory Coast", "cote divoire"},
		{"Iran", "ir iran"},
		{"Czech Republic", "czechia"},
	}
	for _, c := range cases {
		if got := NormalizeNation(c.in); got != c.want {
			t.Errorf("NormalizeNation(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestMatch_ExactPass(t *testing.T) {
	// short_name matches sticker exactly (single-name players like Pedri).
	players := []FC26Player{
		{ID: "p1", Name: "Pedri", LongName: "Pedro González López", Nation: "Spain", Overall: 87},
	}
	stickers := []StickerRecord{
		{ID: "ESP-12", PlayerName: "Pedri", Nation: "Spain"},
	}
	results := Match(stickers, players, nil)
	if results[0].Confidence != "exact" {
		t.Errorf("sticker %s: got %q, want exact", results[0].StickerID, results[0].Confidence)
	}
}

func TestMatch_ExactPassViaFirstLastOfLongName(t *testing.T) {
	// FC26 short_name is abbreviated ("J. Bellingham"); sticker has "Jude Bellingham".
	// The first+last-word key of long_name ("jude bellingham") resolves it as exact.
	players := []FC26Player{
		{ID: "p1", Name: "J. Bellingham", LongName: "Jude Victor William Bellingham", Nation: "England", Overall: 90},
	}
	stickers := []StickerRecord{
		{ID: "ENG-10", PlayerName: "Jude Bellingham", Nation: "England"},
	}
	results := Match(stickers, players, nil)
	if results[0].Confidence != "exact" {
		t.Errorf("expected exact via first+last of long_name, got %s", results[0].Confidence)
	}
}

func TestMatch_ExactPassViaFirstTwoWordsOfLongName(t *testing.T) {
	// FC26 long_name "Kylian Mbappé Lottin"; sticker "Kylian Mbappé".
	// First-two-words key ("kylian mbappe") resolves it as exact.
	players := []FC26Player{
		{ID: "p1", Name: "K. Mbappé", LongName: "Kylian Mbappé Lottin", Nation: "France", Overall: 91},
	}
	stickers := []StickerRecord{
		{ID: "FRA-15", PlayerName: "Kylian Mbappé", Nation: "France"},
	}
	results := Match(stickers, players, nil)
	if results[0].Confidence != "exact" {
		t.Errorf("expected exact via first two words of long_name, got %s", results[0].Confidence)
	}
}

func TestMatch_ExactPassStripsAccent(t *testing.T) {
	// Sticker has accented "James Rodríguez"; FC26 long_name matches after normalization.
	players := []FC26Player{
		{ID: "p1", Name: "James Rodríguez", LongName: "James David Rodríguez Rubio", Nation: "Colombia", Overall: 82},
	}
	stickers := []StickerRecord{
		{ID: "COL-11", PlayerName: "James Rodríguez", Nation: "Colombia"},
	}
	results := Match(stickers, players, nil)
	if results[0].Confidence != "exact" {
		t.Errorf("expected exact (accent stripped), got %s", results[0].Confidence)
	}
}

func TestMatch_FuzzyLastName(t *testing.T) {
	// Sticker uses short name ("Kolo Muani"); FC26 long_name is "Randal Kolo Muani".
	// The penultimate+last key "kolo muani" is now indexed as exact.
	players := []FC26Player{
		{ID: "p1", Name: "R. Kolo Muani", LongName: "Randal Kolo Muani", Nation: "France", Overall: 82},
	}
	stickers := []StickerRecord{
		{ID: "FRA-19", PlayerName: "Kolo Muani", Nation: "France"},
	}
	results := Match(stickers, players, nil)
	if results[0].Confidence != "exact" {
		t.Errorf("expected exact (via penultimate+last key), got %s", results[0].Confidence)
	}
	if results[0].PlayerIndex != 0 {
		t.Errorf("expected player index 0, got %d", results[0].PlayerIndex)
	}
}

func TestMatch_FuzzyAmbiguousSkipsToUnmatched(t *testing.T) {
	// Two players with same last name from same nation → no match.
	players := []FC26Player{
		{ID: "p1", Name: "Carlos Garcia", Nation: "Spain", Overall: 70},
		{ID: "p2", Name: "Luis Garcia", Nation: "Spain", Overall: 72},
	}
	stickers := []StickerRecord{
		{ID: "ESP-99", PlayerName: "Garcia", Nation: "Spain"},
	}
	results := Match(stickers, players, nil)
	if results[0].Confidence != "unmatched" {
		t.Errorf("expected unmatched for ambiguous last name, got %s", results[0].Confidence)
	}
}

func TestMatch_Unmatched(t *testing.T) {
	players := []FC26Player{
		{ID: "p1", Name: "Lionel Messi", Nation: "Argentina", Overall: 90},
	}
	stickers := []StickerRecord{
		{ID: "COL-11", PlayerName: "James Rodriguez", Nation: "Colombia"},
	}
	results := Match(stickers, players, nil)
	if results[0].Confidence != "unmatched" {
		t.Errorf("expected unmatched, got %s", results[0].Confidence)
	}
	if results[0].PlayerIndex != -1 {
		t.Errorf("expected PlayerIndex -1, got %d", results[0].PlayerIndex)
	}
}

func TestMatch_ManualOverride(t *testing.T) {
	players := []FC26Player{
		{ID: "p99", Name: "Cucho Hernandez", Nation: "Colombia", Overall: 77},
	}
	stickers := []StickerRecord{
		{ID: "COL-19", PlayerName: "Cucho Hernández", Nation: "Colombia"},
	}
	overrides := map[string]int{"COL-19": 0}
	results := Match(stickers, players, overrides)
	if results[0].Confidence != "manual" {
		t.Errorf("expected manual, got %s", results[0].Confidence)
	}
	if results[0].PlayerIndex != 0 {
		t.Errorf("expected player index 0, got %d", results[0].PlayerIndex)
	}
}

func TestComputeRarity(t *testing.T) {
	cases := []struct {
		overall int
		want    string
	}{
		{91, "legendary"},
		{88, "legendary"},
		{87, "gold"},
		{83, "gold"},
		{82, "silver"},
		{75, "silver"},
		{74, "bronze"},
		{60, "bronze"},
	}
	for _, c := range cases {
		if got := ComputeRarity(c.overall); got != c.want {
			t.Errorf("ComputeRarity(%d) = %q, want %q", c.overall, got, c.want)
		}
	}
}
