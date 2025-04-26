package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// LoggingMiddleware logs incoming HTTP requests.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Use a response writer wrapper if you need to capture status code
		// For simple logging, this is often sufficient.
		log.Printf("HTTP: Started %s %s from %s", r.Method, r.RequestURI, r.RemoteAddr)

		next.ServeHTTP(w, r) // Call the next handler

		// Log after the request is handled
		// Note: Status code logging requires a response writer wrapper.
		// For now, just log duration.
		log.Printf("HTTP: Completed %s %s in %v", r.Method, r.RequestURI, time.Since(start))
	})
}

// SetupRouter configures the routes for the application.
func SetupRouter(apiHandler *APIHandler, webDir string) *mux.Router {
	r := mux.NewRouter()

	// Apply logging middleware to all routes
	r.Use(LoggingMiddleware)

	// API Routes (prefixed with /api)
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/books", apiHandler.GetBooksHandler).Methods(http.MethodGet)
	apiRouter.HandleFunc("/books", apiHandler.AddBookHandler).Methods(http.MethodPost)
	apiRouter.HandleFunc("/books/{id:[0-9]+}", apiHandler.UpdateBookStatusHandler).Methods(http.MethodPut) // For status update
	apiRouter.HandleFunc("/books/{id:[0-9]+}/details", apiHandler.UpdateBookDetailsHandler).Methods(http.MethodPut) // For rating/comments
	apiRouter.HandleFunc("/search", apiHandler.SearchBooksHandler).Methods(http.MethodGet) // Expects ?q=query

	// TODO: Add DELETE /api/books/{id} route later

	// Static File Server for Frontend
	// Serve files from the 'web' directory.
	// Use PathPrefix("/") to catch all non-API routes.
	// Ensure this is registered *after* the API routes.
	fs := http.FileServer(http.Dir(webDir))
	// Use NoDirListing handler to prevent directory browsing
	r.PathPrefix("/").Handler(NoDirListing(fs))


	log.Println("Router setup complete.")
	return r
}

// NoDirListing wraps a http.Handler (like http.FileServer) and prevents directory listings.
func NoDirListing(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested path ends with a slash (potential directory listing)
		// This isn't foolproof but covers common cases. A better check might involve
		// os.Stat on the underlying file path if possible.
		if strings.HasSuffix(r.URL.Path, "/") && r.URL.Path != "/" {
			// If it looks like a directory path (and isn't the root), return 404.
			// This prevents FileServer from generating a directory listing.
			// Serve index.html for the root path "/" is handled implicitly by FileServer
			// if index.html exists in the webDir.
			http.NotFound(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}
