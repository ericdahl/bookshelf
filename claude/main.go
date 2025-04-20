package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Book represents a book in the system
type Book struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Rating float64 `json:"rating"`
}

var books []Book

func main() {
	// Initialize the database
	InitDB("./vibe_books.db")
	
	router := mux.NewRouter()
	
	// Initialize sample data for in-memory mode
	books = append(books, Book{ID: "1", Title: "The Go Programming Language", Author: "Alan A. A. Donovan", Rating: 4.5})
	books = append(books, Book{ID: "2", Title: "Clean Code", Author: "Robert C. Martin", Rating: 4.7})
	
	// API Routes for in-memory mode
	router.HandleFunc("/api/books", getBooks).Methods("GET")
	router.HandleFunc("/api/books/{id}", getBook).Methods("GET")
	router.HandleFunc("/api/books", createBook).Methods("POST")
	router.HandleFunc("/api/books/{id}", updateBook).Methods("PUT")
	router.HandleFunc("/api/books/{id}", deleteBook).Methods("DELETE")
	
	// New routes for database API
	router.HandleFunc("/api/db/books", GetBooksHandler).Methods("GET")
	router.HandleFunc("/api/db/books/{id}", GetBookHandler).Methods("GET")
	router.HandleFunc("/api/db/books", CreateBookHandler).Methods("POST")
	router.HandleFunc("/api/db/books/{id}", UpdateBookHandler).Methods("PUT")
	router.HandleFunc("/api/db/books/{id}", DeleteBookHandler).Methods("DELETE")
	
	// Shelves routes
	router.HandleFunc("/api/shelves", GetShelvesHandler).Methods("GET")
	router.HandleFunc("/api/shelves/{id}/books", GetBooksInShelfHandler).Methods("GET")
	router.HandleFunc("/api/shelves/{shelfId}/books/{bookId}", AddBookToShelfHandler).Methods("POST")
	router.HandleFunc("/api/shelves/{shelfId}/books/{bookId}", RemoveBookFromShelfHandler).Methods("DELETE")
	
	log.Println("Server started on :8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}

// getBooks returns all books
func getBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

// getBook returns a specific book
func getBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	
	for _, item := range books {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "Book not found"})
}

// createBook adds a new book
func createBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var book Book
	
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}
	
	books = append(books, book)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(book)
}

// updateBook updates an existing book
func updateBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	
	for index, item := range books {
		if item.ID == params["id"] {
			var book Book
			err := json.NewDecoder(r.Body).Decode(&book)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
				return
			}
			
			book.ID = params["id"]
			books[index] = book
			
			json.NewEncoder(w).Encode(book)
			return
		}
	}
	
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "Book not found"})
}

// deleteBook removes a book
func deleteBook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	
	for index, item := range books {
		if item.ID == params["id"] {
			books = append(books[:index], books[index+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{"error": "Book not found"})
}