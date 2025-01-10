package main

import (
    "crypto/rand"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "math/big"
    "net/http"
    "net/url"
    "sort"

    _ "github.com/mattn/go-sqlite3"
)

const (
    dbDriver = "sqlite3"
    dbFile   = "C:\\sqlite\\urlshortener.db"
)

var db *sql.DB
var domainVisitCounts map[string]int

func init() {
    var err error
    db, err = sql.Open(dbDriver, dbFile)
    if err != nil {
        log.Fatal(err)
    }
    _, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS UrlRecord (
        id INTEGER PRIMARY KEY AUTOINCREMENT, 
        longURL TEXT UNIQUE, 
        shortURL TEXT UNIQUE
    )`)
    if err != nil {
        log.Fatal(err)
    }
    domainVisitCounts = make(map[string]int)
}

func Shorten(longURL string) (string, error) {
	var shortURL string
	err := db.QueryRow("SELECT shortURL FROM UrlRecord WHERE longURL = ?", longURL).Scan(&shortURL)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error querying for existing short URL:", err)
		return "", err
	}
	if shortURL != "" {
		log.Printf("URL already shortened: %s -> %s\n", longURL, shortURL)
		return shortURL, nil
	}
	shortURL = randomStringGenerator(8)

	var existingShortURL string
	for {
		shortURL = randomStringGenerator(8)
		err = db.QueryRow("SELECT shortURL FROM UrlRecord WHERE shortURL = ?", shortURL).Scan(&existingShortURL)
		if err == sql.ErrNoRows {
			break
		}
		if err != nil {
			return "", err
		}
	}
	_, err = db.Exec("INSERT INTO UrlRecord (longURL, shortURL) VALUES (?, ?)", longURL, shortURL)
	if err != nil {
		log.Println("Error inserting new URL into database:", err)
		return "", err
	}
	log.Printf("Successfully inserted: %s -> %s\n", longURL, shortURL)
	return shortURL, nil
}
func randomStringGenerator(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	random := make([]byte, length)
	for i := range random {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		random[i] = charset[randomIndex.Int64()]
	}
	return string(random)
}
func updateDomainCount(longURL string) {
	url, err := url.Parse(longURL)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return
	}

	domain := url.Hostname()
	domainVisitCounts[domain]++
}
func Redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[len("/redirect/"):]
	var longURL string
	err := db.QueryRow("SELECT longURL FROM UrlRecord WHERE shortURL = ?", shortURL).Scan(&longURL)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Invalid Short Url", http.StatusNotFound)
		log.Printf("Error fetching long URL for short URL: %s", shortURL)
		return
	}
	log.Printf("Resolved long URL: %s", longURL)

	if longURL == "" {
		http.Error(w, "Short Url not Found", http.StatusNotFound)
		return
	}

	updateDomainCount(longURL)

	log.Printf("Redirecting to: %s", longURL)
	http.Redirect(w, r, longURL, http.StatusFound)
}
func MetricHandler(w http.ResponseWriter, r *http.Request) {
	if len(domainVisitCounts) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	topDomains := getTopDomains()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topDomains)
}
func getTopDomains() []string {
	type domainCountPair struct {
		Domain string
		Count  int
	}
	var pairs []domainCountPair
	for domain, count := range domainVisitCounts {
		pairs = append(pairs, domainCountPair{domain, count})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Count > pairs[j].Count
	})
	topDomains := make([]string, 0, 3)
	for i := 0; i < 3 && i < len(pairs); i++ {
		topDomains = append(topDomains, fmt.Sprintf("%s: %d", pairs[i].Domain, pairs[i].Count))
	}
	return topDomains
}
func ShortHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}
	type RequestData struct {
		LongURL string `json:"longURL"`
	}
	var reqData RequestData
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqData)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	shortURL, err := Shorten(reqData.LongURL)
	if err != nil {
		http.Error(w, "Error shortening URL", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"shortURL": shortURL})
}
