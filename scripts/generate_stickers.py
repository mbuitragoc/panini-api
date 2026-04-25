#!/usr/bin/env python3
"""
Generate sticker seed data for the full Panini FIFA World Cup 2026 album.

Structure per team (20 stickers):
  position  1  → badge (team crest)
  positions 2–12 → 11 player stickers
  position 13  → team photo
  positions 14–20 → 7 player stickers

Special section (FWC code):
  FWC-1  … FWC-22 → one sticker per past World Cup winner (1930–2022)
  FWC-23 … FWC-25 → 2026 trophy, official ball, back-cover Panini logo

Player data (name, position, club) is NULL for teams without confirmed squads.
The importplayers pipeline backfills dob / height_cm / weight_kg after matching.

Run from the panini-app/ root:
  python3 panini-api/scripts/generate_stickers.py
"""

import os
import sqlite3

# ── Helpers ──────────────────────────────────────────────────────────────────

def placeholder_team(code, name):
    """Return 20 placeholder sticker rows for a team whose squad is not yet known."""
    rows = []
    rows.append((f"{code}-1",  code, 1,  "badge", None, None, name, None, None, None, None, None))
    for i in range(2, 13):      # positions 2–12: 11 players
        rows.append((f"{code}-{i}", code, i, "player", None, None, name, None, None, None, None, None))
    rows.append((f"{code}-13", code, 13, "team",  None, None, name, None, None, None, None, None))
    for i in range(14, 21):     # positions 14–20: 7 players
        rows.append((f"{code}-{i}", code, i, "player", None, None, name, None, None, None, None, None))
    return rows


def player(code, name, pos, num, national_team, club=None, club_country=None):
    return (f"{code}-{num}", code, num, "player", pos, num, national_team, club, club_country, None, None, None)


# ── Source of truth ───────────────────────────────────────────────────────────
# Fields: id, country_code, sticker_number, type, player_name, position,
#         national_team, club, club_country, dob, height_cm, weight_kg

STICKERS = []

# ═══════════════════════════════════════════════════════════════════════════════
# CONFIRMED SQUADS (ESP · FRA · COL)
# Badge=1, players=2–12, team photo=13, players=14–20
# ═══════════════════════════════════════════════════════════════════════════════

# ── SPAIN (ESP) ───────────────────────────────────────────────────────────────
STICKERS += [
    ("ESP-1",  "ESP", 1,  "badge",  None,                 None,  "Spain", None,               None,          None, None, None),
    ("ESP-2",  "ESP", 2,  "player", "Unai Simón",         "GK",  "Spain", "Athletic Club",    "Spain",       None, None, None),
    ("ESP-3",  "ESP", 3,  "player", "David Raya",         "GK",  "Spain", "Arsenal",          "England",     None, None, None),
    ("ESP-4",  "ESP", 4,  "player", "Alex Remiro",        "GK",  "Spain", "Real Sociedad",    "Spain",       None, None, None),
    ("ESP-5",  "ESP", 5,  "player", "Dani Carvajal",      "DEF", "Spain", "Real Madrid",      "Spain",       None, None, None),
    ("ESP-6",  "ESP", 6,  "player", "Robin Le Normand",   "DEF", "Spain", "Atlético Madrid",  "Spain",       None, None, None),
    ("ESP-7",  "ESP", 7,  "player", "Aymeric Laporte",    "DEF", "Spain", "Al-Nassr",         "Saudi Arabia",None, None, None),
    ("ESP-8",  "ESP", 8,  "player", "Marc Cucurella",     "DEF", "Spain", "Chelsea",          "England",     None, None, None),
    ("ESP-9",  "ESP", 9,  "player", "Alejandro Grimaldo", "DEF", "Spain", "Bayer Leverkusen", "Germany",     None, None, None),
    ("ESP-10", "ESP", 10, "player", "Rodri",              "MID", "Spain", "Manchester City",  "England",     None, None, None),
    ("ESP-11", "ESP", 11, "player", "Pedri",              "MID", "Spain", "Barcelona",        "Spain",       None, None, None),
    ("ESP-12", "ESP", 12, "player", "Fabián Ruiz",        "MID", "Spain", "PSG",              "France",      None, None, None),
    ("ESP-13", "ESP", 13, "team",   None,                 None,  "Spain", None,               None,          None, None, None),
    ("ESP-14", "ESP", 14, "player", "Mikel Merino",       "MID", "Spain", "Arsenal",          "England",     None, None, None),
    ("ESP-15", "ESP", 15, "player", "Dani Olmo",          "MID", "Spain", "Barcelona",        "Spain",       None, None, None),
    ("ESP-16", "ESP", 16, "player", "Lamine Yamal",       "FWD", "Spain", "Barcelona",        "Spain",       None, None, None),
    ("ESP-17", "ESP", 17, "player", "Nico Williams",      "FWD", "Spain", "Athletic Club",    "Spain",       None, None, None),
    ("ESP-18", "ESP", 18, "player", "Álvaro Morata",      "FWD", "Spain", "AC Milan",         "Italy",       None, None, None),
    ("ESP-19", "ESP", 19, "player", "Ferran Torres",      "FWD", "Spain", "Barcelona",        "Spain",       None, None, None),
    ("ESP-20", "ESP", 20, "player", "Mikel Oyarzabal",    "FWD", "Spain", "Real Sociedad",    "Spain",       None, None, None),
]

