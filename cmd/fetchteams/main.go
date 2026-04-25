// Command fetchteams fetches world football rankings and writes normalized team
// OVR ratings into the bundled stickers.sqlite.
//
// Usage:
//
//	go run ./cmd/fetchteams                          # fetch from eloratings.net (default)
//	go run ./cmd/fetchteams --source fifa            # FIFA API (may be geo-blocked)
//	go run ./cmd/fetchteams --input rankings.json    # use a local FIFA JSON file
//	go run ./cmd/fetchteams --date-id id2025         # specific FIFA ranking date
//	go run ./cmd/fetchteams --dry-run                # print results, don't write
package main

import (
	"bufio"
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
	"strconv"
	"strings"
	"time"

	"github.com/mbuitragoc/panini-api/internal/pipeline"
	_ "modernc.org/sqlite"
)

// fifaEntry is one row from the FIFA ranking API response.
// Also used as a common intermediate when parsing the Elo TSV source.
type fifaEntry struct {
	Rank          int     `json:"rankId"`
	PreviousRank  int     `json:"previousRankId"`
	CountryCode   string  `json:"countryCode"`
	ShortCode     string  `json:"shortCountryCode"`
	TeamName      string  `json:"teamName"`
	Points        float64 `json:"totalPoints"`
	Confederation string  `json:"confederation"`
}

type fifaResponse struct {
	Rankings []fifaEntry `json:"rankings"`
}

