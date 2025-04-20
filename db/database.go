package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"bookshelf/models"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var DB *sql.DB

const dbFileName = "bookshelf.db"

// InitDB initializes the database connection and creates tables if they don't exist.
func InitDB() error {
	dbPath := dbFileName
	// Check if running in test, adjust path if necessary (simple check)
	if os.Getenv("GO_ENV") == "test" {
		// Potentially use an in-memory DB or a test-specific file
		dbPath = "test_bookshelf.db"
	}

	log.Printf("Initializing database: %s", dbPath)
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Check connection
	if err = DB.Ping(); err != nil {
		return err
	}

	// Create tables
	return createTables()
}

// createTables creates the necessary database tables if they don't already exist.
func createTables() error {
	createBooksTableSQL := `
    CREATE TABLE IF NOT EXISTS books (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        author TEXT,
        open_library_id TEXT,
        status TEXT NOT NULL CHECK(status IN ('Want to Read', 'Currently Reading', 'Read')),
        rating INTEGER CHECK(rating >= 1 AND rating <= 10),
        comments TEXT,
        cover_url TEXT
    );`

	_, err := DB.Exec(createBooksTableSQL)
	if err != nil {
		log.Printf("Error creating books table: %v", err)
		return err
	}
	log.Println("Books table checked/created successfully.")
	return nil
}

// GetAllBooks retrieves all books from the database.
func GetAllBooks() ([]models.Book, error) {
	rows, err := DB.Query("SELECT id, title, author, open_library_id, status, rating, comments, cover_url FROM books ORDER BY title ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := []models.Book{}
	for rows.Next() {
		var b models.Book
		// Use pointers for nullable fields when scanning
		err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.OpenLibraryID, &b.Status, &b.Rating, &b.Comments, &b.CoverURL)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

// CloseDB closes the database connection.
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed.")
	}
}
