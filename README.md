# Bookshelf Web Application

A web application to manage a personal bookshelf, built with Go (backend API) and vanilla JavaScript (frontend) using the Pico.css framework and SortableJS for drag-and-drop. Books are searched via the Open Library API and stored locally in an SQLite database.

## Features

*   **View Books:** Display books categorized by status: "Want to Read", "Currently Reading", "Read".
*   **Search & Add Books:** Search the Open Library API by title/author and add selected books to the "Want to Read" shelf.
*   **Update Status:** Drag and drop books between status columns to update their status.
*   **Edit Details:** Update a book's rating (1-10) and add personal comments via a modal dialog.
*   **Data Persistence:** Book data is stored in a local SQLite database (`bookshelf.db` by default).
*   **Basic Logging:** HTTP requests and SQL operations are logged to standard output.

## Project Structure

```
bookshelf/
├── cmd/
│   └── server/
│       └── main.go         # Entrypoint: setup server, db, routes, flags
├── internal/
│   ├── api/
│   │   ├── handler.go      # HTTP handlers (GET /books, POST /books, PUT /books/{id}, etc.)
│   │   └── routes.go       # Router setup (using gorilla/mux), middleware
│   ├── db/
│   │   ├── db.go           # DB connection (SQLite) and schema creation
│   │   └── book_store.go   # CRUD operations interface and implementation for books
│   └── model/
│       └── book.go         # Book struct, Status enum, validation
├── web/                    # Static frontend assets
│   ├── index.html          # Main HTML page (using Pico.css)
│   ├── main.js             # Frontend JavaScript logic (API calls, DOM manipulation, SortableJS)
│   └── style.css           # Custom CSS styles (minimal, complements Pico.css)
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
└── bookshelf.db            # SQLite database file (created on first run if it doesn't exist)
└── README.md               # This file
```

## Setup and Running

1.  **Prerequisites:**
    *   Go (version 1.21 or later recommended)
    *   Git

2.  **Clone the repository:**
    ```bash
    # Replace with your actual repository URL if applicable
    git clone https://github.com/your-username/bookshelf.git
    cd bookshelf
    ```

3.  **Replace Module Path:**
    *   Edit the `go.mod` file and replace `github.com/your-username/bookshelf` with your actual Go module path if you plan to host or modify it significantly.
    *   Update the import paths in `.go` files under `internal/` and `cmd/` accordingly if you changed the module path.

4.  **Install Dependencies:**
    ```bash
    go mod tidy
    ```

5.  **Build the application (Optional):**
    This command compiles the Go code into a single executable named `bookshelf` in the project root.
    ```bash
    go build -o bookshelf ./cmd/server/main.go
    ```
    *Note:* The server expects the `web` directory to be present in the *current working directory* when running the executable, unless specified otherwise with the `--web-dir` flag.

6.  **Run the application:**
    *   **Using `go run` (for development):**
        This command compiles and runs the application directly. The `web` directory and `bookshelf.db` (if it exists) will be relative to the project root.
        ```bash
        go run ./cmd/server/main.go
        ```
    *   **Using the built executable:**
        Make sure you are in the project root directory (where the `web` directory is located).
        ```bash
        ./bookshelf
        ```
    *   **Command-line Flags:**
        *   `--port <number>`: Specify the port number (default: `8080`).
        *   `--db-file <path>`: Specify the path to the SQLite database file (default: `./bookshelf.db`).
        *   `--web-dir <path>`: Specify the directory containing static web assets (default: `./web`).
        *   `--help`: Show help message.
        Example:
        ```bash
        go run ./cmd/server/main.go --port 9000 --db-file /data/my_books.db
        ./bookshelf --port 9000 --db-file /data/my_books.db
        ```

7.  **Access the application:**
    Open your web browser and navigate to `http://localhost:<port>` (e.g., `http://localhost:8080` if using the default port).

## API Documentation

The backend provides a RESTful API under the `/api` prefix:

*   **`GET /api/books`**
    *   Description: Retrieves all books currently on the bookshelf, ordered by title.
    *   Response: `200 OK` with a JSON array of book objects.
        ```json
        [
          {
            "id": 1,
            "title": "The Go Programming Language",
            "author": "Alan A. A. Donovan, Brian W. Kernighan",
            "open_library_id": "OL26248016M",
            "isbn": "9780134190440",
            "status": "Read",
            "rating": 9, // Can be null
            "comments": "Excellent reference.", // Can be null
            "cover_url": "https://covers.openlibrary.org/b/id/8264891-M.jpg" // Can be null
          },
          // ... other books
        ]
        ```

