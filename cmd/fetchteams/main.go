// Command fetchteams fetches FIFA world rankings and writes normalized team OVR
// ratings into the bundled stickers.sqlite.
//
// Usage:
//
//	go run ./cmd/fetchteams                          # auto-fetch latest rankings
//	go run ./cmd/fetchteams --input rankings.json    # use a local JSON file
//	go run ./cmd/fetchteams --date-id id2025         # specific ranking date
//	go run ./cmd/fetchteams --dry-run                # print results, don't write
package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mbuitragoc/panini-api/internal/pipeline"
	_ "modernc.org/sqlite"
)

// fifaEntry is one row from the FIFA ranking API response.
// Field names cover both the current and historical FIFA API response shapes.
type fifaEntry struct {
	Rank         int     `json:"rankId"`
	PreviousRank int     `json:"previousRankId"`
	CountryCode  string  `json:"countryCode"`
	ShortCode    string  `json:"shortCountryCode"`
	TeamName     string  `json:"teamName"`
	Points       float64 `json:"totalPoints"`
	Confederation string `json:"confederation"`
}

type fifaResponse struct {
	Rankings []fifaEntry `json:"rankings"`
}

func main() {
	dbPath := flag.String("db", defaultDB(), "path to stickers.sqlite")
	inputPath := flag.String("input", "", "local JSON file from FIFA API (skips HTTP fetch)")
	dateID := flag.String("date-id", "", "FIFA ranking dateId (e.g. id2025); uses latest if empty")
	dryRun := flag.Bool("dry-run", false, "print results without writing to SQLite")
	flag.Parse()

	data, err := fetchRankings(*inputPath, *dateID)
	if err != nil {
		log.Fatalf("fetch rankings: %v", err)
	}

	entries, err := parseRankings(data)
	if err != nil {
		log.Fatalf("parse rankings: %v", err)
	}
	fmt.Printf("Parsed %d team rankings\n", len(entries))

	teams := normalize(entries)

	if *dryRun {
		printTeams(teams)
		return
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if err := pipeline.EnsureSchema(db); err != nil {
		log.Fatalf("schema: %v", err)
	}

	if err := writeTeams(db, teams); err != nil {
		log.Fatalf("write teams: %v", err)
	}
	fmt.Printf("Done. team_ratings updated in %s\n", *dbPath)
}

func defaultDB() string {
	return "../panini-ios/panini/panini/Resources/stickers.sqlite"
}

// fetchRankings returns the raw JSON bytes from either a local file or the FIFA API.
func fetchRankings(inputPath, dateID string) ([]byte, error) {
	if inputPath != "" {
		return os.ReadFile(inputPath)
	}

	// Auto-fetch: try to discover the latest dateId then fetch rankings.
	if dateID == "" {
		var err error
		dateID, err = latestDateID()
		if err != nil {
			return nil, fmt.Errorf("discover latest ranking date: %w\n\nTip: download the JSON manually and pass --input <file>", err)
		}
		fmt.Printf("Using ranking date: %s\n", dateID)
	}

	url := fmt.Sprintf("https://www.fifa.com/api/ranking-overview?locale=en&dateId=%s", dateID)
	resp, err := http.Get(url) //nolint:gosec,noctx
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("FIFA API returned %d for %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

// latestDateID attempts to discover the most recent FIFA ranking dateId.
func latestDateID() (string, error) {
	type dateEntry struct {
		ID   string `json:"id"`
		Date string `json:"date"`
	}
	type datesResp struct {
		Rankings []dateEntry `json:"rankingDates"`
	}

	resp, err := http.Get("https://www.fifa.com/api/ranking-overview?locale=en") //nolint:gosec,noctx
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Try to extract a dateId from a dates-listing response.
	var dr datesResp
	if err := json.Unmarshal(body, &dr); err == nil && len(dr.Rankings) > 0 {
		return dr.Rankings[0].ID, nil
	}

	// Fallback: try parsing as rankings directly (the endpoint returned live data).
	var rr fifaResponse
	if err := json.Unmarshal(body, &rr); err == nil && len(rr.Rankings) > 0 {
		return "", nil // signal caller that data is already in body
	}

	return "", fmt.Errorf("could not determine latest dateId from FIFA API response")
}

func parseRankings(data []byte) ([]fifaEntry, error) {
	var resp fifaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w\n\nFirst 200 bytes: %s", err, truncate(string(data), 200))
	}
	if len(resp.Rankings) == 0 {
		return nil, fmt.Errorf("no rankings found in response; check --input file format")
	}
	return resp.Rankings, nil
}

// teamRecord is the normalized output ready for insertion.
type teamRecord struct {
	CountryCode       string
	Name              string
	FIFARank          int
	FIFAPoints        float64
	NormalizedOverall int
	Confederation     string
	RankChange        int
}

func normalize(entries []fifaEntry) []teamRecord {
	if len(entries) == 0 {
		return nil
	}

	// Sort by rank ascending to find min/max points among entries we care about.
	// The top-ranked team (rank 1) gets OVR 99; lowest gets 70.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Rank < entries[j].Rank
	})

	maxPts := entries[0].Points
	minPts := entries[len(entries)-1].Points

	out := make([]teamRecord, 0, len(entries))
	for _, e := range entries {
		code := e.CountryCode
		if code == "" {
			code = e.ShortCode
		}

		var ovr int
		if maxPts == minPts {
			ovr = 85 // all same — shouldn't happen
		} else {
			// Linear interpolation: top = 99, bottom = 70.
			ratio := (e.Points - minPts) / (maxPts - minPts)
			ovr = int(math.Round(70 + ratio*29))
		}

		out = append(out, teamRecord{
			CountryCode:       code,
			Name:              e.TeamName,
			FIFARank:          e.Rank,
			FIFAPoints:        e.Points,
			NormalizedOverall: ovr,
			Confederation:     e.Confederation,
			RankChange:        e.PreviousRank - e.Rank, // positive = moved up
		})
	}
	return out
}

