package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetBooks(t *testing.T) {
	// Create a new router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/api/books", getBooks).Methods("GET")

	// Create a new request
	req, _ := http.NewRequest("GET", "/api/books", nil)
	
	// Record the response
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body format
	var books []Book
	err := json.Unmarshal(rr.Body.Bytes(), &books)
	if err != nil {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
	
	// Print the books for debugging
	fmt.Println("Books returned:", books)
}

func TestGetBook(t *testing.T) {
	// Initialize test data
	books = []Book{
		{ID: "1", Title: "Test Book", Author: "Test Author", Rating: 4.5},
	}

	// Create a new router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/api/books/{id}", getBook).Methods("GET")

	// Test for existing book
	req, _ := http.NewRequest("GET", "/api/books/1", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code for existing book: got %v want %v", status, http.StatusOK)
	}

	var book Book
	err := json.Unmarshal(rr.Body.Bytes(), &book)
	if err != nil {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}

	if book.ID != "1" || book.Title != "Test Book" {
		t.Errorf("handler returned unexpected book: got %v", book)
	}
	
	fmt.Println("Book returned:", book)

	// Test for non-existing book
	req, _ = http.NewRequest("GET", "/api/books/999", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code for non-existing book: got %v want %v", status, http.StatusNotFound)
	}
}

func TestCreateBook(t *testing.T) {
	// Initialize empty books slice
	books = []Book{}

	// Create a new router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/api/books", createBook).Methods("POST")

	// Create a new book
	newBook := Book{Title: "New Book", Author: "New Author", Rating: 4.0}
	jsonBook, _ := json.Marshal(newBook)

	// Create a request with the book in the body
	req, _ := http.NewRequest("POST", "/api/books", bytes.NewBuffer(jsonBook))
	req.Header.Set("Content-Type", "application/json")

	// Record the response
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Check that the book was added
	if len(books) != 1 {
		t.Errorf("book was not added to the slice: got len %v want %v", len(books), 1)
	}

	// Check the response body
	var responseBook Book
	err := json.Unmarshal(rr.Body.Bytes(), &responseBook)
	if err != nil {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}

	if responseBook.Title != "New Book" || responseBook.Author != "New Author" {
		t.Errorf("handler returned unexpected book: got %v", responseBook)
	}
	
	fmt.Println("Created book:", responseBook)
}

func TestUpdateBook(t *testing.T) {
	// Initialize test data
	books = []Book{
		{ID: "1", Title: "Original Title", Author: "Original Author", Rating: 3.0},
	}

	// Create a new router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/api/books/{id}", updateBook).Methods("PUT")

	// Create an updated book
	updatedBook := Book{Title: "Updated Title", Author: "Updated Author", Rating: 5.0}
	jsonBook, _ := json.Marshal(updatedBook)

	// Create a request with the book in the body
	req, _ := http.NewRequest("PUT", "/api/books/1", bytes.NewBuffer(jsonBook))
	req.Header.Set("Content-Type", "application/json")

	// Record the response
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the book was updated
	if books[0].Title != "Updated Title" || books[0].Author != "Updated Author" || books[0].Rating != 5.0 {
		t.Errorf("book was not updated correctly: got %v", books[0])
	}
	
	fmt.Println("Updated book:", books[0])
}

func TestDeleteBook(t *testing.T) {
	// Initialize test data
	books = []Book{
		{ID: "1", Title: "Test Book", Author: "Test Author", Rating: 4.5},
	}

	// Create a new router and register the handler
	router := mux.NewRouter()
	router.HandleFunc("/api/books/{id}", deleteBook).Methods("DELETE")

	// Create a request
	req, _ := http.NewRequest("DELETE", "/api/books/1", nil)

	// Record the response
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}

	// Check that the book was deleted
	if len(books) != 0 {
		t.Errorf("book was not deleted from the slice: got len %v want %v", len(books), 0)
	}
	
	fmt.Println("Books after delete:", books)
}