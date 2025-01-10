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
