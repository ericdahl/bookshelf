package api

import (
	"database/sql" // Import database/sql for ErrNoRows
	"encoding/json"
	"fmt" // Import fmt
	"log"
	"net/http"
	"net/url" // Import net/url for query escaping
	"strconv" // Add strconv for parsing ID
	"strings" // Import strings for joining author names etc.
	"time"    // Import time for http client timeout

	"bookshelf/db"
	"bookshelf/models" // Add models import

	// "net/url" // Duplicate removed
	// "time"    // Duplicate removed

	"github.com/gorilla/mux" // Import gorilla/mux
)

// Structure for Open Library Search API response (simplified)
type OpenLibrarySearchResponse struct {
	Docs []struct {
		Key         string   `json:"key"` // e.g., "/works/OL45883W" -> extract OLID
		Title       string   `json:"title"`
		AuthorName  []string `json:"author_name"`
		ISBN        []string `json:"isbn"` // Can have multiple ISBNs (10, 13)
		CoverI      int      `json:"cover_i"` // Cover ID, can be used to build cover URL
	} `json:"docs"`
}

// Structure for simplified search results returned to our frontend
type BookSearchResult struct {
	OpenLibraryID string `json:"open_library_id"`
	Title         string `json:"title"`
	Author        string `json:"author"` // Combined author names
	ISBN          string `json:"isbn"`   // First valid ISBN found
	CoverURL      string `json:"cover_url"` // Optional cover URL
}


// --- HTTP Client for Open Library ---
// Reuse a client for better performance
var openLibraryClient = &http.Client{
	Timeout: 10 * time.Second, // Set a reasonable timeout
}

// SearchOpenLibraryHandler handles requests to search books on Open Library.
func SearchOpenLibraryHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Bad request: Missing query parameter 'q'", http.StatusBadRequest)
		return
	}

	// Construct Open Library API URL
	// We request fields: key, title, author_name, isbn, cover_i
	apiURL := fmt.Sprintf("https://openlibrary.org/search.json?q=%s&fields=key,title,author_name,isbn,cover_i&limit=10", url.QueryEscape(query))
	log.Printf("Querying Open Library: %s", apiURL)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		log.Printf("Error creating Open Library request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	req.Header.Add("Accept", "application/json") // Ensure we get JSON

	resp, err := openLibraryClient.Do(req)
	if err != nil {
		log.Printf("Error querying Open Library: %v", err)
		http.Error(w, "Error contacting Open Library", http.StatusBadGateway) // 502 Bad Gateway might be appropriate
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Open Library API returned non-OK status: %d", resp.StatusCode)
		http.Error(w, "Error response from Open Library", http.StatusBadGateway)
		return
	}

	var olResponse OpenLibrarySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&olResponse); err != nil {
		log.Printf("Error decoding Open Library response: %v", err)
		http.Error(w, "Error processing Open Library response", http.StatusInternalServerError)
		return
	}

	// Process results: Filter for books with ISBN and simplify
	results := []BookSearchResult{}
	for _, doc := range olResponse.Docs {
		var primaryISBN string
		// Find the first valid ISBN (prefer 13-digit if available, otherwise first 10-digit)
		for _, isbn := range doc.ISBN {
			if len(isbn) == 13 { // Basic check for ISBN-13
				primaryISBN = isbn
				break
			}
		}
		if primaryISBN == "" && len(doc.ISBN) > 0 {
			for _, isbn := range doc.ISBN {
				if len(isbn) == 10 { // Basic check for ISBN-10
					primaryISBN = isbn
					break
				}
			}
		}

		// Only include results that have an ISBN
		if primaryISBN != "" && doc.Title != "" {
			olid := ""
			// Extract OLID from key like "/works/OL45883W" or "/books/OL7353617M"
			parts := strings.Split(doc.Key, "/")
			if len(parts) > 0 {
				olid = parts[len(parts)-1]
			}

			author := ""
			if len(doc.AuthorName) > 0 {
				author = strings.Join(doc.AuthorName, ", ")
			}

			coverURL := ""
			if doc.CoverI > 0 {
				// Construct cover URL (medium size)
				coverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", doc.CoverI)
			}


			results = append(results, BookSearchResult{
				OpenLibraryID: olid,
				Title:         doc.Title,
				Author:        author,
				ISBN:          primaryISBN,
				CoverURL:      coverURL,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Printf("Error encoding search results to JSON: %v", err)
	}
}


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

	// Validation: Frontend should now send title, author, isbn, open_library_id
	// derived from the Open Library search selection.
	if book.Title == "" || book.ISBN == "" || book.OpenLibraryID == "" {
		errMsg := fmt.Sprintf("Bad request: Missing required fields from selection (Title: '%s', ISBN: '%s', OLID: '%s')", book.Title, book.ISBN, book.OpenLibraryID)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Set default status if not provided (though frontend should default to 'Want to Read')
	if book.Status == "" {
		book.Status = models.StatusWantToRead
	}

	// Add book to database using the details obtained from the search selection
	// The db.AddBook function already expects a models.Book object.
	id, err := db.AddBook(book)
	if err != nil {
		log.Printf("Error adding book to DB: %v", err)
		// Check for specific validation errors from db.AddBook
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

// UpdateBookDetailsHandler handles PUT requests to update a book's rating and comments.
func UpdateBookDetailsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Define a struct to decode the payload, using pointers for nullable fields
	var payload struct {
		Rating   *int    `json:"rating"`
		Comments *string `json:"comments"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		log.Printf("Error decoding update details JSON: %v", err)
		http.Error(w, "Bad request: Invalid JSON format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate rating if provided (must be between 1 and 10)
	if payload.Rating != nil && (*payload.Rating < 1 || *payload.Rating > 10) {
		http.Error(w, "Bad request: Rating must be between 1 and 10", http.StatusBadRequest)
		return
	}

	// Update book details in the database
	err = db.UpdateBookDetails(id, payload.Rating, payload.Comments)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Attempted to update details for non-existent book ID: %d", id)
			http.Error(w, "Not found: Book with given ID does not exist", http.StatusNotFound)
		} else {
			log.Printf("Error updating book details in DB for ID %d: %v", id, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK) // 200 OK
	// Optionally return the updated book or just status
	// For simplicity, just return OK. Frontend can update its view.
}


// Add other handlers (DeleteBook, etc.) later.
