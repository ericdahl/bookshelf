package main

import (
	"os"
	"testing"
)

func TestDatabase(t *testing.T) {
	// Use an in-memory database for testing
	dbFile := "./test.db"
	InitDB(dbFile)
	defer func() {
		DB.Close()
		os.Remove(dbFile) // Clean up after test
	}()

	// Test adding a book
	testBook := DbBook{
		ID:     generateID(),
		Title:  "Test Database Book",
		Author: "Test Author",
		Rating: 4.8,
	}

	err := AddBook(testBook)
	if err != nil {
		t.Fatalf("Failed to add book: %v", err)
	}

	// Test getting all books
	books, err := GetAllBooks()
	if err != nil {
		t.Fatalf("Failed to get books: %v", err)
	}
	if len(books) != 1 {
		t.Errorf("Expected 1 book, got %d", len(books))
	}
	if books[0].Title != "Test Database Book" {
		t.Errorf("Expected book title 'Test Database Book', got '%s'", books[0].Title)
	}

	// Test getting a specific book
	book, err := GetBookByID(testBook.ID)
	if err != nil {
		t.Fatalf("Failed to get book by ID: %v", err)
	}
	if book.Title != "Test Database Book" {
		t.Errorf("Expected book title 'Test Database Book', got '%s'", book.Title)
	}

	// Test updating a book
	book.Title = "Updated Test Book"
	err = UpdateBook(book)
	if err != nil {
		t.Fatalf("Failed to update book: %v", err)
	}

	// Verify update
	updatedBook, err := GetBookByID(testBook.ID)
	if err != nil {
		t.Fatalf("Failed to get updated book: %v", err)
	}
	if updatedBook.Title != "Updated Test Book" {
		t.Errorf("Expected updated book title 'Updated Test Book', got '%s'", updatedBook.Title)
	}

	// Test shelves
	shelves, err := GetShelves()
	if err != nil {
		t.Fatalf("Failed to get shelves: %v", err)
	}
	if len(shelves) != 3 { // Three predefined shelves
		t.Errorf("Expected 3 shelves, got %d", len(shelves))
	}

	// Test adding book to shelf
	if len(shelves) > 0 {
		err = AddBookToShelf(testBook.ID, shelves[0].ID)
		if err != nil {
			t.Fatalf("Failed to add book to shelf: %v", err)
		}

		// Test getting books in shelf
		shelfBooks, err := GetBooksInShelf(shelves[0].ID)
		if err != nil {
			t.Fatalf("Failed to get books in shelf: %v", err)
		}
		if len(shelfBooks) != 1 {
			t.Errorf("Expected 1 book in shelf, got %d", len(shelfBooks))
		}

		// Test removing book from shelf
		err = RemoveBookFromShelf(testBook.ID, shelves[0].ID)
		if err != nil {
			t.Fatalf("Failed to remove book from shelf: %v", err)
		}

		// Verify removal
		emptyShelfBooks, err := GetBooksInShelf(shelves[0].ID)
		if err != nil {
			t.Fatalf("Failed to get books in shelf after removal: %v", err)
		}
		if len(emptyShelfBooks) != 0 {
			t.Errorf("Expected 0 books in shelf after removal, got %d", len(emptyShelfBooks))
		}
	}

	// Test deleting a book
	err = DeleteBook(testBook.ID)
	if err != nil {
		t.Fatalf("Failed to delete book: %v", err)
	}

	// Verify deletion
	_, err = GetBookByID(testBook.ID)
	if err == nil {
		t.Errorf("Expected error when getting deleted book, got nil")
	}
}