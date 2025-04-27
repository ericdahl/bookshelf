package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/your-username/bookshelf/internal/model" // Adjust import path
)

// BookStore defines the interface for database operations on books.
type BookStore interface {
	AddBook(book *model.Book) (int64, error)
	GetBooks() ([]model.Book, error)
	GetBookByID(id int64) (*model.Book, error)
	UpdateBookStatus(id int64, status model.BookStatus) error
	UpdateBookDetails(id int64, rating *int, comments *string) error
	DeleteBook(id int64) error
	// DeleteBook(id int64) error // Future enhancement
}

// SQLiteBookStore implements the BookStore interface using SQLite.
type SQLiteBookStore struct {
	DB *sql.DB
}

// NewSQLiteBookStore creates a new SQLiteBookStore.
func NewSQLiteBookStore(db *sql.DB) *SQLiteBookStore {
	return &SQLiteBookStore{DB: db}
}

// AddBook inserts a new book into the database.
// It sets the book's ID after successful insertion.
func (s *SQLiteBookStore) AddBook(book *model.Book) (int64, error) {
	// Default status if not provided (though handler should ensure it)
	if book.Status == "" {
		book.Status = model.StatusWantToRead // Or Currently Reading as per initial request? Let's stick to Want to Read for now.
	} else if !book.Status.IsValid() {
		return 0, fmt.Errorf("invalid status: %s", book.Status)
	}

	if err := book.Validate(); err != nil {
		return 0, fmt.Errorf("validation failed: %w", err)
	}

	query := `
        INSERT INTO books (title, author, open_library_id, isbn, status, rating, comments, cover_url)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?);
    `
	log.Printf("SQL: Executing AddBook query: %s with params: title='%s', author='%s', open_library_id='%s', isbn='%s', status='%s', rating=%v, comments='%v', cover_url='%v'",
		query, book.Title, book.Author, book.OpenLibraryID, book.ISBN, book.Status, book.Rating, book.Comments, book.CoverURL)
	stmt, err := s.DB.Prepare(query)
	if err != nil {
		log.Printf("SQL Error: Preparing AddBook statement failed: %v", err)
		return 0, fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(book.Title, book.Author, book.OpenLibraryID, book.ISBN, book.Status, book.Rating, book.Comments, book.CoverURL)
	if err != nil {
		log.Printf("SQL Error: Executing AddBook statement failed: %v", err)
		// Consider checking for UNIQUE constraint violation specifically
		return 0, fmt.Errorf("failed to execute insert statement: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Printf("SQL Error: Failed to get last insert ID: %v", err)
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}
	book.ID = id // Set the ID on the original struct
	log.Printf("SQL: Successfully added book with ID: %d", id)
	return id, nil
}

// GetBooks retrieves all books from the database.
func (s *SQLiteBookStore) GetBooks() ([]model.Book, error) {
	query := `SELECT id, title, author, open_library_id, isbn, status, rating, comments, cover_url FROM books ORDER BY title;`
	log.Printf("SQL: Executing GetBooks query: %s", query)

	rows, err := s.DB.Query(query)
	if err != nil {
		log.Printf("SQL Error: Executing GetBooks query failed: %v", err)
		return nil, fmt.Errorf("failed to query books: %w", err)
	}
	defer rows.Close()

	books := []model.Book{}
	for rows.Next() {
		var book model.Book
		// Ensure pointers are used for nullable fields (rating, comments, cover_url, isbn)
		var rating sql.NullInt64
		var comments sql.NullString
		var coverURL sql.NullString
		var isbn sql.NullString

		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.OpenLibraryID, &isbn, &book.Status, &rating, &comments, &coverURL); err != nil {
			log.Printf("SQL Error: Scanning book row failed: %v", err)
			return nil, fmt.Errorf("failed to scan book row: %w", err)
		}

		// Convert sql.Null types to pointers
		if isbn.Valid {
			book.ISBN = isbn.String
		}
		if rating.Valid {
			r := int(rating.Int64)
			book.Rating = &r
		}
		if comments.Valid {
			book.Comments = &comments.String
		}
		if coverURL.Valid {
			book.CoverURL = &coverURL.String
		}

		books = append(books, book)
	}

	if err = rows.Err(); err != nil {
		log.Printf("SQL Error: Error during row iteration: %v", err)
		return nil, fmt.Errorf("error iterating book rows: %w", err)
	}

	log.Printf("SQL: Retrieved %d books", len(books))
	return books, nil
}

// GetBookByID retrieves a single book by its ID.
func (s *SQLiteBookStore) GetBookByID(id int64) (*model.Book, error) {
	query := `SELECT id, title, author, open_library_id, isbn, status, rating, comments, cover_url FROM books WHERE id = ?;`
	log.Printf("SQL: Executing GetBookByID query: %s with id=%d", query, id)

	row := s.DB.QueryRow(query, id)

	var book model.Book
	var rating sql.NullInt64
	var comments sql.NullString
	var coverURL sql.NullString
	var isbn sql.NullString

	err := row.Scan(&book.ID, &book.Title, &book.Author, &book.OpenLibraryID, &isbn, &book.Status, &rating, &comments, &coverURL)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("SQL: No book found for ID: %d", id)
			return nil, fmt.Errorf("book with ID %d not found", id) // Consider a specific error type (e.g., ErrNotFound)
		}
		log.Printf("SQL Error: Scanning book row failed for ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to scan book row for ID %d: %w", id, err)
	}

	// Convert sql.Null types to pointers
	if isbn.Valid {
		book.ISBN = isbn.String
	}
	if rating.Valid {
		r := int(rating.Int64)
		book.Rating = &r
	}
	if comments.Valid {
		book.Comments = &comments.String
	}
	if coverURL.Valid {
		book.CoverURL = &coverURL.String
	}

	log.Printf("SQL: Retrieved book with ID: %d", id)
	return &book, nil
}

