package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB is the database connection
var DB *sql.DB

// InitDB initializes the database connection
func InitDB(dataSourceName string) {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	// Create tables if they don't exist
	createTables()
}

// LogQuery logs SQL execution for debugging and monitoring
func LogQuery(query string, args ...interface{}) {
	log.Printf("Executing query: %s with args: %v", query, args)
}

// createTables creates the necessary tables
func createTables() {
	// Books table
	createBooksTable := `
	CREATE TABLE IF NOT EXISTS books (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		author TEXT NOT NULL,
		rating REAL,
		created_at TIMESTAMP,
		updated_at TIMESTAMP
	);`

	// Shelves table
	createShelvesTable := `
	CREATE TABLE IF NOT EXISTS shelves (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at TIMESTAMP
	);`

	// Books to shelves relationship
	createBookShelvesTable := `
	CREATE TABLE IF NOT EXISTS book_shelves (
		book_id TEXT,
		shelf_id TEXT,
		added_at TIMESTAMP,
		PRIMARY KEY (book_id, shelf_id),
		FOREIGN KEY (book_id) REFERENCES books (id),
		FOREIGN KEY (shelf_id) REFERENCES shelves (id)
	);`

	LogQuery(createBooksTable)
	_, err := DB.Exec(createBooksTable)
	if err != nil {
		log.Fatal(err)
	}

	LogQuery(createShelvesTable)
	_, err = DB.Exec(createShelvesTable)
	if err != nil {
		log.Fatal(err)
	}

	LogQuery(createBookShelvesTable)
	_, err = DB.Exec(createBookShelvesTable)
	if err != nil {
		log.Fatal(err)
	}

	// Create predefined shelves
	createPredefinedShelves()
}

// createPredefinedShelves adds default shelves
func createPredefinedShelves() {
	predefinedShelves := []string{"Currently Reading", "Want to Read", "Read"}

	for _, shelfName := range predefinedShelves {
		// Check if shelf already exists
		var count int
		err := DB.QueryRow("SELECT COUNT(*) FROM shelves WHERE name = ?", shelfName).Scan(&count)
		if err != nil {
			log.Printf("Error checking for shelf %s: %v", shelfName, err)
			continue
		}

		if count == 0 {
			// Create the shelf
			query := "INSERT INTO shelves (id, name, created_at) VALUES (?, ?, ?)"
			LogQuery(query, generateID(), shelfName, time.Now())
			_, err := DB.Exec(query, generateID(), shelfName, time.Now())
			if err != nil {
				log.Printf("Error creating shelf %s: %v", shelfName, err)
			}
		}
	}
}
