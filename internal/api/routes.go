package api

import (
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/klauspost/compress/gzip"
)

// LoggingMiddleware logs incoming HTTP requests.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Use a response writer wrapper if you need to capture status code
		// For simple logging, this is often sufficient.
		slog.Debug("HTTP Request started",
			"method", r.Method,
			"uri", r.RequestURI,
			"remoteAddr", r.RemoteAddr)

		next.ServeHTTP(w, r) // Call the next handler

		// Log after the request is handled
		// Note: Status code logging requires a response writer wrapper.
		// For now, just log duration.
		slog.Info("HTTP Request completed",
			"method", r.Method,
			"uri", r.RequestURI,
			"duration", time.Since(start))
	})
}

// GzipMiddleware compresses responses using gzip if the client accepts it
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Create a gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Set headers
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		// Create a response writer that writes to the gzip writer
		gzw := gzipResponseWriter{Writer: gz, ResponseWriter: w}

		// Call the next handler with our gzip response writer
		next.ServeHTTP(gzw, r)
	})
}

// gzipResponseWriter is a custom response writer that writes to a gzip writer
type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// SetupRouter configures the routes for the application.
func SetupRouter(apiHandler *APIHandler, webDir string) *mux.Router {
	r := mux.NewRouter()

	// Apply middlewares to all routes
	r.Use(LoggingMiddleware)
	r.Use(GzipMiddleware)

	// API Routes (prefixed with /api)
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/books", apiHandler.GetBooksHandler).Methods(http.MethodGet)
	apiRouter.HandleFunc("/books", apiHandler.AddBookHandler).Methods(http.MethodPost)
	apiRouter.HandleFunc("/books/{id:[0-9]+}", apiHandler.UpdateBookStatusHandler).Methods(http.MethodPut)          // For status update
	apiRouter.HandleFunc("/books/{id:[0-9]+}/type", apiHandler.UpdateBookTypeHandler).Methods(http.MethodPut)       // For type update
	apiRouter.HandleFunc("/books/{id:[0-9]+}/details", apiHandler.UpdateBookDetailsHandler).Methods(http.MethodPut) // For rating/comments
	apiRouter.HandleFunc("/books/search", apiHandler.SearchBooksHandler).Methods(http.MethodGet)                    // Expects ?q=query
	apiRouter.HandleFunc("/books/{id:[0-9]+}", apiHandler.DeleteBookHandler).Methods(http.MethodDelete)             // Delete a book

	// Static File Server for Frontend
	// Serve files from the web directory.
	fs := http.FileServer(http.Dir(webDir))

	// Serve index.html for all non-API routes to support SPA routing
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request is for a file (has an extension)
		if strings.Contains(r.URL.Path, ".") {
			// Serve the file directly
			fs.ServeHTTP(w, r)
			return
		}

		// For all other routes, serve index.html to support client-side routing
		http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
	})

	slog.Info("Router setup complete")
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
