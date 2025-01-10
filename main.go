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
