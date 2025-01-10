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
func incrementRequestCount(longURL string) {
	parsedURL, err := url.Parse(longURL)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return
	}
	fmt.Println("Parsed domain:", parsedURL.Host)
	domainVisitCounts[parsedURL.Host]++
}
func TestShorten(t *testing.T) {
	setupMockDB()
	defer closeMockDB()
	mock.ExpectQuery("SELECT shortURL FROM UrlRecord WHERE longURL = ?").
		WithArgs("https://example.com").
		WillReturnRows(sqlmock.NewRows([]string{"shortURL"}).AddRow("abc123"))
	shortURL, err := Shorten("https://example.com")
	assert.NoError(t, err)
	assert.Equal(t, "abc123", shortURL)
	incrementRequestCount("https://example.com")

	mock.ExpectQuery("SELECT shortURL FROM UrlRecord WHERE longURL = ?").
		WithArgs("https://newsite.com").
		WillReturnRows(sqlmock.NewRows([]string{}))

	mock.ExpectQuery("SELECT shortURL FROM UrlRecord WHERE shortURL = ?").
		WillReturnRows(sqlmock.NewRows([]string{}))

	mock.ExpectExec("INSERT INTO UrlRecord").
		WithArgs("https://newsite.com", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	shortURL, err = Shorten("https://newsite.com")
	assert.NoError(t, err)
	assert.NotEmpty(t, shortURL)
	incrementRequestCount("https://newsite.com")
	fmt.Println("Domain Counts:", domainVisitCounts)
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
func TestRedirect(t *testing.T) {
	setupMockDB()
	defer closeMockDB()
	t.Run("Redirect Found", func(t *testing.T) {
		shortURL := "abc123"
		longURL := "https://example.com"
		mock.ExpectQuery("SELECT longURL FROM UrlRecord WHERE shortURL = ?").
			WithArgs(shortURL).
			WillReturnRows(sqlmock.NewRows([]string{"longURL"}).AddRow(longURL))

		mockRequest, err := http.NewRequest("GET", "/redirect/"+shortURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		response := httptest.NewRecorder()

		handler := http.HandlerFunc(Redirect)
		handler.ServeHTTP(response, mockRequest)
		assert.Equal(t, http.StatusFound, response.Code)
		assert.Equal(t, longURL, response.Header().Get("Location"))
		err = mock.ExpectationsWereMet()
		if err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
	t.Run("Redirect Not Found", func(t *testing.T) {
		shortURL := "xyz789"
		mock.ExpectQuery("SELECT longURL FROM UrlRecord WHERE shortURL = ?").
			WithArgs(shortURL).
			WillReturnError(sql.ErrNoRows)

		mockRequest, err := http.NewRequest("GET", "/redirect/"+shortURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		response := httptest.NewRecorder()
		handler := http.HandlerFunc(Redirect)
		handler.ServeHTTP(response, mockRequest)
		assert.Equal(t, http.StatusNotFound, response.Code)
		err = mock.ExpectationsWereMet()
		if err != nil {
			t.Errorf("There were unfulfilled expectations: %s", err)
		}
	})
}
func TestMetrics(t *testing.T) {
	resetDomainCounts()
	TestShorten(t)
	mockRequest, err := http.NewRequest("GET", "/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler := http.HandlerFunc(MetricHandler)
	handler.ServeHTTP(response, mockRequest)
	assert.Equal(t, http.StatusOK, response.Code)
	expectedResponse := `["example.com: 1", "newsite.com: 1"]`
	assert.JSONEq(t, expectedResponse, response.Body.String())
}