# ── FRANCE (FRA) ──────────────────────────────────────────────────────────────
STICKERS += [
    ("FRA-1",  "FRA", 1,  "badge",  None,                      None,  "France", None,              None,          None, None, None),
    ("FRA-2",  "FRA", 2,  "player", "Mike Maignan",            "GK",  "France", "AC Milan",        "Italy",       None, None, None),
    ("FRA-3",  "FRA", 3,  "player", "Alphonse Areola",         "GK",  "France", "West Ham",        "England",     None, None, None),
    ("FRA-4",  "FRA", 4,  "player", "Brice Samba",             "GK",  "France", "Lens",            "France",      None, None, None),
    ("FRA-5",  "FRA", 5,  "player", "Theo Hernandez",          "DEF", "France", "AC Milan",        "Italy",       None, None, None),
    ("FRA-6",  "FRA", 6,  "player", "William Saliba",          "DEF", "France", "Arsenal",         "England",     None, None, None),
    ("FRA-7",  "FRA", 7,  "player", "Dayot Upamecano",         "DEF", "France", "Bayern Munich",   "Germany",     None, None, None),
    ("FRA-8",  "FRA", 8,  "player", "Jules Koundé",            "DEF", "France", "Barcelona",       "Spain",       None, None, None),
    ("FRA-9",  "FRA", 9,  "player", "Benjamin Pavard",         "DEF", "France", "Inter Milan",     "Italy",       None, None, None),
    ("FRA-10", "FRA", 10, "player", "N'Golo Kanté",            "MID", "France", "Al-Ittihad",      "Saudi Arabia",None, None, None),
    ("FRA-11", "FRA", 11, "player", "Aurélien Tchouaméni",     "MID", "France", "Real Madrid",     "Spain",       None, None, None),
    ("FRA-12", "FRA", 12, "player", "Eduardo Camavinga",       "MID", "France", "Real Madrid",     "Spain",       None, None, None),
    ("FRA-13", "FRA", 13, "team",   None,                      None,  "France", None,              None,          None, None, None),
    ("FRA-14", "FRA", 14, "player", "Warren Zaïre-Emery",      "MID", "France", "PSG",             "France",      None, None, None),
    ("FRA-15", "FRA", 15, "player", "Kylian Mbappé",           "FWD", "France", "Real Madrid",     "Spain",       None, None, None),
    ("FRA-16", "FRA", 16, "player", "Antoine Griezmann",       "FWD", "France", "Atlético Madrid", "Spain",       None, None, None),
    ("FRA-17", "FRA", 17, "player", "Ousmane Dembélé",         "FWD", "France", "PSG",             "France",      None, None, None),
    ("FRA-18", "FRA", 18, "player", "Marcus Thuram",           "FWD", "France", "Inter Milan",     "Italy",       None, None, None),
    ("FRA-19", "FRA", 19, "player", "Randal Kolo Muani",       "FWD", "France", "Juventus",        "Italy",       None, None, None),
    ("FRA-20", "FRA", 20, "player", "Bradley Barcola",         "FWD", "France", "PSG",             "France",      None, None, None),
]

