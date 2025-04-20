package main

import (
	"flag" // Import flag package
	"log"
	"net/http"
	"os"
	"path/filepath"

	"bookshelf/api"
	"bookshelf/db"

	"github.com/gorilla/handlers" // Import handlers for logging
	"github.com/gorilla/mux"
)

func main() {
	// --- CLI Flags ---
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse() // Parse command-line flags

	// If other arguments are provided after parsing flags, treat it like help request?
	// Or just let the app run. For now, we just parse defined flags.
	// A more robust help would check for a specific --help flag, but flag package handles -h/-help automatically.

	// Initialize Database
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB() // Ensure DB connection is closed on exit

	// --- Server Setup using gorilla/mux ---
	r := mux.NewRouter() // Use gorilla/mux router

	// API Routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/search", api.SearchOpenLibraryHandler).Methods(http.MethodGet) // New search route
	apiRouter.HandleFunc("/books", api.GetBooksHandler).Methods(http.MethodGet)
	apiRouter.HandleFunc("/books", api.AddBookHandler).Methods(http.MethodPost)
	apiRouter.HandleFunc("/books/{id:[0-9]+}", api.UpdateBookHandler).Methods(http.MethodPut) // Route for updating status (drag-n-drop)
	apiRouter.HandleFunc("/books/{id:[0-9]+}/details", api.UpdateBookDetailsHandler).Methods(http.MethodPut) // New route for updating details
	// Add other API routes here later (DELETE /api/books/{id}, etc.)

	// Frontend Route - Serve static files from 'web' directory
	// Determine the directory of the executable
	ex, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	exePath := filepath.Dir(ex)
	// Assume 'web' directory is relative to the executable
	// This works well for deployment. For development, ensure 'web' is in the CWD or adjust path.
	webDir := http.Dir(filepath.Join(exePath, "web"))
	// Fallback for development: check current working directory if 'web' not found relative to exe
	if _, err := os.Stat(filepath.Join(exePath, "web", "index.html")); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		webDir = http.Dir(filepath.Join(cwd, "web"))
		log.Printf("Serving static files from CWD: %s", filepath.Join(cwd, "web"))
	} else {
		log.Printf("Serving static files from executable path: %s", filepath.Join(exePath, "web"))
	}

	// Serve static files using PathPrefix, ensuring it doesn't clash with API routes
	// The handler needs to strip the prefix AND serve the file.
	// http.StripPrefix is not directly compatible with gorilla/mux's Handle function in this way.
	// A common pattern is to use PathPrefix("/").Handler(...) as a catch-all AFTER specific routes.
	fs := http.FileServer(webDir)
	r.PathPrefix("/").Handler(fs) // Serve static files for any path not matched above

	// --- Logging Middleware ---
	// Log requests to stdout in Apache Combined Log Format
	loggedRouter := handlers.LoggingHandler(os.Stdout, r)

	// Start Server
	log.Printf("Starting server on http://localhost:%s", *port)
	// Use the loggedRouter which wraps the original router 'r'
	if err := http.ListenAndServe(":"+*port, loggedRouter); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
