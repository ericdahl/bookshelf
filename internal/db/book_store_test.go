package db

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ericdahl/bookshelf/internal/model"
)

// setupTestDB creates a new in-memory SQLite database for testing
func setupTestDB(t *testing.T) (*sql.DB, *SQLiteBookStore) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Create the schema
	err = createSchema(db)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	store := NewSQLiteBookStore(db)
	return db, store
}

// teardownTestDB closes the database connection
func teardownTestDB(db *sql.DB) {
	db.Close()
}

// createTestBook returns a sample book for testing
func createTestBook() *model.Book {
	comments := "Test comments"
	rating := 8
	coverURL := "http://example.com/cover.jpg"
	return &model.Book{
		Title:         "Test Book",
		Author:        "Test Author",
		OpenLibraryID: "OL12345M",
		ISBN:          "9781234567890",
		Status:        model.StatusWantToRead,
		Rating:        &rating,
		Comments:      &comments,
		CoverURL:      &coverURL,
	}
}

// TestAddBook tests adding a book to the database
func TestAddBook(t *testing.T) {
	db, store := setupTestDB(t)
	defer teardownTestDB(db)

	book := createTestBook()
	id, err := store.AddBook(book)
	if err != nil {
		t.Fatalf("AddBook failed: %v", err)
	}

	if id <= 0 {
		t.Errorf("Expected positive ID, got %d", id)
	}

	if book.ID != id {
		t.Errorf("Book ID not updated, expected %d, got %d", id, book.ID)
	}

	// Test adding book with invalid status
	invalidBook := createTestBook()
	invalidBook.Status = "Invalid Status"
	_, err = store.AddBook(invalidBook)
	if err == nil {
		t.Errorf("Expected error when adding book with invalid status")
	}

	// Test adding book with invalid rating
	invalidRating := 11
	invalidBook = createTestBook()
	invalidBook.OpenLibraryID = "OL67890M" // Different ID to avoid uniqueness constraint
	invalidBook.Rating = &invalidRating
	_, err = store.AddBook(invalidBook)
	if err == nil {
		t.Errorf("Expected error when adding book with invalid rating")
	}

	// Test uniqueness constraint
	duplicateBook := createTestBook()
	_, err = store.AddBook(duplicateBook)
	if err == nil {
		t.Errorf("Expected error when adding book with duplicate OpenLibraryID")
	}
}

// TestGetBooks tests retrieving all books from the database
func TestGetBooks(t *testing.T) {
	db, store := setupTestDB(t)
	defer teardownTestDB(db)

	// Add test books
	book1 := createTestBook()
	_, err := store.AddBook(book1)
	if err != nil {
		t.Fatalf("Failed to add test book 1: %v", err)
	}

	book2 := createTestBook()
	book2.Title = "Test Book 2"
	book2.OpenLibraryID = "OL67890M"
	book2.Status = model.StatusCurrentlyReading
	_, err = store.AddBook(book2)
	if err != nil {
		t.Fatalf("Failed to add test book 2: %v", err)
	}

	// Test GetBooks
	books, err := store.GetBooks()
	if err != nil {
		t.Fatalf("GetBooks failed: %v", err)
	}

	if len(books) != 2 {
		t.Errorf("Expected 2 books, got %d", len(books))
	}
}

// TestGetBookByID tests retrieving a specific book by ID
func TestGetBookByID(t *testing.T) {
	db, store := setupTestDB(t)
	defer teardownTestDB(db)

	// Add a test book
	book := createTestBook()
	id, err := store.AddBook(book)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	// Test getting the book by ID
	retrievedBook, err := store.GetBookByID(id)
	if err != nil {
		t.Fatalf("GetBookByID failed: %v", err)
	}

	if retrievedBook.ID != id {
		t.Errorf("Expected ID %d, got %d", id, retrievedBook.ID)
	}
	if retrievedBook.Title != book.Title {
		t.Errorf("Expected title %s, got %s", book.Title, retrievedBook.Title)
	}
	if retrievedBook.Author != book.Author {
		t.Errorf("Expected author %s, got %s", book.Author, retrievedBook.Author)
	}
	if retrievedBook.Status != book.Status {
		t.Errorf("Expected status %s, got %s", book.Status, retrievedBook.Status)
	}
	if retrievedBook.OpenLibraryID != book.OpenLibraryID {
		t.Errorf("Expected OpenLibraryID %s, got %s", book.OpenLibraryID, retrievedBook.OpenLibraryID)
	}
	if retrievedBook.ISBN != book.ISBN {
		t.Errorf("Expected ISBN %s, got %s", book.ISBN, retrievedBook.ISBN)
	}
	if !reflect.DeepEqual(retrievedBook.Rating, book.Rating) {
		t.Errorf("Expected rating %v, got %v", *book.Rating, *retrievedBook.Rating)
	}
	if !reflect.DeepEqual(retrievedBook.Comments, book.Comments) {
		t.Errorf("Expected comments %v, got %v", *book.Comments, *retrievedBook.Comments)
	}
	if !reflect.DeepEqual(retrievedBook.CoverURL, book.CoverURL) {
		t.Errorf("Expected cover URL %v, got %v", *book.CoverURL, *retrievedBook.CoverURL)
	}

	// Test getting non-existent book
	_, err = store.GetBookByID(999)
	if err == nil {
		t.Errorf("Expected error when getting non-existent book")
	}
}