# ── COLOMBIA (COL) ────────────────────────────────────────────────────────────
STICKERS += [
    ("COL-1",  "COL", 1,  "badge",  None,                     None,  "Colombia", None,             None,          None, None, None),
    ("COL-2",  "COL", 2,  "player", "Camilo Vargas",          "GK",  "Colombia", "Atlas",          "Mexico",      None, None, None),
    ("COL-3",  "COL", 3,  "player", "David Ospina",           "GK",  "Colombia", "Al-Qadsiah",     "Saudi Arabia",None, None, None),
    ("COL-4",  "COL", 4,  "player", "Kevin Mier",             "GK",  "Colombia", "Club Nacional",  "Uruguay",     None, None, None),
    ("COL-5",  "COL", 5,  "player", "Dávinson Sánchez",       "DEF", "Colombia", "Galatasaray",    "Turkey",      None, None, None),
    ("COL-6",  "COL", 6,  "player", "Yerry Mina",             "DEF", "Colombia", "Fiorentina",     "Italy",       None, None, None),
    ("COL-7",  "COL", 7,  "player", "Daniel Muñoz",           "DEF", "Colombia", "Crystal Palace", "England",     None, None, None),
    ("COL-8",  "COL", 8,  "player", "Johan Mojica",           "DEF", "Colombia", "Villarreal",     "Spain",       None, None, None),
    ("COL-9",  "COL", 9,  "player", "Carlos Cuesta",          "DEF", "Colombia", "Genk",           "Belgium",     None, None, None),
    ("COL-10", "COL", 10, "player", "James Rodríguez",        "MID", "Colombia", "Rayo Vallecano", "Spain",       None, None, None),
    ("COL-11", "COL", 11, "player", "Wilmar Barrios",         "MID", "Colombia", "Zenit",          "Russia",      None, None, None),
    ("COL-12", "COL", 12, "player", "Mateus Uribe",           "MID", "Colombia", "Porto",          "Portugal",    None, None, None),
    ("COL-13", "COL", 13, "team",   None,                     None,  "Colombia", None,             None,          None, None, None),
    ("COL-14", "COL", 14, "player", "Richard Ríos",           "MID", "Colombia", "Palmeiras",      "Brazil",      None, None, None),
    ("COL-15", "COL", 15, "player", "Jhon Arias",             "MID", "Colombia", "Fluminense",     "Brazil",      None, None, None),
    ("COL-16", "COL", 16, "player", "Luis Díaz",              "FWD", "Colombia", "Liverpool",      "England",     None, None, None),
    ("COL-17", "COL", 17, "player", "Jhon Córdoba",           "FWD", "Colombia", "Krasnodar",      "Russia",      None, None, None),
    ("COL-18", "COL", 18, "player", "Rafael Santos Borré",    "FWD", "Colombia", "Internacional",  "Brazil",      None, None, None),
    ("COL-19", "COL", 19, "player", "Cucho Hernández",        "FWD", "Colombia", "Columbus Crew",  "USA",         None, None, None),
    ("COL-20", "COL", 20, "player", "Miguel Ángel Borja",     "FWD", "Colombia", "River Plate",    "Argentina",   None, None, None),
]

# ═══════════════════════════════════════════════════════════════════════════════
# PLACEHOLDER SQUADS — all 45 remaining WC2026 teams
# Player names, positions, clubs are NULL (filled via scan + importplayers)
# ═══════════════════════════════════════════════════════════════════════════════

PLACEHOLDER_TEAMS = [
    # CONMEBOL
    ("ARG", "Argentina"),
    ("BRA", "Brazil"),
    ("ECU", "Ecuador"),
    ("PAR", "Paraguay"),
    ("URU", "Uruguay"),
    # UEFA
    ("AUT", "Austria"),
    ("BEL", "Belgium"),
    ("BIH", "Bosnia & Herzegovina"),
    ("CRO", "Croatia"),
    ("CZE", "Czech Republic"),
    ("ENG", "England"),
    ("GER", "Germany"),
    ("NED", "Netherlands"),
    ("NOR", "Norway"),
    ("POR", "Portugal"),
    ("SCO", "Scotland"),
    ("SWE", "Sweden"),
    ("SUI", "Switzerland"),
    ("TUR", "Türkiye"),
    # AFC
    ("AUS", "Australia"),
    ("IRN", "Iran"),
    ("IRQ", "Iraq"),
    ("JPN", "Japan"),
    ("JOR", "Jordan"),
    ("QAT", "Qatar"),
    ("KSA", "Saudi Arabia"),
    ("KOR", "South Korea"),
    ("UZB", "Uzbekistan"),
    # CAF
    ("ALG", "Algeria"),
    ("CPV", "Cape Verde"),
    ("COD", "DR Congo"),
    ("EGY", "Egypt"),
    ("GHA", "Ghana"),
    ("CIV", "Côte d'Ivoire"),
    ("MAR", "Morocco"),
    ("SEN", "Senegal"),
    ("RSA", "South Africa"),
    ("TUN", "Tunisia"),
    # CONCACAF
    ("CAN", "Canada"),
    ("CUW", "Curaçao"),
    ("HAI", "Haiti"),
    ("MEX", "Mexico"),
    ("PAN", "Panama"),
    ("USA", "United States"),
    # OFC
    ("NZL", "New Zealand"),
]

for code, name in PLACEHOLDER_TEAMS:
    STICKERS += placeholder_team(code, name)

# ═══════════════════════════════════════════════════════════════════════════════
# FWC — Special section: past World Cup winners + 2026 specials
# One sticker per tournament edition (22 editions, 1930–2022), plus
# the 2026 trophy, official ball, and back-cover Panini logo sticker.
# ═══════════════════════════════════════════════════════════════════════════════

