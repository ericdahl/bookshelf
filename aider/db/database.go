package db

import (
	"database/sql"
	"log"
	"os"
	// "path/filepath" // Removed as it was unused

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
        isbn TEXT NOT NULL, -- Added ISBN field, required
        status TEXT NOT NULL CHECK(status IN ('Want to Read', 'Currently Reading', 'Read')),
        rating INTEGER CHECK(rating >= 1 AND rating <= 10),
        comments TEXT,
        cover_url TEXT
    );`

	log.Println("Executing SQL:", createBooksTableSQL) // Log SQL
	_, err := DB.Exec(createBooksTableSQL)
	if err != nil {
		log.Printf("Error creating books table: %v", err)
		return err
	}
	log.Println("Books table checked/created successfully.")
	return nil
}

// GetAllBooks retrieves all books from the database, ordered by title.
func GetAllBooks() ([]models.Book, error) {
	query := "SELECT id, title, author, open_library_id, isbn, status, rating, comments, cover_url FROM books ORDER BY title ASC"
	log.Println("Executing SQL:", query) // Log SQL
	rows, err := DB.Query(query)
	if err != nil {
		log.Printf("Error executing query '%s': %v", query, err)
		return nil, err
	}
	defer rows.Close()

	books := []models.Book{}
	for rows.Next() {
		var b models.Book
		// Use pointers for nullable fields (Rating, Comments, CoverURL) when scanning
		err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.OpenLibraryID, &b.ISBN, &b.Status, &b.Rating, &b.Comments, &b.CoverURL)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}
		books = append(books, b)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}

// AddBook inserts a new book into the database. Requires Title and ISBN.
func AddBook(book models.Book) (int64, error) {
	// Basic validation at DB layer (could also be done in handler)
	if book.Title == "" || book.ISBN == "" {
		log.Printf("Attempted to add book with missing title or ISBN: Title='%s', ISBN='%s'", book.Title, book.ISBN)
		return 0, fmt.Errorf("book title and ISBN are required")
	}

	query := "INSERT INTO books(title, author, open_library_id, isbn, status, rating, comments, cover_url) VALUES(?, ?, ?, ?, ?, ?, ?, ?)"
	log.Println("Preparing SQL:", query) // Log SQL Prepare
	stmt, err := DB.Prepare(query)
	if err != nil {
		log.Printf("Error preparing query '%s': %v", query, err)
		return 0, err
	}
	defer stmt.Close()

	// Use default status if not provided
	if book.Status == "" {
		book.Status = models.StatusWantToRead
	}

	log.Printf("Executing SQL Insert with params: Title=%s, Author=%s, OLID=%s, ISBN=%s, Status=%s, Rating=%v, Comments=%v, CoverURL=%v",
		book.Title, book.Author, book.OpenLibraryID, book.ISBN, book.Status, book.Rating, book.Comments, book.CoverURL) // Log Params

	res, err := stmt.Exec(book.Title, book.Author, book.OpenLibraryID, book.ISBN, book.Status, book.Rating, book.Comments, book.CoverURL)
	if err != nil {
		log.Printf("Error executing insert: %v", err)
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	log.Printf("Added book with ID: %d", id)
	return id, nil
}

import (
	"database/sql"
	"fmt" // Import fmt for error formatting
	"log"
	"os"
	// "path/filepath" // Removed as it was unused
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		log.Printf("No book found with ID %d to update status", id)
		// Consider returning a specific error like sql.ErrNoRows here if needed
		return sql.ErrNoRows // Or a custom error
	}

	log.Printf("Updated status for book ID %d to %s", id, status)
	return nil
}


// CloseDB closes the database connection.
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed.")
	}
}