func main() {
	dbPath := flag.String("db", defaultDB(), "path to stickers.sqlite")
	source := flag.String("source", "elo", "ranking source: elo (eloratings.net) or fifa")
	inputPath := flag.String("input", "", "local JSON file from FIFA API (skips HTTP fetch, implies --source fifa)")
	dateID := flag.String("date-id", "", "FIFA ranking dateId (e.g. id2025); uses latest if empty")
	dryRun := flag.Bool("dry-run", false, "print results without writing to SQLite")
	flag.Parse()

	var entries []fifaEntry
	var err error

	if *inputPath != "" || *source == "fifa" {
		var data []byte
		data, err = fetchFIFARankings(*inputPath, *dateID)
		if err != nil {
			log.Fatalf("fetch FIFA rankings: %v", err)
		}
		entries, err = parseFIFARankings(data)
	} else {
		entries, err = fetchEloRankings()
	}
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

// ── Elo source ────────────────────────────────────────────────────────────────

type eloMeta struct {
	code string // FIFA 3-letter code
	name string
	conf string // confederation
}

// eloCodeMap maps eloratings.net 2-letter codes → FIFA metadata.
// eloratings.net uses custom codes (EN=England, SQ=Scotland, etc.).
var eloCodeMap = map[string]eloMeta{
	// UEFA
	"ES": {"ESP", "Spain", "UEFA"},
	"FR": {"FRA", "France", "UEFA"},
	"EN": {"ENG", "England", "UEFA"},
	"PT": {"POR", "Portugal", "UEFA"},
	"NL": {"NED", "Netherlands", "UEFA"},
	"HR": {"CRO", "Croatia", "UEFA"},
	"DE": {"GER", "Germany", "UEFA"},
	"TR": {"TUR", "Türkiye", "UEFA"},
	"CH": {"SUI", "Switzerland", "UEFA"},
	"DK": {"DEN", "Denmark", "UEFA"},
	"BE": {"BEL", "Belgium", "UEFA"},
	"IT": {"ITA", "Italy", "UEFA"},
	"AT": {"AUT", "Austria", "UEFA"},
	"RS": {"SRB", "Serbia", "UEFA"},
	"UA": {"UKR", "Ukraine", "UEFA"},
	"NO": {"NOR", "Norway", "UEFA"},
	"SQ": {"SCO", "Scotland", "UEFA"},
	"GR": {"GRE", "Greece", "UEFA"},
	"PL": {"POL", "Poland", "UEFA"},
	"CZ": {"CZE", "Czech Republic", "UEFA"},
	"SE": {"SWE", "Sweden", "UEFA"},
	"HU": {"HUN", "Hungary", "UEFA"},
	"WA": {"WAL", "Wales", "UEFA"},
	"SI": {"SVN", "Slovenia", "UEFA"},
	"IE": {"IRL", "Republic of Ireland", "UEFA"},
	"KO": {"KOS", "Kosovo", "UEFA"},
	"SK": {"SVK", "Slovakia", "UEFA"},
	"GE": {"GEO", "Georgia", "UEFA"},
	"AL": {"ALB", "Albania", "UEFA"},
	"IL": {"ISR", "Israel", "UEFA"},
	"RO": {"ROU", "Romania", "UEFA"},
	"BA": {"BIH", "Bosnia & Herzegovina", "UEFA"},
	"NM": {"MKD", "North Macedonia", "UEFA"},
	"IS": {"ISL", "Iceland", "UEFA"},
	"FI": {"FIN", "Finland", "UEFA"},
	"BG": {"BUL", "Bulgaria", "UEFA"},
	"ME": {"MNE", "Montenegro", "UEFA"},
	"LU": {"LUX", "Luxembourg", "UEFA"},
	"LV": {"LVA", "Latvia", "UEFA"},
	"EE": {"EST", "Estonia", "UEFA"},
	"LT": {"LTU", "Lithuania", "UEFA"},
	"CY": {"CYP", "Cyprus", "UEFA"},
	"AM": {"ARM", "Armenia", "UEFA"},
	"AZ": {"AZE", "Azerbaijan", "UEFA"},
	"BY": {"BLR", "Belarus", "UEFA"},
	"MD": {"MDA", "Moldova", "UEFA"},
	"FO": {"FRO", "Faroe Islands", "UEFA"},
	"NS": {"NIR", "Northern Ireland", "UEFA"},
	"EI": {"IRL", "Republic of Ireland", "UEFA"}, // eloratings duplicate for Eire
	// CONMEBOL
	"AR": {"ARG", "Argentina", "CONMEBOL"},
	"BR": {"BRA", "Brazil", "CONMEBOL"},
	"CO": {"COL", "Colombia", "CONMEBOL"},
	"EC": {"ECU", "Ecuador", "CONMEBOL"},
	"UY": {"URU", "Uruguay", "CONMEBOL"},
	"PY": {"PAR", "Paraguay", "CONMEBOL"},
	"VE": {"VEN", "Venezuela", "CONMEBOL"},
	"CL": {"CHI", "Chile", "CONMEBOL"},
	"PE": {"PER", "Peru", "CONMEBOL"},
	"BO": {"BOL", "Bolivia", "CONMEBOL"},
	// CONCACAF
	"US": {"USA", "United States", "CONCACAF"},
	"CA": {"CAN", "Canada", "CONCACAF"},
	"MX": {"MEX", "Mexico", "CONCACAF"},
	"PA": {"PAN", "Panama", "CONCACAF"},
	"JM": {"JAM", "Jamaica", "CONCACAF"},
	"CR": {"CRC", "Costa Rica", "CONCACAF"},
	"HN": {"HON", "Honduras", "CONCACAF"},
	"GT": {"GUA", "Guatemala", "CONCACAF"},
	"SV": {"SLV", "El Salvador", "CONCACAF"},
	"HT": {"HAI", "Haiti", "CONCACAF"},
	"TT": {"TRI", "Trinidad & Tobago", "CONCACAF"},
	"CW": {"CUW", "Curaçao", "CONCACAF"},
	"NI": {"NCA", "Nicaragua", "CONCACAF"},
	"DO": {"DOM", "Dominican Republic", "CONCACAF"},
	"GY": {"GUY", "Guyana", "CONCACAF"},
	"SR": {"SUR", "Suriname", "CONCACAF"},
	// AFC
	"JP": {"JPN", "Japan", "AFC"},
	"KR": {"KOR", "South Korea", "AFC"},
	"IR": {"IRN", "Iran", "AFC"},
	"AU": {"AUS", "Australia", "AFC"},
	"JO": {"JOR", "Jordan", "AFC"},
	"UZ": {"UZB", "Uzbekistan", "AFC"},
	"SA": {"KSA", "Saudi Arabia", "AFC"},
	"IQ": {"IRQ", "Iraq", "AFC"},
	"OM": {"OMA", "Oman", "AFC"},
	"QA": {"QAT", "Qatar", "AFC"},
	"AE": {"UAE", "United Arab Emirates", "AFC"},
	"CN": {"CHN", "China", "AFC"},
	"KP": {"PRK", "North Korea", "AFC"},
	"KZ": {"KAZ", "Kazakhstan", "AFC"},
	"KG": {"KGZ", "Kyrgyzstan", "AFC"},
	"TJ": {"TJK", "Tajikistan", "AFC"},
	"TH": {"THA", "Thailand", "AFC"},
	"VN": {"VIE", "Vietnam", "AFC"},
	"MY": {"MAS", "Malaysia", "AFC"},
	"ID": {"IDN", "Indonesia", "AFC"},
	"KW": {"KUW", "Kuwait", "AFC"},
	"BH": {"BHR", "Bahrain", "AFC"},
	"SY": {"SYR", "Syria", "AFC"},
	"PS": {"PLE", "Palestine", "AFC"},
	"LB": {"LBN", "Lebanon", "AFC"},
	// CAF
	"MA": {"MAR", "Morocco", "CAF"},
	"SN": {"SEN", "Senegal", "CAF"},
	"NG": {"NGA", "Nigeria", "CAF"},
	"DZ": {"ALG", "Algeria", "CAF"},
	"CI": {"CIV", "Côte d'Ivoire", "CAF"},
	"CD": {"COD", "DR Congo", "CAF"},
	"EG": {"EGY", "Egypt", "CAF"},
	"ML": {"MLI", "Mali", "CAF"},
	"ZA": {"RSA", "South Africa", "CAF"},
	"CM": {"CMR", "Cameroon", "CAF"},
	"TN": {"TUN", "Tunisia", "CAF"},
	"GH": {"GHA", "Ghana", "CAF"},
	"CV": {"CPV", "Cape Verde", "CAF"},
	"AO": {"ANG", "Angola", "CAF"},
	"BF": {"BFA", "Burkina Faso", "CAF"},
	"BJ": {"BEN", "Benin", "CAF"},
	"GN": {"GUI", "Guinea", "CAF"},
	"GQ": {"EQG", "Equatorial Guinea", "CAF"},
	"TG": {"TOG", "Togo", "CAF"},
	"ZW": {"ZIM", "Zimbabwe", "CAF"},
	"ZM": {"ZAM", "Zambia", "CAF"},
	"MZ": {"MOZ", "Mozambique", "CAF"},
	"KM": {"COM", "Comoros", "CAF"},
	"UG": {"UGA", "Uganda", "CAF"},
	"KE": {"KEN", "Kenya", "CAF"},
	"TZ": {"TAN", "Tanzania", "CAF"},
	"RW": {"RWA", "Rwanda", "CAF"},
	"SD": {"SDN", "Sudan", "CAF"},
	"ET": {"ETH", "Ethiopia", "CAF"},
	"LY": {"LBA", "Libya", "CAF"},
	"MR": {"MTN", "Mauritania", "CAF"},
	"GA": {"GAB", "Gabon", "CAF"},
	"NA": {"NAM", "Namibia", "CAF"},
	"BW": {"BOT", "Botswana", "CAF"},
	"MW": {"MWI", "Malawi", "CAF"},
	"GW": {"GNB", "Guinea-Bissau", "CAF"},
	"SL": {"SLE", "Sierra Leone", "CAF"},
	"LR": {"LBR", "Liberia", "CAF"},
	"BI": {"BDI", "Burundi", "CAF"},
	"MG": {"MDG", "Madagascar", "CAF"},
	"NE": {"NIG", "Niger", "CAF"},
	// OFC
	"NZ": {"NZL", "New Zealand", "OFC"},
	"NC": {"NCL", "New Caledonia", "OFC"},
}

// fetchEloRankings downloads the eloratings.net world TSV and converts it to
// the common fifaEntry slice used by the rest of the pipeline.
func fetchEloRankings() ([]fifaEntry, error) {
	const url = "https://www.eloratings.net/World.tsv"
	resp, err := http.Get(url) //nolint:gosec,noctx
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eloratings.net returned %d", resp.StatusCode)
	}
	return parseEloTSV(resp.Body)
}

