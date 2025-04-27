package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// InitDB initializes the SQLite database connection and creates the necessary tables if they don't exist.
func InitDB(dataSourceName string) (*sql.DB, error) {
	// Ensure the directory for the database file exists
	dir := filepath.Dir(dataSourceName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		slog.Info("Creating database directory", "dir", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to check database directory %s: %w", dir, err)
	}


	slog.Info("Initializing database connection", "dataSourceName", dataSourceName)
	db, err := sql.Open("sqlite3", dataSourceName+"?_foreign_keys=on") // Enable foreign key support if needed later
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Check the connection
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("Database connection successful")

	// Create tables if they don't exist
	if err = CreateSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create database schema: %w", err)
	}

	slog.Info("Database schema verified/created")
	return db, nil
}

// CreateSchema defines and executes the SQL statements to create the database tables.
// Exported for testing purposes.
func CreateSchema(db *sql.DB) error {
	// Use TEXT for status, INTEGER for rating (nullable), TEXT for comments (nullable)
	// Use TEXT for OpenLibraryID and ISBN
	// Add UNIQUE constraint on OpenLibraryID to prevent duplicates? Or handle in application logic.
	schema := `
    CREATE TABLE IF NOT EXISTS books (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        author TEXT NOT NULL,
        open_library_id TEXT NOT NULL UNIQUE,
        isbn TEXT,
        status TEXT NOT NULL CHECK(status IN ('Want to Read', 'Currently Reading', 'Read')),
        rating INTEGER CHECK(rating IS NULL OR (rating >= 1 AND rating <= 10)),
        comments TEXT,
        cover_url TEXT,
        series TEXT,
        series_index INTEGER
    );
    `
	slog.Info("Executing schema creation SQL")
	_, err := db.Exec(schema)
	if err != nil {
		slog.Error("Error executing schema SQL", "error", err)
		return fmt.Errorf("failed to execute schema creation: %w", err)
	}
	slog.Info("Schema execution successful")
	return nil
}