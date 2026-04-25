// Command importplayers imports FC 26 player data into the bundled stickers.sqlite.
// It runs a 3-pass name-matching algorithm against existing player stickers and
// populates the player_ratings and sticker_player_link tables.
//
// Usage:
//
//	go run ./cmd/importplayers --csv fc26_players.csv
//	go run ./cmd/importplayers --csv fc26_players.csv --review   # interactive review for unmatched
package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mbuitragoc/panini-api/internal/pipeline"
	_ "modernc.org/sqlite"
)

func main() {
	csvPath := flag.String("csv", "", "path to FC 26 player CSV file (required)")
	dbPath := flag.String("db", defaultDB(), "path to stickers.sqlite")
	review := flag.Bool("review", false, "interactive manual review for unmatched stickers")
	flag.Parse()

	if *csvPath == "" {
		fmt.Fprintln(os.Stderr, "usage: importplayers --csv <fc26.csv> [--db <path>] [--review]")
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1) // SQLite single-writer

	if err := pipeline.EnsureSchema(db); err != nil {
		log.Fatalf("schema: %v", err)
	}
	if err := pipeline.MigrateStickersTable(db); err != nil {
		log.Fatalf("migrate stickers: %v", err)
	}

	stickers, err := loadPlayerStickers(db)
	if err != nil {
		log.Fatalf("load stickers: %v", err)
	}
	fmt.Printf("Loaded %d player stickers from SQLite\n", len(stickers))

	players, err := pipeline.ParseFC26CSV(*csvPath)
	if err != nil {
		log.Fatalf("parse csv: %v", err)
	}
	fmt.Printf("Loaded %d players from FC26 CSV\n", len(players))

	overrides, err := loadOverrides(db, players)
	if err != nil {
		log.Fatalf("load overrides: %v", err)
	}
	if len(overrides) > 0 {
		fmt.Printf("Loaded %d manual overrides\n", len(overrides))
	}

	results := pipeline.Match(stickers, players, overrides)
	printStats(results)

	if *review {
		results = runManualReview(db, stickers, players, results)
		fmt.Println()
		fmt.Println("After review:")
		printStats(results)
	}

	if err := upsertResults(db, stickers, players, results); err != nil {
		log.Fatalf("upsert: %v", err)
	}
	if err := backfillStickers(db, stickers, players, results); err != nil {
		log.Fatalf("backfill stickers: %v", err)
	}

	fmt.Printf("\nDone. Written to %s\n", *dbPath)
}

// defaultDB returns the path to the bundled SQLite, assuming the tool is run
// from the panini-api/ directory.
func defaultDB() string {
	return "../panini-ios/panini/panini/Resources/stickers.sqlite"
}

