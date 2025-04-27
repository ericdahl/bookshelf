package main

import (
	"flag"
	"fmt"
	"log/slog"
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
	verbose := flag.Bool("verbose", false, "Enable verbose logging (Debug level)")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExample:\n  %s --port 8081 --db-file /data/mybooks.db --web-dir ./static --verbose\n", os.Args[0])
	}

	flag.Parse()

	// --- Logging Setup ---
	// Using slog structured logging with JSON handler
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting Bookshelf application...")
	slog.Info("Configuration",
		"port", *port,
		"dbFile", *dbFile,
		"webDir", *webDir,
		"verbose", *verbose)

	// --- Dependency Injection ---
	// Initialize Database
	database, err := db.InitDB(*dbFile)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer func() {
		slog.Info("Closing database connection...")
		if err := database.Close(); err != nil {
			slog.Error("Error closing database", "error", err)
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
		slog.Error("Could not determine absolute path for web directory", "webDir", *webDir, "error", err)
		os.Exit(1)
	}

	if err := checkWebDir(*webDir); err != nil {
		slog.Error("Web directory error", "error", err, "help", "Please create it or specify a valid directory using --web-dir.")
		os.Exit(1)
	}

	slog.Info("Serving static files", "path", webDirAbs)

	router := api.SetupRouter(apiHandler, webDirAbs) // Pass absolute path

	// --- Server Setup ---
	serverAddr := fmt.Sprintf(":%d", *port)
	slog.Info("Starting HTTP server", "address", serverAddr)

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
		slog.Error("Could not start server", "error", err)
		os.Exit(1)
	}

	slog.Info("Bookshelf application stopped")
}