// TestUpdateBookStatus tests updating a book's status
func TestUpdateBookStatus(t *testing.T) {
	db, store := setupTestDB(t)
	defer teardownTestDB(db)

	// Add a test book
	book := createTestBook()
	id, err := store.AddBook(book)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	// Test updating status
	err = store.UpdateBookStatus(id, model.StatusCurrentlyReading)
	if err != nil {
		t.Fatalf("UpdateBookStatus failed: %v", err)
	}

	// Verify the update
	updatedBook, err := store.GetBookByID(id)
	if err != nil {
		t.Fatalf("Failed to get book after update: %v", err)
	}

	if updatedBook.Status != model.StatusCurrentlyReading {
		t.Errorf("Expected status %s, got %s", model.StatusCurrentlyReading, updatedBook.Status)
	}

	// Test updating with invalid status
	err = store.UpdateBookStatus(id, "Invalid Status")
	if err == nil {
		t.Errorf("Expected error when updating with invalid status")
	}

	// Test updating non-existent book
	err = store.UpdateBookStatus(999, model.StatusRead)
	if err == nil {
		t.Errorf("Expected error when updating non-existent book")
	}
}

// TestUpdateBookDetails tests updating a book's rating and comments
func TestUpdateBookDetails(t *testing.T) {
	db, store := setupTestDB(t)
	defer teardownTestDB(db)

	// Add a test book
	book := createTestBook()
	id, err := store.AddBook(book)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	// Test updating details
	newRating := 10
	newComments := "Updated comments"
	err = store.UpdateBookDetails(id, &newRating, &newComments)
	if err != nil {
		t.Fatalf("UpdateBookDetails failed: %v", err)
	}

	// Verify the update
	updatedBook, err := store.GetBookByID(id)
	if err != nil {
		t.Fatalf("Failed to get book after update: %v", err)
	}

	if *updatedBook.Rating != newRating {
		t.Errorf("Expected rating %d, got %d", newRating, *updatedBook.Rating)
	}
	if *updatedBook.Comments != newComments {
		t.Errorf("Expected comments %s, got %s", newComments, *updatedBook.Comments)
	}

	// Test clearing details (setting to null)
	err = store.UpdateBookDetails(id, nil, nil)
	if err != nil {
		t.Fatalf("UpdateBookDetails with nil values failed: %v", err)
	}

	// Verify nulls were set
	updatedBook, err = store.GetBookByID(id)
	if err != nil {
		t.Fatalf("Failed to get book after update: %v", err)
	}

	if updatedBook.Rating != nil {
		t.Errorf("Expected nil rating, got %v", *updatedBook.Rating)
	}
	if updatedBook.Comments != nil {
		t.Errorf("Expected nil comments, got %v", *updatedBook.Comments)
	}

	// Test with invalid rating
	invalidRating := 11
	err = store.UpdateBookDetails(id, &invalidRating, nil)
	if err == nil {
		t.Errorf("Expected error when updating with invalid rating")
	}

	// Test updating non-existent book
	err = store.UpdateBookDetails(999, &newRating, &newComments)
	if err == nil {
		t.Errorf("Expected error when updating non-existent book")
	}
}

// TestDeleteBook tests deleting a book from the database
func TestDeleteBook(t *testing.T) {
	db, store := setupTestDB(t)
	defer teardownTestDB(db)

	// Add a test book
	book := createTestBook()
	id, err := store.AddBook(book)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	// Test deleting the book
	err = store.DeleteBook(id)
	if err != nil {
		t.Fatalf("DeleteBook failed: %v", err)
	}

	// Verify the book was deleted
	_, err = store.GetBookByID(id)
	if err == nil {
		t.Errorf("Expected error when getting deleted book")
	}

	// Test deleting non-existent book
	err = store.DeleteBook(999)
	if err == nil {
		t.Errorf("Expected error when deleting non-existent book")
	}
}