func writeTeams(db *sql.DB, teams []teamRecord) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	now := time.Now().UTC().Format(time.RFC3339)
	for _, t := range teams {
		_, err := tx.Exec(`
			INSERT INTO team_ratings
				(country_code, name, fifa_rank, fifa_points, normalized_overall, confederation, rank_change, last_updated)
			VALUES (?,?,?,?,?,?,?,?)
			ON CONFLICT (country_code) DO UPDATE SET
				name=excluded.name, fifa_rank=excluded.fifa_rank,
				fifa_points=excluded.fifa_points, normalized_overall=excluded.normalized_overall,
				confederation=excluded.confederation, rank_change=excluded.rank_change,
				last_updated=excluded.last_updated`,
			t.CountryCode, t.Name, t.FIFARank, t.FIFAPoints,
			t.NormalizedOverall, t.Confederation, t.RankChange, now,
		)
		if err != nil {
			return fmt.Errorf("upsert team %s: %w", t.CountryCode, err)
		}
	}

	return tx.Commit()
}

func printTeams(teams []teamRecord) {
	fmt.Printf("\n%-4s %-30s %4s %8s %3s %4s %s\n", "Code", "Name", "Rank", "Points", "OVR", "Δ", "Confederation")
	fmt.Println(strings.Repeat("─", 72))
	for _, t := range teams {
		delta := ""
		if t.RankChange > 0 {
			delta = fmt.Sprintf("↑%d", t.RankChange)
		} else if t.RankChange < 0 {
			delta = fmt.Sprintf("↓%d", -t.RankChange)
		}
		fmt.Printf("%-4s %-30s %4d %8.1f %3d %4s %s\n",
			t.CountryCode, t.Name, t.FIFARank, t.FIFAPoints,
			t.NormalizedOverall, delta, t.Confederation)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
