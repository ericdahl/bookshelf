package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"bookshelf/api"
	"bookshelf/db"

	"github.com/gorilla/mux" // Import gorilla/mux
)

func main() {
	// Initialize Database
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB() // Ensure DB connection is closed on exit

	// --- Server Setup using gorilla/mux ---
	r := mux.NewRouter() // Use gorilla/mux router

	// API Routes
	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/books", api.GetBooksHandler).Methods(http.MethodGet)
	apiRouter.HandleFunc("/books", api.AddBookHandler).Methods(http.MethodPost)
	apiRouter.HandleFunc("/books/{id:[0-9]+}", api.UpdateBookHandler).Methods(http.MethodPut) // Route for updating status
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

	// Start Server
	port := "8080"
	log.Printf("Starting server on http://localhost:%s", port)
	// Use the gorilla/mux router 'r' instead of the default 'mux'
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
