package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// InitDB initializes the SQLite database connection and creates the necessary tables if they don't exist.
func InitDB(dataSourceName string) (*sql.DB, error) {
	// Ensure the directory for the database file exists
	dir := filepath.Dir(dataSourceName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("Creating database directory: %s", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory %s: %w", dir, err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to check database directory %s: %w", dir, err)
	}


	log.Printf("Initializing database connection to: %s", dataSourceName)
	db, err := sql.Open("sqlite3", dataSourceName+"?_foreign_keys=on") // Enable foreign key support if needed later
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Check the connection
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection successful.")

	// Create tables if they don't exist
	if err = CreateSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create database schema: %w", err)
	}

	log.Println("Database schema verified/created.")
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
        cover_url TEXT
    );
    `
	log.Println("Executing schema creation SQL...")
	_, err := db.Exec(schema)
	if err != nil {
		log.Printf("Error executing schema SQL: %v", err)
		return fmt.Errorf("failed to execute schema creation: %w", err)
	}
	log.Println("Schema execution successful.")
	return nil
}