*   **`POST /api/books`**
    *   Description: Adds a new book to the bookshelf, typically based on a selection from an Open Library search result. The book is added with status "Want to Read" by default.
    *   Request Body: JSON object with book details. `title` and `open_library_id` are required. `author`, `isbn`, and `cover_url` are recommended. `status` can be optionally provided but defaults to "Want to Read". `rating` and `comments` are ignored (set to null initially).
        ```json
        {
          "title": "The Hobbit",
          "author": "J. R. R. Tolkien",
          "open_library_id": "OL7353617M",
          "isbn": "9780547928227", // Optional
          "cover_url": "https://covers.openlibrary.org/b/id/103187-M.jpg" // Optional
          // "status": "Want to Read" // Optional, defaults if omitted
        }
        ```
    *   Response:
        *   `201 Created`: Success, returns the newly created book object (including its assigned `id` and default status).
        *   `400 Bad Request`: Invalid JSON, missing required fields (`title`, `open_library_id`), or validation error.
        *   `500 Internal Server Error`: Database error (e.g., UNIQUE constraint violation on `open_library_id`).

*   **`GET /api/search?q={query}`**
    *   Description: Searches the Open Library API for books matching the `query` (title/author). Returns a simplified list of results suitable for selection.
    *   Query Parameter: `q` - The search term (URL encoded).
    *   Response: `200 OK` with a JSON array of search result objects.
        ```json
        [
          {
            "open_library_id": "OL7353617M",
            "title": "The Hobbit",
            "author": "J. R. R. Tolkien",
            "isbn": "9780547928227", // First ISBN-13 or ISBN-10 found
            "cover_url": "https://covers.openlibrary.org/b/id/103187-M.jpg" // Medium cover URL if available
          },
          // ... other results (limit 20)
        ]
        ```
    *   Error Responses:
        *   `400 Bad Request`: Missing `q` parameter.
        *   `500 Internal Server Error`: Error creating/processing the request or decoding the Open Library response.
        *   `502 Bad Gateway`: Error contacting the Open Library API or receiving an invalid response from it.

*   **`PUT /api/books/{id}`**
    *   Description: Updates the **status** of a specific book (identified by its integer `id`). Used by the drag-and-drop feature.
    *   URL Parameter: `{id}` - The integer ID of the book to update.
    *   Request Body: JSON object containing the new status.
        ```json
        {
          "status": "Currently Reading" // Must be "Want to Read", "Currently Reading", or "Read"
        }
        ```
    *   Response:
        *   `200 OK`: Success, returns `{"message": "Book status updated successfully"}`.
        *   `400 Bad Request`: Invalid JSON, invalid status value, or invalid ID format.
        *   `404 Not Found`: Book with the specified ID does not exist.
        *   `500 Internal Server Error`: Database error during update.

*   **`PUT /api/books/{id}/details`**
    *   Description: Updates the **rating and/or comments** for a specific book.
    *   URL Parameter: `{id}` - The integer ID of the book to update.
    *   Request Body: JSON object containing the fields to update. Omit fields to leave them unchanged. Send `null` or an empty string for a field to clear its value in the database. Rating must be 1-10 if provided.
        ```json
        // Example: Update rating only
        { "rating": 8 }

        // Example: Update comments only
        { "comments": "A fantastic read!" }

        // Example: Update both
        { "rating": 9, "comments": "Highly recommended." }

        // Example: Clear rating, update comments
        { "rating": null, "comments": "Finished, okay." }

        // Example: Clear comments
        { "comments": null }
        ```
    *   Response:
        *   `200 OK`: Success, returns `{"message": "Book details updated successfully"}`.
        *   `400 Bad Request`: Invalid JSON, invalid rating value (not 1-10 or null), or invalid ID format.
        *   `404 Not Found`: Book with the specified ID does not exist.
        *   `500 Internal Server Error`: Database error during update.

## Future Enhancements

*   Implement book deletion functionality (`DELETE /api/books/{id}`).
*   Add user authentication/accounts.
*   Improve frontend UI/UX (e.g., better loading indicators, error handling display).
*   Add pagination for large bookshelves.
*   Implement more robust error handling and reporting.
*   Add unit and integration tests.
*   Consider configuration file instead of only flags.
*   Optionally allow manual book entry (without Open Library search).
