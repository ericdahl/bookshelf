package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/ericdahl/bookshelf/internal/db"
	"github.com/ericdahl/bookshelf/internal/model"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var (
	testDB      *sql.DB
	testStore   *db.SQLiteBookStore
	testHandler *APIHandler
	testRouter  *mux.Router
)

// Helper function to convert int64 to string
func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

// setupTestAPI sets up a test database and API handler for testing
func setupTestAPI() error {
	// Create an in-memory SQLite database for testing
	var err error
	testDB, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		return err
	}

	// Initialize the schema
	err = db.CreateSchema(testDB) // Make sure this is exported in db package
	if err != nil {
		return err
	}

	// Create the test store and handler
	testStore = db.NewSQLiteBookStore(testDB)
	testHandler = NewAPIHandler(testStore)

	// Set up the router
	testRouter = mux.NewRouter()
	testRouter.HandleFunc("/api/books", testHandler.GetBooksHandler).Methods(http.MethodGet)
	testRouter.HandleFunc("/api/books", testHandler.AddBookHandler).Methods(http.MethodPost)
	testRouter.HandleFunc("/api/books/{id:[0-9]+}", testHandler.UpdateBookStatusHandler).Methods(http.MethodPut)
	testRouter.HandleFunc("/api/books/{id:[0-9]+}/details", testHandler.UpdateBookDetailsHandler).Methods(http.MethodPut)
	testRouter.HandleFunc("/api/books/{id:[0-9]+}", testHandler.DeleteBookHandler).Methods(http.MethodDelete)
	testRouter.HandleFunc("/api/books/search", testHandler.SearchBooksHandler).Methods(http.MethodGet)

	return nil
}

// teardownTestAPI cleans up after tests
func teardownTestAPI() {
	if testDB != nil {
		testDB.Close()
	}
}

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Set up
	if err := setupTestAPI(); err != nil {
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Tear down
	teardownTestAPI()

	os.Exit(code)
}

// createTestBook returns a sample book for testing
func createTestBook(status model.BookStatus, suffix string) *model.Book {
	comments := "Test comments"
	rating := 8
	coverURL := "http://example.com/cover.jpg"
	return &model.Book{
		Title:         "Test Book " + suffix,
		Author:        "Test Author " + suffix,
		OpenLibraryID: "OL12345M" + suffix,
		ISBN:          "9781234567890",
		Status:        status,
		Rating:        &rating,
		Comments:      &comments,
		CoverURL:      &coverURL,
	}
}

// TestGetBooksHandler tests the GET /api/books endpoint
func TestGetBooksHandler(t *testing.T) {
	// Add test books
	book1 := createTestBook(model.StatusWantToRead, "1")
	_, err := testStore.AddBook(book1)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/api/books", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	testRouter.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var books []model.Book
	if err := json.Unmarshal(rr.Body.Bytes(), &books); err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	if len(books) != 1 {
		t.Errorf("Expected 1 book, got %d", len(books))
	}

	if books[0].Title != book1.Title {
		t.Errorf("Expected book title %s, got %s", book1.Title, books[0].Title)
	}
}

// TestAddBookHandler tests the POST /api/books endpoint
func TestAddBookHandler(t *testing.T) {
	// Create a test book
	book := createTestBook(model.StatusWantToRead, "2")

	// Convert to JSON for request body
	jsonData, err := json.Marshal(book)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Create the request
	req, err := http.NewRequest("POST", "/api/books", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	testRouter.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v, body: %s", status, http.StatusCreated, rr.Body.String())
	}

	// Check the response body
	var addedBook model.Book
	if err := json.Unmarshal(rr.Body.Bytes(), &addedBook); err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	if addedBook.Title != book.Title {
		t.Errorf("Expected book title %s, got %s", book.Title, addedBook.Title)
	}

	if addedBook.ID <= 0 {
		t.Errorf("Expected positive book ID, got %d", addedBook.ID)
	}
}

// TestUpdateBookStatusHandler tests the PUT /api/books/{id} endpoint
func TestUpdateBookStatusHandler(t *testing.T) {
	// Add test book
	book := createTestBook(model.StatusWantToRead, "3")
	id, err := testStore.AddBook(book)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	// Create update payload
	updateData := map[string]string{
		"status": string(model.StatusCurrentlyReading),
	}
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Create request
	req, err := http.NewRequest("PUT", "/api/books/"+itoa(id), bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder
	rr := httptest.NewRecorder()

	// Call the handler
	testRouter.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v, body: %s", status, http.StatusOK, rr.Body.String())
	}

	// Verify the status was updated in the database
	updatedBook, err := testStore.GetBookByID(id)
	if err != nil {
		t.Fatalf("Failed to retrieve updated book: %v", err)
	}

	if updatedBook.Status != model.StatusCurrentlyReading {
		t.Errorf("Book status not updated, expected %s, got %s", model.StatusCurrentlyReading, updatedBook.Status)
	}
}

