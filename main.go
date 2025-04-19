package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Book represents a book entry
type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Shelf  string `json:"shelf"`
	Rating int    `json:"rating"`
}

var db *sql.DB

// Predefined allowed shelves
var allowedShelves = []string{"Read", "To-Read"}

// isValidShelf returns true if s is one of the allowedShelves
func isValidShelf(s string) bool {
	for _, sh := range allowedShelves {
		if sh == s {
			return true
		}
	}
	return false
}

// ExecSQL executes a SQL statement and logs the query and its arguments.
func ExecSQL(query string, args ...interface{}) (sql.Result, error) {
	log.Printf("[SQL Exec] %s; args: %v", query, args)
	return db.Exec(query, args...)
}

// QuerySQL executes a SQL query that returns rows and logs the query and its arguments.
func QuerySQL(query string, args ...interface{}) (*sql.Rows, error) {
	log.Printf("[SQL Query] %s; args: %v", query, args)
	return db.Query(query, args...)
}

// QueryRowSQL executes a SQL query expected to return at most one row and logs the query and its arguments.
func QueryRowSQL(query string, args ...interface{}) *sql.Row {
	log.Printf("[SQL QueryRow] %s; args: %v", query, args)
	return db.QueryRow(query, args...)
}

func main() {
	log.SetOutput(os.Stdout)

	// Parse command-line flags
	var port int
	var help bool
	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: %s [options]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if help {
		flag.Usage()
		os.Exit(0)
	}

	var err error
	db, err = sql.Open("sqlite3", "./books.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	if err := initDB(); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/books", booksHandler)
	mux.HandleFunc("/api/books/", bookHandler)
	mux.HandleFunc("/api/shelves", shelvesHandler)

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server started at %s", addr)
	if err := http.ListenAndServe(addr, loggingMiddleware(mux)); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func initDB() error {
	query := `
   CREATE TABLE IF NOT EXISTS books (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       title TEXT NOT NULL,
       author TEXT,
       rating INTEGER DEFAULT 0,
       shelf TEXT NOT NULL,
       created_at DATETIME DEFAULT CURRENT_TIMESTAMP
   );`
	_, err := ExecSQL(query)
	if err != nil {
		return err
	}

	//   // Add rating column if it does not already exist
	//   rows, err := db.Query("PRAGMA table_info(books)")
	//   if err != nil {
	//       return err
	//   }
	//   defer rows.Close()
	//   exists := false
	//   for rows.Next() {
	//       var cid int
	//       var name string
	//       var ctype string
	//       var notnull int
	//       var dfltValue sql.NullString
	//       var pk int
	//       if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
	//           return err
	//       }
	//       if name == "rating" {
	//           exists = true
	//           break
	//       }
	//   }
	//   if !exists {
	//       if _, err := ExecSQL("ALTER TABLE books ADD COLUMN rating INTEGER DEFAULT 0"); err != nil {
	//           return err
	//       }
	//   }
	return nil
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := QuerySQL("SELECT id, title, author, shelf, rating FROM books")
		if err != nil {
			log.Printf("booksHandler GET: QuerySQL error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		books := []Book{}
		for rows.Next() {
			var b Book
			if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Shelf, &b.Rating); err != nil {
				log.Printf("booksHandler GET: rows.Scan error: %v", err)
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			books = append(books, b)
		}
		writeJSON(w, books)
	case http.MethodPost:
		var b Book
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			log.Printf("booksHandler POST: JSON decode error: %v", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		b.Title = strings.TrimSpace(b.Title)
		b.Shelf = strings.TrimSpace(b.Shelf)
		// Validate rating (0 = unrated, otherwise 1-10)
		if b.Rating < 0 || b.Rating > 10 {
			http.Error(w, "Invalid rating", http.StatusBadRequest)
			return
		}
		if b.Title == "" || b.Shelf == "" {
			http.Error(w, "Missing title or shelf", http.StatusBadRequest)
			return
		}
		if !isValidShelf(b.Shelf) {
			http.Error(w, "Invalid shelf", http.StatusBadRequest)
			return
		}
		res, err := ExecSQL("INSERT INTO books (title, author, shelf, rating) VALUES (?, ?, ?, ?)", b.Title, b.Author, b.Shelf, b.Rating)
		if err != nil {
			log.Printf("booksHandler POST: ExecSQL error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		id, err := res.LastInsertId()
		if err != nil {
			log.Printf("booksHandler POST: LastInsertId error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		b.ID = int(id)
		writeJSON(w, b)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/books/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid book id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		var b Book
		err := QueryRowSQL("SELECT id, title, author, shelf, rating FROM books WHERE id = ?", id).
			Scan(&b.ID, &b.Title, &b.Author, &b.Shelf, &b.Rating)
		if err == sql.ErrNoRows {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		} else if err != nil {
			log.Printf("bookHandler GET: QueryRow error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, b)
	case http.MethodPut:
		var b Book
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			log.Printf("bookHandler PUT: JSON decode error: %v", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		b.Title = strings.TrimSpace(b.Title)
		b.Shelf = strings.TrimSpace(b.Shelf)
		if b.Title == "" || b.Shelf == "" {
			http.Error(w, "Missing title or shelf", http.StatusBadRequest)
			return
		}
		if !isValidShelf(b.Shelf) {
			http.Error(w, "Invalid shelf", http.StatusBadRequest)
			return
		}
		// Validate rating (0 = unrated, otherwise 1-10)
		if b.Rating < 0 || b.Rating > 10 {
			http.Error(w, "Invalid rating", http.StatusBadRequest)
			return
		}
		_, err := ExecSQL("UPDATE books SET title = ?, author = ?, shelf = ?, rating = ? WHERE id = ?", b.Title, b.Author, b.Shelf, b.Rating, id)
		if err != nil {
			log.Printf("bookHandler PUT: ExecSQL error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		b.ID = id
		writeJSON(w, b)
	case http.MethodDelete:
		_, err := ExecSQL("DELETE FROM books WHERE id = ?", id)
		if err != nil {
			log.Printf("bookHandler DELETE: ExecSQL error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func shelvesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("shelvesHandler: method not allowed %s", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Return predefined shelves
	writeJSON(w, allowedShelves)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}

// loggingResponseWriter captures HTTP status codes
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures status code
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs requests with status and duration
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		log.Printf("%s %s %d %s", r.Method, r.RequestURI, lrw.statusCode, duration)
	})
}