// parseEloTSV converts tab-separated eloratings.net rows to fifaEntry values.
// TSV columns: rank, prev_rank, elo_code, elo_rating, ...
// Rows whose code has no mapping in eloCodeMap are silently skipped.
func parseEloTSV(r io.Reader) ([]fifaEntry, error) {
	var entries []fifaEntry
	seen := map[string]bool{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 4 {
			continue
		}

		rank, err1 := strconv.Atoi(strings.TrimSpace(fields[0]))
		prevRank, err2 := strconv.Atoi(strings.TrimSpace(fields[1]))
		eloCode := strings.TrimSpace(fields[2])
		// Elo ratings use Unicode minus signs (U+2212); replace before parsing.
		eloStr := strings.ReplaceAll(strings.TrimSpace(fields[3]), "−", "-")
		points, err3 := strconv.ParseFloat(eloStr, 64)
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}

		meta, ok := eloCodeMap[eloCode]
		if !ok {
			continue
		}
		// Deduplicate: some teams appear twice with the same FIFA code
		// (e.g. eloratings has both IE and EI for Ireland).
		if seen[meta.code] {
			continue
		}
		seen[meta.code] = true

		entries = append(entries, fifaEntry{
			Rank:          rank,
			PreviousRank:  prevRank,
			CountryCode:   meta.code,
			TeamName:      meta.name,
			Points:        points,
			Confederation: meta.conf,
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("no recognized teams found in eloratings TSV")
	}
	return entries, nil
}

// ── FIFA source ───────────────────────────────────────────────────────────────

// fetchFIFARankings returns the raw JSON bytes from either a local file or the FIFA API.
func fetchFIFARankings(inputPath, dateID string) ([]byte, error) {
	if inputPath != "" {
		return os.ReadFile(inputPath)
	}

	if dateID == "" {
		var err error
		dateID, err = latestFIFADateID()
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

func latestFIFADateID() (string, error) {
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

	var dr datesResp
	if err := json.Unmarshal(body, &dr); err == nil && len(dr.Rankings) > 0 {
		return dr.Rankings[0].ID, nil
	}

	var rr fifaResponse
	if err := json.Unmarshal(body, &rr); err == nil && len(rr.Rankings) > 0 {
		return "", nil
	}

	return "", fmt.Errorf("could not determine latest dateId from FIFA API response")
}

func parseFIFARankings(data []byte) ([]fifaEntry, error) {
	var resp fifaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w\n\nFirst 200 bytes: %s", err, truncate(string(data), 200))
	}
	if len(resp.Rankings) == 0 {
		return nil, fmt.Errorf("no rankings found in response; check --input file format")
	}
	return resp.Rankings, nil
}

// ── Shared normalize / write ──────────────────────────────────────────────────

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
			ovr = 85
		} else {
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
			RankChange:        e.PreviousRank - e.Rank,
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
