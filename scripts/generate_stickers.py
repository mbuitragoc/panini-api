#!/usr/bin/env python3
"""
Generate sticker seed data for development (3-team subset: ESP, FRA, COL).
Writes:
  - panini-api/migrations/002_seed_stickers.sql   (Postgres INSERT statements)
  - panini-ios/panini/panini/Resources/stickers.sqlite

Run from the panini-app/ root:
  python3 panini-api/scripts/generate_stickers.py

When the full dataset is available, replace STICKERS and re-run.
"""

import os
import sqlite3

# ── Source of truth ─────────────────────────────────────────────────────────
# Fields: id, country_code, sticker_number, type, player_name, position,
#         national_team, club, club_country, dob, height_cm, weight_kg
# type: badge | team | player | stadium | mascot | legend | special
# dob/height_cm/weight_kg are NULL here; backfilled by importplayers pipeline.

STICKERS = [
    # ── SPAIN (ESP) ─────────────────────────────────────────────────────────
    ("ESP-1",  "ESP", 1,  "badge",  None,                      None,  "Spain",    None,               None,  None, None, None),
    ("ESP-2",  "ESP", 2,  "team",   None,                      None,  "Spain",    None,               None,  None, None, None),
    ("ESP-3",  "ESP", 3,  "player", "Unai Simón",              "GK",  "Spain",    "Athletic Club",    "Spain",    None, None, None),
    ("ESP-4",  "ESP", 4,  "player", "David Raya",              "GK",  "Spain",    "Arsenal",          "England",  None, None, None),
    ("ESP-5",  "ESP", 5,  "player", "Alex Remiro",             "GK",  "Spain",    "Real Sociedad",    "Spain",    None, None, None),
    ("ESP-6",  "ESP", 6,  "player", "Dani Carvajal",           "DEF", "Spain",    "Real Madrid",      "Spain",    None, None, None),
    ("ESP-7",  "ESP", 7,  "player", "Robin Le Normand",        "DEF", "Spain",    "Atlético Madrid",  "Spain",    None, None, None),
    ("ESP-8",  "ESP", 8,  "player", "Aymeric Laporte",         "DEF", "Spain",    "Al-Nassr",         "Saudi Arabia", None, None, None),
    ("ESP-9",  "ESP", 9,  "player", "Marc Cucurella",          "DEF", "Spain",    "Chelsea",          "England",  None, None, None),
    ("ESP-10", "ESP", 10, "player", "Alejandro Grimaldo",      "DEF", "Spain",    "Bayer Leverkusen", "Germany",  None, None, None),
    ("ESP-11", "ESP", 11, "player", "Rodri",                   "MID", "Spain",    "Manchester City",  "England",  None, None, None),
    ("ESP-12", "ESP", 12, "player", "Pedri",                   "MID", "Spain",    "Barcelona",        "Spain",    None, None, None),
    ("ESP-13", "ESP", 13, "player", "Fabián Ruiz",             "MID", "Spain",    "PSG",              "France",   None, None, None),
    ("ESP-14", "ESP", 14, "player", "Mikel Merino",            "MID", "Spain",    "Arsenal",          "England",  None, None, None),
    ("ESP-15", "ESP", 15, "player", "Dani Olmo",               "MID", "Spain",    "Barcelona",        "Spain",    None, None, None),
    ("ESP-16", "ESP", 16, "player", "Lamine Yamal",            "FWD", "Spain",    "Barcelona",        "Spain",    None, None, None),
    ("ESP-17", "ESP", 17, "player", "Nico Williams",           "FWD", "Spain",    "Athletic Club",    "Spain",    None, None, None),
    ("ESP-18", "ESP", 18, "player", "Álvaro Morata",           "FWD", "Spain",    "AC Milan",         "Italy",    None, None, None),
    ("ESP-19", "ESP", 19, "player", "Ferran Torres",           "FWD", "Spain",    "Barcelona",        "Spain",    None, None, None),
    ("ESP-20", "ESP", 20, "player", "Mikel Oyarzabal",         "FWD", "Spain",    "Real Sociedad",    "Spain",    None, None, None),

    # ── FRANCE (FRA) ────────────────────────────────────────────────────────
    ("FRA-1",  "FRA", 1,  "badge",  None,                      None,  "France",   None,               None,  None, None, None),
    ("FRA-2",  "FRA", 2,  "team",   None,                      None,  "France",   None,               None,  None, None, None),
    ("FRA-3",  "FRA", 3,  "player", "Mike Maignan",            "GK",  "France",   "AC Milan",         "Italy",    None, None, None),
    ("FRA-4",  "FRA", 4,  "player", "Alphonse Areola",         "GK",  "France",   "West Ham",         "England",  None, None, None),
    ("FRA-5",  "FRA", 5,  "player", "Brice Samba",             "GK",  "France",   "Lens",             "France",   None, None, None),
    ("FRA-6",  "FRA", 6,  "player", "Theo Hernandez",          "DEF", "France",   "AC Milan",         "Italy",    None, None, None),
    ("FRA-7",  "FRA", 7,  "player", "William Saliba",          "DEF", "France",   "Arsenal",          "England",  None, None, None),
    ("FRA-8",  "FRA", 8,  "player", "Dayot Upamecano",         "DEF", "France",   "Bayern Munich",    "Germany",  None, None, None),
    ("FRA-9",  "FRA", 9,  "player", "Jules Koundé",            "DEF", "France",   "Barcelona",        "Spain",    None, None, None),
    ("FRA-10", "FRA", 10, "player", "Benjamin Pavard",         "DEF", "France",   "Inter Milan",      "Italy",    None, None, None),
    ("FRA-11", "FRA", 11, "player", "N'Golo Kanté",            "MID", "France",   "Al-Ittihad",       "Saudi Arabia", None, None, None),
    ("FRA-12", "FRA", 12, "player", "Aurélien Tchouaméni",     "MID", "France",   "Real Madrid",      "Spain",    None, None, None),
    ("FRA-13", "FRA", 13, "player", "Eduardo Camavinga",       "MID", "France",   "Real Madrid",      "Spain",    None, None, None),
    ("FRA-14", "FRA", 14, "player", "Warren Zaïre-Emery",      "MID", "France",   "PSG",              "France",   None, None, None),
    ("FRA-15", "FRA", 15, "player", "Kylian Mbappé",           "FWD", "France",   "Real Madrid",      "Spain",    None, None, None),
    ("FRA-16", "FRA", 16, "player", "Antoine Griezmann",       "FWD", "France",   "Atlético Madrid",  "Spain",    None, None, None),
    ("FRA-17", "FRA", 17, "player", "Ousmane Dembélé",         "FWD", "France",   "PSG",              "France",   None, None, None),
    ("FRA-18", "FRA", 18, "player", "Marcus Thuram",           "FWD", "France",   "Inter Milan",      "Italy",    None, None, None),
    ("FRA-19", "FRA", 19, "player", "Randal Kolo Muani",       "FWD", "France",   "Juventus",         "Italy",    None, None, None),
    ("FRA-20", "FRA", 20, "player", "Bradley Barcola",         "FWD", "France",   "PSG",              "France",   None, None, None),

    # ── COLOMBIA (COL) ──────────────────────────────────────────────────────
    ("COL-1",  "COL", 1,  "badge",  None,                      None,  "Colombia", None,               None,  None, None, None),
    ("COL-2",  "COL", 2,  "team",   None,                      None,  "Colombia", None,               None,  None, None, None),
    ("COL-3",  "COL", 3,  "player", "Camilo Vargas",           "GK",  "Colombia", "Atlas",            "Mexico",   None, None, None),
    ("COL-4",  "COL", 4,  "player", "David Ospina",            "GK",  "Colombia", "Al-Qadsiah",       "Saudi Arabia", None, None, None),
    ("COL-5",  "COL", 5,  "player", "Kevin Mier",              "GK",  "Colombia", "Club Nacional",    "Uruguay",  None, None, None),
    ("COL-6",  "COL", 6,  "player", "Dávinson Sánchez",        "DEF", "Colombia", "Galatasaray",      "Turkey",   None, None, None),
    ("COL-7",  "COL", 7,  "player", "Yerry Mina",              "DEF", "Colombia", "Fiorentina",       "Italy",    None, None, None),
    ("COL-8",  "COL", 8,  "player", "Daniel Muñoz",            "DEF", "Colombia", "Crystal Palace",   "England",  None, None, None),
    ("COL-9",  "COL", 9,  "player", "Johan Mojica",            "DEF", "Colombia", "Villarreal",       "Spain",    None, None, None),
    ("COL-10", "COL", 10, "player", "Carlos Cuesta",           "DEF", "Colombia", "Genk",             "Belgium",  None, None, None),
    ("COL-11", "COL", 11, "player", "James Rodríguez",         "MID", "Colombia", "Rayo Vallecano",   "Spain",    None, None, None),
    ("COL-12", "COL", 12, "player", "Wilmar Barrios",          "MID", "Colombia", "Zenit",            "Russia",   None, None, None),
    ("COL-13", "COL", 13, "player", "Mateus Uribe",            "MID", "Colombia", "Porto",            "Portugal", None, None, None),
    ("COL-14", "COL", 14, "player", "Richard Ríos",            "MID", "Colombia", "Palmeiras",        "Brazil",   None, None, None),
    ("COL-15", "COL", 15, "player", "Jhon Arias",              "MID", "Colombia", "Fluminense",       "Brazil",   None, None, None),
    ("COL-16", "COL", 16, "player", "Luis Díaz",               "FWD", "Colombia", "Liverpool",        "England",  None, None, None),
    ("COL-17", "COL", 17, "player", "Jhon Córdoba",            "FWD", "Colombia", "Krasnodar",        "Russia",   None, None, None),
    ("COL-18", "COL", 18, "player", "Rafael Santos Borré",     "FWD", "Colombia", "Internacional",    "Brazil",   None, None, None),
    ("COL-19", "COL", 19, "player", "Cucho Hernández",         "FWD", "Colombia", "Columbus Crew",    "USA",      None, None, None),
    ("COL-20", "COL", 20, "player", "Miguel Ángel Borja",      "FWD", "Colombia", "River Plate",      "Argentina", None, None, None),
]

