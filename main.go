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
}

var db *sql.DB

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
       shelf TEXT NOT NULL,
       created_at DATETIME DEFAULT CURRENT_TIMESTAMP
   );`
   _, err := ExecSQL(query)
   return err
}

func booksHandler(w http.ResponseWriter, r *http.Request) {
   switch r.Method {
   case http.MethodGet:
       rows, err := QuerySQL("SELECT id, title, author, shelf FROM books")
       if err != nil {
           http.Error(w, "Database error", http.StatusInternalServerError)
           return
       }
       defer rows.Close()

       books := []Book{}
       for rows.Next() {
           var b Book
           if err := rows.Scan(&b.ID, &b.Title, &b.Author, &b.Shelf); err != nil {
               http.Error(w, "Database error", http.StatusInternalServerError)
               return
           }
           books = append(books, b)
       }
       writeJSON(w, books)
   case http.MethodPost:
       var b Book
       if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
           http.Error(w, "Invalid JSON", http.StatusBadRequest)
           return
       }
       if strings.TrimSpace(b.Title) == "" || strings.TrimSpace(b.Shelf) == "" {
           http.Error(w, "Missing title or shelf", http.StatusBadRequest)
           return
       }
       res, err := ExecSQL("INSERT INTO books (title, author, shelf) VALUES (?, ?, ?)", b.Title, b.Author, b.Shelf)
       if err != nil {
           http.Error(w, "Database error", http.StatusInternalServerError)
           return
       }
       id, err := res.LastInsertId()
       if err != nil {
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
       err := QueryRowSQL("SELECT id, title, author, shelf FROM books WHERE id = ?", id).
           Scan(&b.ID, &b.Title, &b.Author, &b.Shelf)
       if err == sql.ErrNoRows {
           http.Error(w, "Not found", http.StatusNotFound)
           return
       } else if err != nil {
           http.Error(w, "Database error", http.StatusInternalServerError)
           return
       }
       writeJSON(w, b)
   case http.MethodPut:
       var b Book
       if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
           http.Error(w, "Invalid JSON", http.StatusBadRequest)
           return
       }
       if strings.TrimSpace(b.Title) == "" || strings.TrimSpace(b.Shelf) == "" {
           http.Error(w, "Missing title or shelf", http.StatusBadRequest)
           return
       }
       _, err := ExecSQL("UPDATE books SET title = ?, author = ?, shelf = ? WHERE id = ?", b.Title, b.Author, b.Shelf, id)
       if err != nil {
           http.Error(w, "Database error", http.StatusInternalServerError)
           return
       }
       b.ID = id
       writeJSON(w, b)
   case http.MethodDelete:
       _, err := ExecSQL("DELETE FROM books WHERE id = ?", id)
       if err != nil {
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
       http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
       return
   }
   rows, err := QuerySQL("SELECT DISTINCT shelf FROM books")
   if err != nil {
       http.Error(w, "Database error", http.StatusInternalServerError)
       return
   }
   defer rows.Close()

   shelves := []string{}
   for rows.Next() {
       var s string
       if err := rows.Scan(&s); err != nil {
           http.Error(w, "Database error", http.StatusInternalServerError)
           return
       }
       shelves = append(shelves, s)
   }
   writeJSON(w, shelves)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
   w.Header().Set("Content-Type", "application/json")
   json.NewEncoder(w).Encode(v)
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