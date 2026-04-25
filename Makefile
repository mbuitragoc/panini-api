.PHONY: build run test migrate tidy lint import-players review-players fetch-teams

DB ?= ../panini-ios/panini/panini/Resources/stickers.sqlite

build:
	go build ./cmd/api/

run:
	go run ./cmd/api/

test:
	go test ./...

migrate:
	psql $(DATABASE_URL) -f migrations/001_initial_schema.sql

tidy:
	go mod tidy

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run || echo "golangci-lint not found, skipping"

# ── Data pipeline ────────────────────────────────────────────────────────────

# Import FC 26 player data and match to stickers.
# Usage: make import-players CSV=/path/to/fc26_players.csv
import-players:
	@test -n "$(CSV)" || (echo "usage: make import-players CSV=/path/to/fc26_players.csv"; exit 1)
	go run ./cmd/importplayers --csv=$(CSV) --db=$(DB)

# Same as import-players but opens interactive review for unmatched stickers.
# Usage: make review-players CSV=/path/to/fc26_players.csv
review-players:
	@test -n "$(CSV)" || (echo "usage: make review-players CSV=/path/to/fc26_players.csv"; exit 1)
	go run ./cmd/importplayers --csv=$(CSV) --db=$(DB) --review

# Fetch FIFA world rankings and write team_ratings to SQLite.
# Pass INPUT= to use a local JSON file instead of the live FIFA API.
# Usage: make fetch-teams
#        make fetch-teams INPUT=/path/to/rankings.json
fetch-teams:
	go run ./cmd/fetchteams --db=$(DB) $(if $(INPUT),--input=$(INPUT),) $(if $(DATE_ID),--date-id=$(DATE_ID),)
