package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ericdahl/bookshelf/internal/api"
	"github.com/ericdahl/bookshelf/internal/db"
)

func TestSetupRouter(t *testing.T) {
	// Create temporary directory for web assets
	tempDir, err := os.MkdirTemp("", "bookshelf-web-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test database in memory
	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Create book store and API handler
	bookStore := db.NewSQLiteBookStore(database)
	apiHandler := api.NewAPIHandler(bookStore)

	// Set up router
	router := api.SetupRouter(apiHandler, tempDir)
	if router == nil {
		t.Fatal("SetupRouter returned nil router")
	}

	// Test API endpoint routing
	server := httptest.NewServer(router)
	defer server.Close()

	// Make a simple request to test router is working
	resp, err := http.Get(server.URL + "/api/books")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}
}