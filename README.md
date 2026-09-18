# panini-api

Go backend for a FIFA 2026 Panini sticker-collection tracker: you record which stickers you own, see who among your friends has your duplicates, and negotiate a swap in a live trading session.

An iOS client ([`panini-ios`](https://github.com/mbuitragoc/panini-ios)) and a web client consume the same API.

## Stack

| | |
|---|---|
| Router | `go-chi/chi` v5 |
| Database | PostgreSQL via `jackc/pgx` v5, plain SQL migrations |
| Auth | Clerk-issued JWTs, verified with `lestrrat-go/jwx` |
| Runtime | Go 1.25 |

## Layout

```
cmd/
  api/              the HTTP server
  fetchteams/       one-off importer for squads
  importplayers/    one-off importer for player ratings
internal/
  app/              composition root: wiring, not logic
  http/             router, middleware, route table
  auth/             JWT verification and the authenticated-user context
  sessions/         live trading sessions, including the SSE event stream
  trades/           trade offers and their state transitions
  collections/      what a user owns
  stickers/         the sticker catalogue
  friends/          friend requests and visibility into a friend's collection
  countries/        national squads and crests
  sync/             one call that returns everything a cold client needs
  notify/           device tokens and push delivery
  pipeline/         the ratings ingestion pipeline
  admin/            operational endpoints
migrations/         001..008, applied in order
```

Each domain package owns its own handlers, queries and types. `internal/app` is the only place that knows how they fit together, so a domain can be read end to end without leaving its folder.

## Design notes

**Sessions are the interesting part.** A trade is not a single request. Two collectors join a session by code, each publishes what they hold and what they will part with, and both sides watch the other's changes arrive over `GET /v1/sessions/{code}/events` as Server-Sent Events until someone confirms. The endpoints (`join`, `stickers`, `offered-stickers`, `singletons`, `confirm`, `leave`) are the state machine for that negotiation.

**Auth is a router concern, not a handler concern.** Public reads (the catalogue, a session by code) sit outside the authenticated `chi.Group`; everything touching a user's own data sits inside it. A handler never asks whether the caller is signed in, because it cannot be reached unless they are.

**`/v1/sync` exists so a cold client makes one request.** A mobile app opening after a week needs the catalogue, the user's collection, friends and open trades. Four round trips on a bad connection is four chances to fail, so the server composes the payload instead.

## Running it

```bash
cp .env.example .env          # DATABASE_URL, Clerk keys
psql "$DATABASE_URL" -f migrations/001_initial_schema.sql   # ...through 008
go run ./cmd/api
```

`GET /health` answers once it is up.

## Status

A personal project, actively built. Test coverage is currently thin and concentrated in the trade-matching logic, which is where the rules are subtle enough to be worth pinning down.