// TestUpdateBookDetailsHandler tests the PUT /api/books/{id}/details endpoint
func TestUpdateBookDetailsHandler(t *testing.T) {
	// Add test book
	book := createTestBook(model.StatusWantToRead, "4")
	id, err := testStore.AddBook(book)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	// Create update payload
	newRating := 10
	newComments := "Updated test comments"
	updateData := map[string]interface{}{
		"rating":   newRating,
		"comments": newComments,
	}
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Create request
	req, err := http.NewRequest("PUT", "/api/books/"+itoa(id)+"/details", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder
	rr := httptest.NewRecorder()

	// Call the handler
	testRouter.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v, body: %s", status, http.StatusOK, rr.Body.String())
	}

	// Verify the details were updated in the database
	updatedBook, err := testStore.GetBookByID(id)
	if err != nil {
		t.Fatalf("Failed to retrieve updated book: %v", err)
	}

	if *updatedBook.Rating != newRating {
		t.Errorf("Book rating not updated, expected %d, got %d", newRating, *updatedBook.Rating)
	}

	if *updatedBook.Comments != newComments {
		t.Errorf("Book comments not updated, expected %s, got %s", newComments, *updatedBook.Comments)
	}
}

// TestDeleteBookHandler tests the DELETE /api/books/{id} endpoint
func TestDeleteBookHandler(t *testing.T) {
	// Add test book
	book := createTestBook(model.StatusWantToRead, "5")
	id, err := testStore.AddBook(book)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	// Create request
	req, err := http.NewRequest("DELETE", "/api/books/"+itoa(id), nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Create a ResponseRecorder
	rr := httptest.NewRecorder()

	// Call the handler
	testRouter.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}

	// Verify the book was deleted from the database
	_, err = testStore.GetBookByID(id)
	if err == nil {
		t.Errorf("Book was not deleted from the database")
	}
}

// TestSearchBooksHandler tests the GET /api/books/search endpoint
func TestSearchBooksHandler(t *testing.T) {
	// Add test books with different titles and authors
	book1 := createTestBook(model.StatusWantToRead, "Search1")
	book1.Title = "The Great Gatsby"
	book1.Author = "F. Scott Fitzgerald"
	_, err := testStore.AddBook(book1)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	book2 := createTestBook(model.StatusCurrentlyReading, "Search2")
	book2.Title = "The Great Adventure"
	book2.Author = "John Smith"
	_, err = testStore.AddBook(book2)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	// Test search by title
	req, err := http.NewRequest("GET", "/api/books/search?q=Great", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var books []model.Book
	if err := json.Unmarshal(rr.Body.Bytes(), &books); err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}

	// Verify that both books are in the results
	foundGatsby := false
	foundAdventure := false
	for _, book := range books {
		if book.Title == "The Great Gatsby" {
			foundGatsby = true
		}
		if book.Title == "The Great Adventure" {
			foundAdventure = true
		}
	}

	if !foundGatsby {
		t.Error("The Great Gatsby was not found in search results")
	}
	if !foundAdventure {
		t.Error("The Great Adventure was not found in search results")
	}
}

// TestAddBookHandlerInvalidInput tests the POST /api/books endpoint with invalid input
func TestAddBookHandlerInvalidInput(t *testing.T) {
	// Test with missing required fields
	invalidBook := map[string]interface{}{
		"title":  "", // Empty title
		"author": "Test Author",
	}
	jsonData, err := json.Marshal(invalidBook)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", "/api/books", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

// TestUpdateBookStatusHandlerInvalidStatus tests the PUT /api/books/{id} endpoint with invalid status
func TestUpdateBookStatusHandlerInvalidStatus(t *testing.T) {
	// Add test book
	book := createTestBook(model.StatusWantToRead, "InvalidStatus")
	id, err := testStore.AddBook(book)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	// Try to update with invalid status
	updateData := map[string]string{
		"status": "InvalidStatus",
	}
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("PUT", "/api/books/"+itoa(id), bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

// TestUpdateNonExistentBook tests updating a book that doesn't exist
func TestUpdateNonExistentBook(t *testing.T) {
	// Try to update a non-existent book
	updateData := map[string]string{
		"status": string(model.StatusCurrentlyReading),
	}
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("PUT", "/api/books/99999", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

// TestDeleteNonExistentBook tests deleting a book that doesn't exist
func TestDeleteNonExistentBook(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/api/books/99999", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

// TestMalformedJSON tests handling of malformed JSON input
func TestMalformedJSON(t *testing.T) {
	// Create request with malformed JSON
	malformedJSON := []byte(`{"title": "Test Book", "author": "Test Author", "status": invalid}`)
	req, err := http.NewRequest("POST", "/api/books", bytes.NewBuffer(malformedJSON))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

// TestUpdateBookDetailsHandlerPartialUpdate tests partial updates of book details
func TestUpdateBookDetailsHandlerPartialUpdate(t *testing.T) {
	// Add test book
	book := createTestBook(model.StatusWantToRead, "PartialUpdate")
	id, err := testStore.AddBook(book)
	if err != nil {
		t.Fatalf("Failed to add test book: %v", err)
	}

	// Update only the rating
	updateData := map[string]interface{}{
		"rating": 9,
	}
	jsonData, err := json.Marshal(updateData)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("PUT", "/api/books/"+itoa(id)+"/details", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Verify only rating was updated
	updatedBook, err := testStore.GetBookByID(id)
	if err != nil {
		t.Fatalf("Failed to retrieve updated book: %v", err)
	}

	if updatedBook.Rating == nil {
		t.Error("Book rating should not be nil after update")
	} else if *updatedBook.Rating != 9 {
		t.Errorf("Book rating not updated correctly, expected 9, got %d", *updatedBook.Rating)
	}

	// Only check comments if they were set in the original book
	if book.Comments != nil {
		if updatedBook.Comments == nil {
			t.Error("Book comments should not be nil after update")
		} else if *updatedBook.Comments != *book.Comments {
			t.Errorf("Book comments should not have changed, expected %s, got %s", *book.Comments, *updatedBook.Comments)
		}
	}
}