// UpdateBookStatus updates the status of a specific book.
func (s *SQLiteBookStore) UpdateBookStatus(id int64, status model.BookStatus) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid status provided: %s", status)
	}

	query := `UPDATE books SET status = ? WHERE id = ?;`
	log.Printf("SQL: Executing UpdateBookStatus query: %s with status='%s', id=%d", query, status, id)

	stmt, err := s.DB.Prepare(query)
	if err != nil {
		log.Printf("SQL Error: Preparing UpdateBookStatus statement failed: %v", err)
		return fmt.Errorf("failed to prepare update status statement: %w", err)
	}
	defer stmt.Close()

	res, err := stmt.Exec(status, id)
	if err != nil {
		log.Printf("SQL Error: Executing UpdateBookStatus statement failed: %v", err)
		return fmt.Errorf("failed to execute update status statement: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("SQL Error: Failed to get rows affected for UpdateBookStatus: %v", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Printf("SQL: No book found to update status for ID: %d", id)
		return fmt.Errorf("book with ID %d not found", id) // Consider ErrNotFound
	}

	log.Printf("SQL: Successfully updated status for book ID: %d", id)
	return nil
}

// UpdateBookDetails updates the rating and/or comments of a specific book.
// It handles NULL values correctly.
func (s *SQLiteBookStore) UpdateBookDetails(id int64, rating *int, comments *string) error {
	// Validate rating if provided
	if rating != nil && (*rating < 1 || *rating > 10) {
		return fmt.Errorf("rating must be between 1 and 10")
	}

	query := `UPDATE books SET rating = ?, comments = ? WHERE id = ?;`
	log.Printf("SQL: Executing UpdateBookDetails query: %s with rating=%v, comments='%v', id=%d", query, rating, comments, id)

	stmt, err := s.DB.Prepare(query)
	if err != nil {
		log.Printf("SQL Error: Preparing UpdateBookDetails statement failed: %v", err)
		return fmt.Errorf("failed to prepare update details statement: %w", err)
	}
	defer stmt.Close()

	// Handle potential nil values for rating and comments when passing to Exec
	var sqlRating interface{}
	if rating != nil {
		sqlRating = *rating
	} else {
		sqlRating = nil // This will be translated to NULL by the driver
	}

	var sqlComments interface{}
	if comments != nil {
		sqlComments = *comments
	} else {
		sqlComments = nil // This will be translated to NULL by the driver
	}

	res, err := stmt.Exec(sqlRating, sqlComments, id)
	if err != nil {
		log.Printf("SQL Error: Executing UpdateBookDetails statement failed: %v", err)
		return fmt.Errorf("failed to execute update details statement: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Printf("SQL Error: Failed to get rows affected for UpdateBookDetails: %v", err)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Printf("SQL: No book found to update details for ID: %d", id)
		return fmt.Errorf("book with ID %d not found", id) // Consider ErrNotFound
	}

	log.Printf("SQL: Successfully updated details for book ID: %d", id)
	return nil
}

// DeleteBook removes a book from the database by its ID.
func (s *SQLiteBookStore) DeleteBook(id int64) error {
	query := `DELETE FROM books WHERE id = ?;`

	result, err := s.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("book with ID %d not found", id)
	}

	return nil
}
