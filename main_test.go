package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var mock sqlmock.Sqlmock

func resetDomainCounts() {
	domainVisitCounts = make(map[string]int)
}
func setupMockDB() {
	var err error
	db, mock, err = sqlmock.New()
	if err != nil {
		fmt.Println("Failed to open mock database:", err)
		return
	}
}
func closeMockDB() {
	if db != nil {
		db.Close()
	}
}
