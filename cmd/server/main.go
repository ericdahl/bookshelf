package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ericdahl/bookshelf/internal/api"
	"github.com/ericdahl/bookshelf/internal/db"
)

func checkWebDir(webDir string) error {
	webDirAbs, err := filepath.Abs(webDir)
	if err != nil {
		return fmt.Errorf("could not determine absolute path for web directory '%s': %v", webDir, err)
	}
	
	if _, err := os.Stat(webDirAbs); os.IsNotExist(err) {
		return fmt.Errorf("web directory '%s' (absolute: '%s') does not exist", webDir, webDirAbs)
	} else if err != nil {
		return fmt.Errorf("error checking web directory '%s': %v", webDirAbs, err)
	}
	
	return nil
}

func main() {
	// --- Configuration ---
	// Define command-line flags
	port := flag.Int("port", 8080, "Port number for the HTTP server")
	// Default DB location relative to executable or CWD
	defaultDbPath := "./bookshelf.db"
	dbFile := flag.String("db-file", defaultDbPath, "Path to the SQLite database file")
	webDir := flag.String("web-dir", "./web", "Directory containing static web assets (HTML, CSS, JS)")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExample:\n  %s --port 8081 --db-file /data/mybooks.db --web-dir ./static\n", os.Args[0])
	}

	flag.Parse()

	// --- Logging Setup ---
	// Using standard log package, outputting to stdout by default.
	// You could enhance this with structured logging (e.g., slog in Go 1.21+)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile) // Include file/line number for easier debugging
	log.Println("Starting Bookshelf application...")
	log.Printf("Configuration: Port=%d, DBFile=%s, WebDir=%s", *port, *dbFile, *webDir)


	// --- Dependency Injection ---
	// Initialize Database
	database, err := db.InitDB(*dbFile)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize database: %v", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Create Book Store
	bookStore := db.NewSQLiteBookStore(database)

	// Create API Handler
	apiHandler := api.NewAPIHandler(bookStore)

	// --- Router Setup ---
	// Ensure the web directory exists before setting up the router/server
	webDirAbs, err := filepath.Abs(*webDir)
	if err != nil {
		log.Fatalf("FATAL: Could not determine absolute path for web directory '%s': %v", *webDir, err)
	}
	
	if err := checkWebDir(*webDir); err != nil {
		log.Fatalf("FATAL: %v. Please create it or specify a valid directory using --web-dir.", err)
	}
	
	log.Printf("Serving static files from: %s", webDirAbs)

	router := api.SetupRouter(apiHandler, webDirAbs) // Pass absolute path

	// --- Server Setup ---
	serverAddr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting HTTP server on %s", serverAddr)

	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
		// Add timeouts for production readiness
		// ReadTimeout:  5 * time.Second,
		// WriteTimeout: 10 * time.Second,
		// IdleTimeout:  120 * time.Second,
	}

	// --- Start Server ---
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("FATAL: Could not start server: %v", err)
	}

	log.Println("Bookshelf application stopped.")
}
