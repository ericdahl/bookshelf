package api

import (
	"encoding/json"
	"log"
	"net/http"

	"bookshelf/db"
)

// GetBooksHandler handles requests to retrieve all books.
func GetBooksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

// Add other handlers here (AddBook, UpdateBook, SearchOpenLibrary, etc.) later.
