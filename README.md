# vibe-books

"Vibe Coding" - AI generated code - to create a book tracking app:

## Stack

- web based
- backend API in golang
- storage in local sqlite DB

## Features

- add books to shelf (read, for later, etc)
# Bookshelf Web Application

A simple web application to manage a personal bookshelf, built with Go (backend) and vanilla JavaScript (frontend).

## Features

*   View books categorized by status: "Want to Read", "Currently Reading", "Read".
*   Add new books with title and author.
*   Update book status by dragging and dropping between columns.
*   Data persistence using SQLite.

## Project Structure

```
.
├── api/
│   └── handlers.go      # API request handlers
├── db/
│   └── database.go      # Database interaction logic (SQLite)
├── models/
│   └── book.go          # Data structures (Book model)
├── web/
│   ├── index.html       # Main HTML page
│   ├── style.css        # CSS styles
│   └── app.js           # Frontend JavaScript logic
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── main.go              # Main application entry point, server setup
└── bookshelf.db         # SQLite database file (created on first run)
```

## Setup and Running

1.  **Prerequisites:**
    *   Go (version 1.18 or later recommended)
    *   Git

2.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd bookshelf
    ```

3.  **Build the application:**
    This command compiles the Go code and places the executable in the project root (or GOPATH/bin depending on your setup). It also copies the `web` directory next to the executable, which is necessary for serving static files correctly after building.
    ```bash
    go build -o bookshelf .
    # Ensure the 'web' directory is copied next to the executable if building outside the source dir
    # If 'bookshelf' executable is in the root, this step might not be needed,
    # but it's good practice for deployment.
    # Example (if needed, adjust paths): cp -R web /path/to/executable/
    ```
    *Alternatively, for development:* You can run directly without building using:
    ```bash
    go run main.go
    ```
    This usually works fine as the `web` directory is found relative to the current working directory.

4.  **Run the application:**
    *If built:*
    ```bash
    ./bookshelf
    ```
    *If using `go run`:*
    ```bash
    go run main.go
    ```

5.  **Access the application:**
    Open your web browser and navigate to `http://localhost:8080`.

## API Documentation

The backend provides a simple REST API:

*   **`GET /api/books`**
    *   Description: Retrieves all books from the bookshelf.
    *   Response: `200 OK` with a JSON array of book objects.
        ```json
        [
          {
            "id": 1,
            "title": "The Go Programming Language",
            "author": "Alan A. A. Donovan, Brian W. Kernighan",
            "open_library_id": "OL26248016M",
            "status": "Read",
            "rating": 9,
            "comments": "Excellent reference.",
            "cover_url": null
          },
          // ... other books
        ]
        ```

*   **`POST /api/books`**
    *   Description: Adds a new book to the bookshelf, based on a selection from Open Library search.
    *   Request Body: JSON object representing the book details obtained from the search selection. `title`, `isbn`, and `open_library_id` are required. `status` defaults to "Want to Read". `author` and `cover_url` are optional but recommended.
        ```json
        {
          "title": "The Hobbit",
          "author": "J. R. R. Tolkien",
          "open_library_id": "OL7353617M",
          "isbn": "9780547928227",
          "cover_url": "https://covers.openlibrary.org/b/id/103187-M.jpg", // Optional
          "status": "Want to Read" // Optional, defaults if omitted
          // rating and comments will be added later
        }
        ```
    *   Response:
        *   `201 Created`: Success, returns the newly created book object (including its assigned `id`).
        *   `400 Bad Request`: Invalid JSON or missing required fields (`title`, `isbn`, `open_library_id`).
        *   `500 Internal Server Error`: Database error.

*   **`GET /api/search?q={query}`**
    *   Description: Searches Open Library for books matching the `query` (title/author). Returns a list of simplified book results that include an ISBN.
    *   Query Parameter: `q` - The search term (URL encoded).
    *   Response: `200 OK` with a JSON array of search result objects.
        ```json
        [
          {
            "open_library_id": "OL7353617M",
            "title": "The Hobbit",
            "author": "J. R. R. Tolkien",
            "isbn": "9780547928227", // Example ISBN
            "cover_url": "https://covers.openlibrary.org/b/id/103187-M.jpg" // Example cover URL
          },
          // ... other results
        ]
        ```
    *   Error Responses:
        *   `400 Bad Request`: Missing `q` parameter.
        *   `500 Internal Server Error`: Error creating/processing request or decoding response.
        *   `502 Bad Gateway`: Error contacting Open Library API.

*   **`PUT /api/books/{id}`**
    *   Description: Updates the **status** of a specific book (identified by `id`). Primarily used for drag-and-drop functionality.
    *   URL Parameter: `{id}` - The integer ID of the book to update.
    *   Request Body: JSON object containing the new status.
        ```json
        {
          "status": "Currently Reading" // Must be "Want to Read", "Currently Reading", or "Read"
        }
        ```
    *   Response:
        *   `200 OK`: Success.
        *   `400 Bad Request`: Invalid JSON, invalid status value, or invalid/missing ID.
        *   `404 Not Found`: Book with the specified ID does not exist.
        *   `500 Internal Server Error`: Database error.

*   **`PUT /api/books/{id}/details`**
    *   Description: Updates the **rating and/or comments** for a specific book.
    *   URL Parameter: `{id}` - The integer ID of the book to update.
    *   Request Body: JSON object containing the fields to update. Send `null` to clear a field. Rating must be 1-10 if provided.
        ```json
        {
          "rating": 8,          // Optional (number 1-10 or null)
          "comments": "Great read!" // Optional (string or null)
        }
        ```
    *   Response:
        *   `200 OK`: Success.
        *   `400 Bad Request`: Invalid JSON, invalid rating value, or invalid/missing ID.
        *   `404 Not Found`: Book with the specified ID does not exist.
        *   `500 Internal Server Error`: Database error.


## Future Enhancements

*   Implement delete book functionality.
*   Add search integration with Open Library API to fetch book details (author, cover).
*   Allow editing book details (rating, comments).
*   Add user authentication.
*   Improve error handling and user feedback on the frontend.
*   Add tests (unit and integration).
