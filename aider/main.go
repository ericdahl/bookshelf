package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"bookshelf/api"
	"bookshelf/db"
)

func main() {
	// Initialize Database
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB() // Ensure DB connection is closed on exit

	// --- Server Setup ---
	mux := http.NewServeMux()

	// API Routes
	mux.HandleFunc("/api/books", api.GetBooksHandler)
	// Add other API routes here later (POST /api/books, PUT /api/books/{id}, etc.)

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

	fs := http.FileServer(webDir)
	mux.Handle("/", fs) // Serve index.html and other assets

	// Start Server
	port := "8080"
	log.Printf("Starting server on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
