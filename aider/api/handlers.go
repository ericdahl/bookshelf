package api

import (
	"database/sql" // Import database/sql for ErrNoRows
	"encoding/json"
	"fmt" // Import fmt
	"log"
	"net/http"
	"strconv" // Add strconv for parsing ID

	"bookshelf/db"
	"bookshelf/models" // Add models import

	"github.com/gorilla/mux" // Import gorilla/mux
)

// GetBooksHandler handles GET requests to retrieve all books.
func GetBooksHandler(w http.ResponseWriter, r *http.Request) {
	// Method check is handled by mux router
	// if r.Method != http.MethodGet {
	// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	// 	return
	// }

	books, err := db.GetAllBooks()
	if err != nil {
		log.Printf("Error getting books from DB: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(books); err != nil {
		log.Printf("Error encoding books to JSON: %v", err)
		// Don't write error header if already started writing response body
	}
}

// AddBookHandler handles POST requests to add a new book.
func AddBookHandler(w http.ResponseWriter, r *http.Request) {
	var book models.Book
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&book); err != nil {
		log.Printf("Error decoding book JSON: %v", err)
		http.Error(w, "Bad request: Invalid JSON format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Basic validation (can be expanded)
	if book.Title == "" {
		http.Error(w, "Bad request: Title is required", http.StatusBadRequest)
		return
	}
	// Ensure default status if not provided by client (though frontend does)
	if book.Status == "" {
		book.Status = models.StatusWantToRead
	}

	// TODO: Later, this handler will expect OpenLibraryID and ISBN from search results,
	// and potentially fetch Author/CoverURL itself if needed.
	// For now, just ensure the model structure matches for compilation.
	// We'll add validation for ISBN presence here once the frontend sends it.
	// if book.ISBN == "" {
	// 	http.Error(w, "Bad request: ISBN is required (will be obtained from Open Library search)", http.StatusBadRequest)
	// 	return
	// }


	// Add book to database
	id, err := db.AddBook(book) // AddBook now expects ISBN
	if err != nil {
		log.Printf("Error adding book to DB: %v", err)
		// Check for specific errors like the one from AddBook validation
		if err.Error() == "book title and ISBN are required" {
			http.Error(w, fmt.Sprintf("Bad request: %v", err), http.StatusBadRequest)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Return the newly created book (or just the ID)
	book.ID = id // Set the ID in the response object
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created
	if err := json.NewEncoder(w).Encode(book); err != nil {
		log.Printf("Error encoding newly added book to JSON: %v", err)
	}
}

// UpdateBookHandler handles PUT requests to update a book's status.
func UpdateBookHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		http.Error(w, "Bad request: Missing book ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Bad request: Invalid book ID format", http.StatusBadRequest)
		return
	}

	var payload struct {
		Status string `json:"status"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		log.Printf("Error decoding update status JSON: %v", err)
		http.Error(w, "Bad request: Invalid JSON format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate status (optional but good practice)
	isValidStatus := false
	validStatuses := []string{models.StatusWantToRead, models.StatusCurrentlyReading, models.StatusRead}
	for _, s := range validStatuses {
		if payload.Status == s {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus || payload.Status == "" {
		http.Error(w, "Bad request: Invalid or missing status value", http.StatusBadRequest)
		return
	}

	// Update book status in database
	err = db.UpdateBookStatus(id, payload.Status)
	if err != nil {
		// Handle specific errors like not found
		if err == sql.ErrNoRows { // Use errors.Is(err, sql.ErrNoRows) in Go 1.13+
			log.Printf("Attempted to update non-existent book ID: %d", id)
			http.Error(w, "Not found: Book with given ID does not exist", http.StatusNotFound)
		} else {
			log.Printf("Error updating book status in DB for ID %d: %v", id, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK) // 200 OK (or 204 No Content if preferred)
	// Optionally return the updated status or book object
	// json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// Add other handlers (SearchOpenLibrary, DeleteBook, etc.) later.