func loadPlayerStickers(db *sql.DB) ([]pipeline.StickerRecord, error) {
	rows, err := db.Query(
		`SELECT id, player_name, national_team, COALESCE(position, '') FROM stickers WHERE type = 'player' AND player_name IS NOT NULL`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []pipeline.StickerRecord
	for rows.Next() {
		var s pipeline.StickerRecord
		if err := rows.Scan(&s.ID, &s.PlayerName, &s.Nation, &s.Position); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// loadOverrides reads manual_overrides and resolves them to player slice indices.
func loadOverrides(db *sql.DB, players []pipeline.FC26Player) (map[string]int, error) {
	rows, err := db.Query(`SELECT sticker_id, fc_player_id FROM manual_overrides`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	idxByFCID := make(map[string]int, len(players))
	for i, p := range players {
		idxByFCID[p.ID] = i
	}

	overrides := map[string]int{}
	for rows.Next() {
		var stickerID, fcPlayerID string
		if err := rows.Scan(&stickerID, &fcPlayerID); err != nil {
			return nil, err
		}
		if idx, ok := idxByFCID[fcPlayerID]; ok {
			overrides[stickerID] = idx
		}
	}
	return overrides, rows.Err()
}

func printStats(results []pipeline.MatchResult) {
	var exact, fuzzy, manual, unmatched int
	for _, r := range results {
		switch r.Confidence {
		case "exact":
			exact++
		case "fuzzy":
			fuzzy++
		case "manual":
			manual++
		default:
			unmatched++
		}
	}
	total := exact + fuzzy + manual + unmatched
	fmt.Printf("\nMatch results (%d stickers):\n", total)
	fmt.Printf("  exact:     %d (%.0f%%)\n", exact, pct(exact, total))
	fmt.Printf("  fuzzy:     %d (%.0f%%)\n", fuzzy, pct(fuzzy, total))
	fmt.Printf("  manual:    %d (%.0f%%)\n", manual, pct(manual, total))
	fmt.Printf("  unmatched: %d (%.0f%%)\n", unmatched, pct(unmatched, total))
}

func pct(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}

func runManualReview(db *sql.DB, stickers []pipeline.StickerRecord, players []pipeline.FC26Player, results []pipeline.MatchResult) []pipeline.MatchResult {
	stickerByID := make(map[string]pipeline.StickerRecord, len(stickers))
	for _, s := range stickers {
		stickerByID[s.ID] = s
	}

	scanner := bufio.NewScanner(os.Stdin)

	// Count unmatched first
	unmatched := 0
	for _, r := range results {
		if r.Confidence == "unmatched" {
			unmatched++
		}
	}
	fmt.Printf("\n%d unmatched stickers to review. Search by name fragment. Enter to skip, 'q' to quit.\n", unmatched)

	for i, r := range results {
		if r.Confidence != "unmatched" {
			continue
		}
		s := stickerByID[r.StickerID]
		fmt.Printf("\n─── %s | %s | %s\n", r.StickerID, s.PlayerName, s.Nation)
		fmt.Print("Search: ")

		if !scanner.Scan() {
			break
		}
		query := strings.TrimSpace(scanner.Text())
		if strings.EqualFold(query, "q") {
			break
		}
		if query == "" {
			continue
		}

		hits := searchPlayers(players, query)
		if len(hits) == 0 {
			fmt.Println("  No results.")
			continue
		}
		if len(hits) > 8 {
			hits = hits[:8]
		}
		for n, idx := range hits {
			p := players[idx]
			fmt.Printf("  %d. %s (%s) OVR %d %s\n", n+1, p.Name, p.Nation, p.Overall, p.Position)
		}

		fmt.Print("Select (number, or Enter to skip): ")
		if !scanner.Scan() {
			break
		}
		choice := strings.TrimSpace(scanner.Text())
		if choice == "" {
			continue
		}
		var n int
		if _, err := fmt.Sscan(choice, &n); err != nil || n < 1 || n > len(hits) {
			fmt.Println("  Invalid — skipping.")
			continue
		}

		idx := hits[n-1]
		results[i].PlayerIndex = idx
		results[i].Confidence = "manual"

		_, err := db.Exec(
			`INSERT OR REPLACE INTO manual_overrides (sticker_id, fc_player_id, created_at) VALUES (?, ?, ?)`,
			r.StickerID, players[idx].ID, time.Now().UTC().Format(time.RFC3339),
		)
		if err != nil {
			fmt.Printf("  Warning: could not persist override: %v\n", err)
		} else {
			fmt.Printf("  Saved: %s → %s [manual]\n", r.StickerID, players[idx].Name)
		}
	}

	return results
}

// nullStr returns nil for empty strings so SQLite stores NULL rather than "".
func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// nullInt returns nil for zero values where 0 is not a meaningful value.
func nullInt(n int) interface{} {
	if n == 0 {
		return nil
	}
	return n
}

// backfillStickers writes dob, height_cm, weight_kg from the matched FC26 player
// back into the stickers table so the iOS StickerStore picks them up during seeding.
func backfillStickers(db *sql.DB, stickers []pipeline.StickerRecord, players []pipeline.FC26Player, results []pipeline.MatchResult) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	updated := 0
	for _, r := range results {
		if r.PlayerIndex < 0 {
			continue
		}
		p := players[r.PlayerIndex]
		if p.DOB == "" && p.HeightCM == 0 && p.WeightKG == 0 {
			continue
		}
		_, err := tx.Exec(
			`UPDATE stickers SET dob=?, height_cm=?, weight_kg=? WHERE id=?`,
			nullStr(p.DOB), nullInt(p.HeightCM), nullInt(p.WeightKG), r.StickerID,
		)
		if err != nil {
			return fmt.Errorf("backfill sticker %s: %w", r.StickerID, err)
		}
		updated++
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Printf("Backfilled dob/height/weight for %d stickers\n", updated)
	return nil
}

func searchPlayers(players []pipeline.FC26Player, query string) []int {
	normQuery := pipeline.NormalizeName(query)
	var hits []int
	for i, p := range players {
		if strings.Contains(pipeline.NormalizeName(p.Name), normQuery) {
			hits = append(hits, i)
		}
	}
	return hits
}

func upsertResults(db *sql.DB, stickers []pipeline.StickerRecord, players []pipeline.FC26Player, results []pipeline.MatchResult) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	stickerByID := make(map[string]pipeline.StickerRecord, len(stickers))
	for _, s := range stickers {
		stickerByID[s.ID] = s
	}

	now := time.Now().UTC().Format(time.RFC3339)

	for _, r := range results {
		if r.PlayerIndex < 0 {
			_, err := tx.Exec(
				`INSERT OR REPLACE INTO sticker_player_link (sticker_id, player_rating_id, confidence, matched_at) VALUES (?, NULL, 'unmatched', ?)`,
				r.StickerID, now,
			)
			if err != nil {
				return fmt.Errorf("link unmatched %s: %w", r.StickerID, err)
			}
			continue
		}

		p := players[r.PlayerIndex]
		normName := pipeline.NormalizeName(p.Name)
		normNation := pipeline.NormalizeNation(p.Nation)
		rarity := pipeline.ComputeRarity(p.Overall)
		playStylesJSON, _ := json.Marshal(p.PlayStyles)

		var ratingID int64
		err := tx.QueryRow(`
			INSERT INTO player_ratings (
				fc_player_id, name, normalized_name, nation, club, position,
				overall, potential, age, dob, height_cm, weight_kg,
				nation_position, nation_jersey_number,
				preferred_foot, weak_foot, skill_moves,
				pace, shooting, passing, dribbling, defending, physical,
				gk_diving, gk_handling, gk_kicking, gk_reflexes, gk_speed, gk_positioning,
				play_styles, rarity
			) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT (normalized_name, nation) DO UPDATE SET
				fc_player_id=excluded.fc_player_id, club=excluded.club,
				position=excluded.position, overall=excluded.overall,
				potential=excluded.potential, age=excluded.age,
				dob=excluded.dob, height_cm=excluded.height_cm, weight_kg=excluded.weight_kg,
				nation_position=excluded.nation_position,
				nation_jersey_number=excluded.nation_jersey_number,
				preferred_foot=excluded.preferred_foot,
				weak_foot=excluded.weak_foot, skill_moves=excluded.skill_moves,
				pace=excluded.pace, shooting=excluded.shooting, passing=excluded.passing,
				dribbling=excluded.dribbling, defending=excluded.defending,
				physical=excluded.physical, gk_diving=excluded.gk_diving,
				gk_handling=excluded.gk_handling, gk_kicking=excluded.gk_kicking,
				gk_reflexes=excluded.gk_reflexes, gk_speed=excluded.gk_speed,
				gk_positioning=excluded.gk_positioning,
				play_styles=excluded.play_styles, rarity=excluded.rarity
			RETURNING id`,
			p.ID, p.Name, normName, normNation, p.Club, p.Position,
			p.Overall, p.Potential, p.Age, nullStr(p.DOB), p.HeightCM, p.WeightKG,
			nullStr(p.NationPosition), nullInt(p.NationJerseyNumber),
			nullStr(p.PreferredFoot), p.WeakFoot, p.SkillMoves,
			p.Pace, p.Shooting, p.Passing, p.Dribbling, p.Defending, p.Physical,
			p.GKDiving, p.GKHandling, p.GKKicking, p.GKReflexes, p.GKSpeed, p.GKPositioning,
			string(playStylesJSON), rarity,
		).Scan(&ratingID)
		if err != nil {
			return fmt.Errorf("upsert player_rating %q: %w", p.Name, err)
		}

		_, err = tx.Exec(
			`INSERT OR REPLACE INTO sticker_player_link (sticker_id, player_rating_id, confidence, matched_at) VALUES (?, ?, ?, ?)`,
			r.StickerID, ratingID, r.Confidence, now,
		)
		if err != nil {
			return fmt.Errorf("upsert link %s: %w", r.StickerID, err)
		}
	}

	return tx.Commit()
}
