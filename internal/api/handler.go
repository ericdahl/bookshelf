package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/ericdahl/bookshelf/internal/db"
	"github.com/ericdahl/bookshelf/internal/model"
)

// APIHandler holds dependencies for API handlers, like the database store.
type APIHandler struct {
	Store      db.BookStore
	HTTPClient *http.Client // For Open Library calls
}

// NewAPIHandler creates a new APIHandler with dependencies.
func NewAPIHandler(store db.BookStore) *APIHandler {
	return &APIHandler{
		Store: store,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second, // Sensible timeout for external API calls
		},
	}
}

// --- Helper Functions ---

// respondWithError sends a JSON error response.
func respondWithError(w http.ResponseWriter, code int, message string) {
	slog.Error("HTTP Error", "code", code, "message", message)
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON sends a JSON response.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Error marshalling JSON response", "error", err)
		// Fallback to plain text error if marshalling fails
		http.Error(w, `{"error":"Failed to marshal JSON response"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		slog.Error("Error writing JSON response", "error", err)
	}
}

// --- Book Handlers ---

// GetBooksHandler handles GET /api/books requests.
func (h *APIHandler) GetBooksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := h.Store.GetBooks()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve books: "+err.Error())
		return
	}
	if books == nil {
		books = []model.Book{} // Return empty array instead of null
	}
	respondWithJSON(w, http.StatusOK, books)
}

// AddBookHandler handles POST /api/books requests.
// Expects JSON body based on Open Library search result selection.
func (h *APIHandler) AddBookHandler(w http.ResponseWriter, r *http.Request) {
	var book model.Book

	// Limit request body size to prevent potential abuse
	r.Body = http.MaxBytesReader(w, r.Body, 1*1024*1024) // 1 MB limit

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Prevent unexpected fields

	if err := decoder.Decode(&book); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			respondWithError(w, http.StatusBadRequest, msg)
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			respondWithError(w, http.StatusBadRequest, msg)
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			respondWithError(w, http.StatusBadRequest, msg)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			respondWithError(w, http.StatusBadRequest, msg)
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			respondWithError(w, http.StatusBadRequest, msg)
		case errors.As(err, &maxBytesError):
			msg := fmt.Sprintf("Request body must not be larger than %d bytes", maxBytesError.Limit)
			respondWithError(w, http.StatusRequestEntityTooLarge, msg)
		default:
			respondWithError(w, http.StatusBadRequest, "Failed to decode request body: "+err.Error())
		}
		return
	}

	// Basic validation for required fields from search result
	if book.Title == "" || book.OpenLibraryID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing required fields: title and open_library_id")
		return
	}
	// Author is highly recommended but might be missing in some OL entries
	if book.Author == "" {
		slog.Warn("Adding book with missing author", 
			"title", book.Title, 
			"openLibraryID", book.OpenLibraryID)
		book.Author = "Unknown Author" // Provide a default or handle differently
	}

	// Set default status if not provided or invalid
	if book.Status == "" || !book.Status.IsValid() {
		// Defaulting to "Want to Read" as per README, not "Currently Reading" as per initial prompt.
		// Let's stick to "Want to Read" as a safer default.
		book.Status = model.StatusWantToRead
	}

	// Ensure Rating and Comments are initially null unless explicitly provided (unlikely for Add)
	book.Rating = nil
	book.Comments = nil

	// Validate the model (e.g., rating range, although unlikely here)
	if err := book.Validate(); err != nil {
		var validationErr *model.ValidationError
		if errors.As(err, &validationErr) {
			respondWithError(w, http.StatusBadRequest, validationErr.Message)
		} else {
			respondWithError(w, http.StatusBadRequest, "Invalid book data: "+err.Error())
		}
		return
	}

	// Add the book to the database
	newID, err := h.Store.AddBook(&book)
	if err != nil {
		// TODO: Check for specific DB errors like UNIQUE constraint violation
		respondWithError(w, http.StatusInternalServerError, "Failed to add book to database: "+err.Error())
		return
	}

	book.ID = newID // Ensure the returned book has the ID
	respondWithJSON(w, http.StatusCreated, book)
}

// UpdateBookStatusHandler handles PUT /api/books/{id} requests (for status update).
func (h *APIHandler) UpdateBookStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing book ID")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid book ID format")
		return
	}

	var payload struct {
		Status model.BookStatus `json:"status"`
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		// Basic error handling, can be expanded like in AddBookHandler
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	if !payload.Status.IsValid() {
		respondWithError(w, http.StatusBadRequest, "Invalid status value. Must be 'Want to Read', 'Currently Reading', or 'Read'")
		return
	}

	err = h.Store.UpdateBookStatus(id, payload.Status)
	if err != nil {
		// TODO: Differentiate between Not Found (404) and other errors (500)
		if strings.Contains(err.Error(), "not found") { // Basic check, better to use custom errors
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to update book status: "+err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Book status updated successfully"})
}

// UpdateBookDetailsHandler handles PUT /api/books/{id}/details requests (for rating, comments, and series info).
func (h *APIHandler) UpdateBookDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Missing book ID")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid book ID format")
		return
	}

	// Use pointers in the payload struct to detect if a field was provided (even if null)
	var payload struct {
		Rating      *int    `json:"rating"`        // Pointer allows distinguishing between 0 and not provided/null
		Comments    *string `json:"comments"`      // Pointer allows distinguishing between "" and not provided/null
		Series      *string `json:"series"`        // Name of series
		SeriesIndex *int    `json:"series_index"` // Position in series
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Important here

	err = decoder.Decode(&payload)
	if err != nil {
		// Basic error handling
		respondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate rating if provided
	if payload.Rating != nil && (*payload.Rating < 1 || *payload.Rating > 10) {
		respondWithError(w, http.StatusBadRequest, "Rating must be between 1 and 10")
		return
	}
	
	// Validate series index if provided
	if payload.SeriesIndex != nil && *payload.SeriesIndex <= 0 {
		respondWithError(w, http.StatusBadRequest, "Series index must be greater than 0")
		return
	}
	
	// If series is null/empty but series_index is provided, return error
	if (payload.Series == nil || *payload.Series == "") && payload.SeriesIndex != nil {
		respondWithError(w, http.StatusBadRequest, "Cannot provide series_index without series name")
		return
	}

	// Perform the update
	err = h.Store.UpdateBookDetails(id, payload.Rating, payload.Comments, payload.Series, payload.SeriesIndex)
	if err != nil {
		// Differentiate between Not Found (404) and other errors (500)
		if strings.Contains(err.Error(), "not found") { // Basic check
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to update book details: "+err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Book details updated successfully"})
}

// DeleteBookHandler handles the deletion of a book
func (h *APIHandler) DeleteBookHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		http.Error(w, "Invalid book ID", http.StatusBadRequest)
		return
	}

	err = h.Store.DeleteBook(id)
	if err != nil {
		slog.Error("Error deleting book", "error", err, "id", id)
		http.Error(w, "Failed to delete book", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Open Library Search Handler ---

// OpenLibrarySearchResult defines the structure we want to return from our search endpoint.
type OpenLibrarySearchResult struct {
	OpenLibraryID string  `json:"open_library_id"` // e.g., OL7353617M
	Title         string  `json:"title"`
	Author        string  `json:"author"`              // Combined author names
	ISBN          *string `json:"isbn,omitempty"`      // First available ISBN-13 or ISBN-10
	CoverURL      *string `json:"cover_url,omitempty"` // URL for medium cover
	// Fields to identify if book already exists in library
	ExistingID    *int64       `json:"existing_id,omitempty"`    // ID if book already in library
	ExistingShelf *string      `json:"existing_shelf,omitempty"` // Shelf name if already in library
}

// openLibrarySearchResponse is the structure matching the Open Library Search API JSON response.
// We only map the fields we need. See: https://openlibrary.org/dev/docs/api/search
type openLibrarySearchResponse struct {
	NumFound int `json:"numFound"`
	Docs     []struct {
		Key              string   `json:"key"` // e.g., "/works/OL7353617M"
		Title            string   `json:"title"`
		AuthorName       []string `json:"author_name"` // Array of author names
		ISBN             []string `json:"isbn"`        // Array of ISBNs (10 and 13)
		CoverI           int      `json:"cover_i"`     // Cover ID (integer)
		AuthorKey        []string `json:"author_key"`  // Array of author IDs
		FirstPublishYear int      `json:"first_publish_year"`
	} `json:"docs"`
}

// SearchBooksHandler handles GET /api/search?q={query}
func (h *APIHandler) SearchBooksHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondWithError(w, http.StatusBadRequest, "Missing search query parameter 'q'")
		return
	}

	// Construct Open Library API URL
	// Using the works search endpoint as it often has better consolidated data
	apiURL := fmt.Sprintf("https://openlibrary.org/search.json?q=%s&fields=key,title,author_name,isbn,cover_i,author_key,first_publish_year&limit=20", url.QueryEscape(query))
	slog.Info("Querying Open Library", "url", apiURL)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, apiURL, nil)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create Open Library request: "+err.Error())
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "BookshelfApp/1.0 (github.com/ericdahl/bookshelf; contact@example.com)") // Be a good API citizen

	start := time.Now()
	resp, err := h.HTTPClient.Do(req)
	elapsed := time.Since(start)
	if resp != nil {
		slog.Info("OpenLibrary API response", 
			"url", apiURL, 
			"status", resp.StatusCode, 
			"responseTime", elapsed)
	}

	if err != nil {
		respondWithError(w, http.StatusBadGateway, "Failed to contact Open Library API: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) // Read body for context, ignore error
		errMsg := fmt.Sprintf("Open Library API returned status %d: %s", resp.StatusCode, string(bodyBytes))
		respondWithError(w, http.StatusBadGateway, errMsg)
		return
	}

	// Decode the response
	var olResponse openLibrarySearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&olResponse); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to decode Open Library response: "+err.Error())
		return
	}

	// Get all existing books and create a map for quick lookup
	existingBooks, err := h.Store.GetBooks()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve existing books: "+err.Error())
		return
	}

	// Create a map of OpenLibraryID -> Book for quick lookup
	existingBooksMap := make(map[string]model.Book)
	for _, book := range existingBooks {
		existingBooksMap[book.OpenLibraryID] = book
	}

	// Transform the results into our desired format
	results := []OpenLibrarySearchResult{}
	for _, doc := range olResponse.Docs {
		// Extract OpenLibrary ID from the key (e.g., "/works/OL7353617M" -> "OL7353617M")
		parts := strings.Split(doc.Key, "/")
		olid := ""
		if len(parts) > 0 {
			olid = parts[len(parts)-1]
		}
		if olid == "" {
			continue // Skip if we can't get an ID
		}

		// Find a suitable ISBN (prefer ISBN-13)
		var isbn *string
		for _, code := range doc.ISBN {
			if len(code) == 13 { // Basic check for ISBN-13
				isbn = &code
				break
			}
		}
		if isbn == nil && len(doc.ISBN) > 0 { // Fallback to first ISBN if no 13 found
			isbn = &doc.ISBN[0]
		}

		// Construct cover URL (Medium size)
		var coverURL *string
		if doc.CoverI > 0 {
			url := fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", doc.CoverI)
			coverURL = &url
		}

		result := OpenLibrarySearchResult{
			OpenLibraryID: olid,
			Title:         doc.Title,
			Author:        strings.Join(doc.AuthorName, ", "), // Combine authors
			ISBN:          isbn,
			CoverURL:      coverURL,
		}
		
		// Check if the book exists in the user's library
		if existingBook, exists := existingBooksMap[olid]; exists {
			// Book exists in the library, set the ExistingID and ExistingShelf fields
			result.ExistingID = &existingBook.ID
			shelf := string(existingBook.Status)
			result.ExistingShelf = &shelf
		}
		
		results = append(results, result)
	}

	respondWithJSON(w, http.StatusOK, results)
}
