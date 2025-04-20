package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
)

// GetBooksHandler returns all books
func GetBooksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	books, err := GetAllBooks()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve books"})
		return
	}

	json.NewEncoder(w).Encode(books)
}

// GetBookHandler returns a specific book
func GetBookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]

	book, err := GetBookByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Book not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve book"})
		}
		return
	}

	json.NewEncoder(w).Encode(book)
}

// CreateBookHandler adds a new book
func CreateBookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var book DbBook

	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	book.ID = generateID()

	err = AddBook(book)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create book"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

// UpdateBookHandler updates an existing book
func UpdateBookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]

	// Check if book exists
	_, err := GetBookByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Book not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve book"})
		}
		return
	}

	var book DbBook
	err = json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}

	book.ID = id

	err = UpdateBook(book)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update book"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)
}

// DeleteBookHandler removes a book
func DeleteBookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	id := params["id"]

	// Check if book exists
	_, err := GetBookByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "Book not found"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve book"})
		}
		return
	}

	err = DeleteBook(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to delete book"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetShelvesHandler returns all shelves
func GetShelvesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	shelves, err := GetShelves()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve shelves"})
		return
	}

	json.NewEncoder(w).Encode(shelves)
}

// AddBookToShelfHandler adds a book to a shelf
func AddBookToShelfHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	bookID := params["bookId"]
	shelfID := params["shelfId"]

	err := AddBookToShelf(bookID, shelfID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to add book to shelf"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveBookFromShelfHandler removes a book from a shelf
func RemoveBookFromShelfHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	bookID := params["bookId"]
	shelfID := params["shelfId"]

	err := RemoveBookFromShelf(bookID, shelfID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to remove book from shelf"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetBooksInShelfHandler returns all books in a shelf
func GetBooksInShelfHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	shelfID := params["id"]

	books, err := GetBooksInShelf(shelfID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve books in shelf"})
		return
	}

	json.NewEncoder(w).Encode(books)
}

// OpenLibraryBookResponse represents a book result from the Open Library API
type OpenLibraryBookResponse struct {
	Key         string   `json:"key"`
	Title       string   `json:"title"`
	AuthorName  []string `json:"author_name,omitempty"`
	ISBN        []string `json:"isbn,omitempty"`
	CoverID     int      `json:"cover_i,omitempty"`
	PublishYear int      `json:"first_publish_year,omitempty"`
	Publisher   []string `json:"publisher,omitempty"`
}

// OpenLibrarySearchResponse represents the search response from the Open Library API
type OpenLibrarySearchResponse struct {
	NumFound int                     `json:"numFound"`
	Start    int                     `json:"start"`
	Docs     []OpenLibraryBookResponse `json:"docs"`
}

// SearchOpenLibraryHandler handles search requests to the Open Library API
func SearchOpenLibraryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Query parameter 'q' is required"})
		return
	}
	
	// Make request to Open Library API
	escapedQuery := url.QueryEscape(query)
	apiURL := fmt.Sprintf("https://openlibrary.org/search.json?q=%s&limit=10", escapedQuery)
	
	resp, err := http.Get(apiURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to connect to Open Library API"})
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Open Library API returned an error"})
		return
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read response from Open Library API"})
		return
	}
	
	var olResponse OpenLibrarySearchResponse
	if err := json.Unmarshal(body, &olResponse); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse response from Open Library API"})
		return
	}
	
	// Format and send back the results
	json.NewEncoder(w).Encode(olResponse.Docs)
}

// GetOpenLibraryBookHandler gets detailed information about a specific Open Library book
func GetOpenLibraryBookHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	params := mux.Vars(r)
	bookID := params["id"]
	if bookID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Book ID is required"})
		return
	}
	
	// Remove any path prefix
	bookID = strings.TrimPrefix(bookID, "/works/")
	apiURL := fmt.Sprintf("https://openlibrary.org/works/%s.json", bookID)
	
	resp, err := http.Get(apiURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to connect to Open Library API"})
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Open Library API returned an error"})
		return
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read response from Open Library API"})
		return
	}
	
	// Pass through the raw response for the frontend to parse
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse response from Open Library API"})
		return
	}
	
	json.NewEncoder(w).Encode(rawResponse)
}