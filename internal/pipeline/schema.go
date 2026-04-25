package pipeline

import (
	"database/sql"
	"fmt"
	"strings"
)

// EnsureSchema creates all pipeline tables in the SQLite database if they don't exist.
// Safe to call on every run — all statements use CREATE TABLE IF NOT EXISTS.
func EnsureSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS player_ratings (
			id                   INTEGER PRIMARY KEY AUTOINCREMENT,
			fc_player_id         TEXT,
			name                 TEXT NOT NULL,
			normalized_name      TEXT NOT NULL,
			nation               TEXT NOT NULL,
			club                 TEXT,
			position             TEXT,
			overall              INTEGER NOT NULL,
			potential            INTEGER,
			age                  INTEGER,
			dob                  TEXT,
			height_cm            INTEGER,
			weight_kg            INTEGER,
			nation_position      TEXT,
			nation_jersey_number INTEGER,
			preferred_foot       TEXT,
			weak_foot            INTEGER,
			skill_moves          INTEGER,
			pace                 INTEGER,
			shooting             INTEGER,
			passing              INTEGER,
			dribbling            INTEGER,
			defending            INTEGER,
			physical             INTEGER,
			gk_diving            INTEGER,
			gk_handling          INTEGER,
			gk_kicking           INTEGER,
			gk_reflexes          INTEGER,
			gk_speed             INTEGER,
			gk_positioning       INTEGER,
			play_styles          TEXT,
			rarity               TEXT,
			UNIQUE (normalized_name, nation)
		)`,
		`CREATE TABLE IF NOT EXISTS team_ratings (
			country_code        TEXT PRIMARY KEY,
			name                TEXT NOT NULL,
			fifa_rank           INTEGER,
			fifa_points         REAL,
			normalized_overall  INTEGER,
			confederation       TEXT,
			rank_change         INTEGER,
			last_updated        TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS sticker_player_link (
			sticker_id       TEXT PRIMARY KEY,
			player_rating_id INTEGER REFERENCES player_ratings(id),
			confidence       TEXT,
			matched_at       TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS country_context (
			country_code    TEXT PRIMARY KEY,
			narrative_en    TEXT,
			narrative_es    TEXT,
			fun_fact_en     TEXT,
			fun_fact_es     TEXT,
			backdrop_config TEXT,
			updated_at      TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS manual_overrides (
			sticker_id   TEXT PRIMARY KEY,
			fc_player_id TEXT NOT NULL,
			note         TEXT,
			created_at   TEXT
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("create schema: %w", err)
		}
	}
	return nil
}

// MigrateStickersTable adds dob, height_cm, and weight_kg columns to the stickers
// table if they don't already exist. The pipeline writes these back after matching
// so the iOS StickerStore picks them up automatically during seeding.
func MigrateStickersTable(db *sql.DB) error {
	existing, err := tableColumns(db, "stickers")
	if err != nil {
		return fmt.Errorf("read stickers columns: %w", err)
	}
	additions := []struct{ name, typ string }{
		{"dob", "TEXT"},
		{"height_cm", "REAL"},
		{"weight_kg", "REAL"},
	}
	for _, col := range additions {
		if !existing[col.name] {
			if _, err := db.Exec(fmt.Sprintf("ALTER TABLE stickers ADD COLUMN %s %s", col.name, col.typ)); err != nil {
				return fmt.Errorf("alter stickers add %s: %w", col.name, err)
			}
		}
	}
	return nil
}

func tableColumns(db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var cid, notnull, pk int
		var name, typ string
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols[strings.ToLower(name)] = true
	}
	return cols, rows.Err()
}