FWC_WINNERS = [
    (1,  "URU", "Uruguay 1930"),
    (2,  "ITA", "Italy 1934"),
    (3,  "ITA", "Italy 1938"),
    (4,  "URU", "Uruguay 1950"),
    (5,  "GER", "West Germany 1954"),
    (6,  "BRA", "Brazil 1958"),
    (7,  "BRA", "Brazil 1962"),
    (8,  "ENG", "England 1966"),
    (9,  "BRA", "Brazil 1970"),
    (10, "GER", "West Germany 1974"),
    (11, "ARG", "Argentina 1978"),
    (12, "ITA", "Italy 1982"),
    (13, "ARG", "Argentina 1986"),
    (14, "GER", "West Germany 1990"),
    (15, "BRA", "Brazil 1994"),
    (16, "FRA", "France 1998"),
    (17, "BRA", "Brazil 2002"),
    (18, "ITA", "Italy 2006"),
    (19, "ESP", "Spain 2010"),
    (20, "GER", "Germany 2014"),
    (21, "FRA", "France 2018"),
    (22, "ARG", "Argentina 2022"),
]

for num, winner_code, label in FWC_WINNERS:
    STICKERS.append((
        f"FWC-{num}", "FWC", num, "special", label, None,
        "FIFA World Cup", None, None, None, None, None,
    ))

# FWC-23  2026 trophy sticker
STICKERS.append(("FWC-23", "FWC", 23, "special", "FIFA World Cup 2026™ Trophy", None, "FIFA World Cup", None, None, None, None, None))
# FWC-24  official match ball
STICKERS.append(("FWC-24", "FWC", 24, "special", "FIFA World Cup 2026™ Official Ball", None, "FIFA World Cup", None, None, None, None, None))
# FWC-25  back-cover Panini logo
STICKERS.append(("FWC-25", "FWC", 25, "special", "Panini Official", None, "FIFA World Cup", None, None, None, None, None))

# ── Paths ─────────────────────────────────────────────────────────────────────

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
ROOT_DIR   = os.path.abspath(os.path.join(SCRIPT_DIR, "..", ".."))
SQL_OUT    = os.path.join(ROOT_DIR, "panini-api", "migrations", "002_seed_stickers.sql")
SQLITE_OUT = os.path.join(ROOT_DIR, "panini-ios", "panini", "panini", "Resources", "stickers.sqlite")

# ── SQLite ────────────────────────────────────────────────────────────────────

def build_sqlite():
    os.makedirs(os.path.dirname(SQLITE_OUT), exist_ok=True)
    if os.path.exists(SQLITE_OUT):
        os.remove(SQLITE_OUT)

    conn = sqlite3.connect(SQLITE_OUT)
    c = conn.cursor()
    c.executescript("""
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
        );
        CREATE INDEX idx_stickers_country  ON stickers(country_code);
        CREATE INDEX idx_stickers_position ON stickers(position);
        CREATE INDEX idx_stickers_type     ON stickers(type);
    """)
    c.executemany("INSERT INTO stickers VALUES (?,?,?,?,?,?,?,?,?,?,?,?)", STICKERS)
    conn.commit()
    conn.close()
    print(f"SQLite  → {SQLITE_OUT}  ({len(STICKERS)} stickers)")

# ── Postgres SQL ──────────────────────────────────────────────────────────────

def build_sql():
    def q(v):
        if v is None:
            return "NULL"
        return "'" + str(v).replace("'", "''") + "'"

    lines = [
        "-- Sticker seed data — Panini FIFA World Cup 2026 full album (48 teams + FWC specials).",
        "-- Generated by panini-api/scripts/generate_stickers.py — do not edit by hand.",
        "",
        "INSERT INTO stickers",
        "  (id, country_code, sticker_number, type, player_name, position, national_team, club, club_country, dob, height_cm, weight_kg)",
        "VALUES",
    ]
    rows = [f"  ({', '.join(q(v) for v in s)})" for s in STICKERS]
    lines.append(",\n".join(rows) + ";")

    with open(SQL_OUT, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")
    print(f"SQL     → {SQL_OUT}")

# ── Main ──────────────────────────────────────────────────────────────────────

if __name__ == "__main__":
    build_sqlite()
    build_sql()

    confirmed = sum(1 for s in STICKERS if s[4] is not None)   # has player_name
    placeholder = sum(1 for s in STICKERS if s[3] == "player" and s[4] is None)
    print(f"Done.   {len(STICKERS)} stickers total")
    print(f"        {confirmed} with confirmed names · {placeholder} player placeholders to fill via scan")