# ── Paths ────────────────────────────────────────────────────────────────────

SCRIPT_DIR   = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR     = os.path.abspath(os.path.join(SCRIPT_DIR, "..", ".."))
SQL_OUT      = os.path.join(ROOT_DIR, "panini-api", "migrations", "002_seed_stickers.sql")
SQLITE_OUT   = os.path.join(ROOT_DIR, "panini-ios", "panini", "panini", "Resources", "stickers.sqlite")

# ── SQLite ───────────────────────────────────────────────────────────────────

def build_sqlite():
    os.makedirs(os.path.dirname(SQLITE_OUT), exist_ok=True)
    if os.path.exists(SQLITE_OUT):
        os.remove(SQLITE_OUT)

    conn = sqlite3.connect(SQLITE_OUT)
    c = conn.cursor()
    c.execute("""
        CREATE TABLE stickers (
            id              TEXT PRIMARY KEY,
            country_code    TEXT NOT NULL,
            sticker_number  INTEGER NOT NULL,
            type            TEXT NOT NULL,
            player_name     TEXT,
            position        TEXT,
            national_team   TEXT NOT NULL,
            club            TEXT,
            club_country    TEXT,
            dob             TEXT,
            height_cm       REAL,
            weight_kg       REAL
        )
    """)
    c.execute("CREATE INDEX idx_stickers_country ON stickers(country_code)")
    c.execute("CREATE INDEX idx_stickers_position ON stickers(position)")
    c.execute("CREATE INDEX idx_stickers_type ON stickers(type)")
    c.executemany(
        "INSERT INTO stickers VALUES (?,?,?,?,?,?,?,?,?,?,?,?)",
        STICKERS
    )
    conn.commit()
    conn.close()
    print(f"SQLite  → {SQLITE_OUT}  ({len(STICKERS)} stickers)")

# ── Postgres SQL ─────────────────────────────────────────────────────────────

def build_sql():
    def q(v):
        if v is None:
            return "NULL"
        return "'" + str(v).replace("'", "''") + "'"

    lines = [
        "-- Sticker seed data (dev subset: ESP, FRA, COL).",
        "-- Generated by panini-api/scripts/generate_stickers.py — do not edit by hand.",
        "",
        "INSERT INTO stickers",
        "  (id, country_code, sticker_number, type, player_name, position, national_team, club, club_country, dob, height_cm, weight_kg)",
        "VALUES",
    ]
    rows = []
    for s in STICKERS:
        vals = ", ".join(q(v) for v in s)
        rows.append(f"  ({vals})")
    lines.append(",\n".join(rows) + ";")

    with open(SQL_OUT, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")
    print(f"SQL     → {SQL_OUT}")

# ── Main ─────────────────────────────────────────────────────────────────────

if __name__ == "__main__":
    build_sqlite()
    build_sql()
    print(f"Done. {len(STICKERS)} stickers written.")
