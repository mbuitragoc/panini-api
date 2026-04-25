package pipeline

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ParseFC26CSV reads a FC 26 player CSV and returns all parsed players.
// Targets column names from the Kaggle rovnez/fc-26-fifa-26-player-data dataset.
func ParseFC26CSV(path string) ([]FC26Player, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()
	return parseFC26Reader(f)
}

func parseFC26Reader(r io.Reader) ([]FC26Player, error) {
	rd := csv.NewReader(r)
	rd.LazyQuotes = true
	rd.TrimLeadingSpace = true

	header, err := rd.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	col := buildColIndex(header)

	required := []string{"player_id", "short_name", "long_name", "nationality_name", "overall"}
	for _, name := range required {
		if _, ok := col[name]; !ok {
			return nil, fmt.Errorf("required column %q not found; got: %v", name, header)
		}
	}

	var players []FC26Player
	for {
		row, err := rd.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}

		shortName := get(row, col, "short_name")
		overall := parseInt(get(row, col, "overall"))
		if shortName == "" || overall == 0 {
			continue
		}

		players = append(players, FC26Player{
			ID:                 get(row, col, "player_id"),
			Name:               shortName,
			LongName:           get(row, col, "long_name"),
			Nation:             get(row, col, "nationality_name"),
			Club:               get(row, col, "club_name"),
			Position:           primaryPosition(get(row, col, "player_positions")),
			Overall:            overall,
			Potential:          parseInt(get(row, col, "potential")),
			Age:                parseInt(get(row, col, "age")),
			DOB:                get(row, col, "dob"),
			HeightCM:           parseInt(get(row, col, "height_cm")),
			WeightKG:           parseInt(get(row, col, "weight_kg")),
			NationPosition:     get(row, col, "nation_position"),
			NationJerseyNumber: parseInt(get(row, col, "nation_jersey_number")),
			PreferredFoot:      get(row, col, "preferred_foot"),
			WeakFoot:           parseInt(get(row, col, "weak_foot")),
			SkillMoves:         parseInt(get(row, col, "skill_moves")),
			Pace:               parseInt(get(row, col, "pace")),
			Shooting:           parseInt(get(row, col, "shooting")),
			Passing:            parseInt(get(row, col, "passing")),
			Dribbling:          parseInt(get(row, col, "dribbling")),
			Defending:          parseInt(get(row, col, "defending")),
			Physical:           parseInt(get(row, col, "physic")),
			GKDiving:           parseInt(get(row, col, "goalkeeping_diving")),
			GKHandling:         parseInt(get(row, col, "goalkeeping_handling")),
			GKKicking:          parseInt(get(row, col, "goalkeeping_kicking")),
			GKReflexes:         parseInt(get(row, col, "goalkeeping_reflexes")),
			GKSpeed:            parseInt(get(row, col, "goalkeeping_speed")),
			GKPositioning:      parseInt(get(row, col, "goalkeeping_positioning")),
			// player_traits holds the actual play style mechanics ("Finesse Shot", "Quick Step", etc.)
			// player_tags holds descriptive labels (#Playmaker, #Dribbler) — we don't use those.
			PlayStyles: parsePlayStyles(get(row, col, "player_traits")),
		})
	}
	return players, nil
}

func buildColIndex(header []string) map[string]int {
	m := make(map[string]int, len(header))
	for i, h := range header {
		m[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return m
}

func get(row []string, col map[string]int, name string) string {
	i, ok := col[name]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}

func primaryPosition(s string) string {
	if s == "" {
		return ""
	}
	return strings.TrimSpace(strings.SplitN(s, ",", 2)[0])
}

// parsePlayStyles parses player_traits values, e.g.:
// "Relentless +, Low Driven Shot, Gamechanger, Inventive"
// Values are comma-separated and may contain \xa0 (non-breaking space).
func parsePlayStyles(s string) []string {
	if s == "" {
		return nil
	}
	// Normalize non-breaking spaces to regular spaces.
	s = strings.ReplaceAll(s, "\xa0", " ")
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
