# Vibe Books API

A RESTful API for managing books and shelves with both in-memory and database storage options.

## Features

- Book management (CRUD operations)
- Shelf management (predefined and custom)
- Book ratings
- REST API with JSON
- SQLite database persistence
- SQL execution logging

## API Endpoints

### In-Memory API (for testing)

- `GET /api/books` - Get all books
- `GET /api/books/{id}` - Get a specific book
- `POST /api/books` - Create a new book
- `PUT /api/books/{id}` - Update a book
- `DELETE /api/books/{id}` - Delete a book

### Database API (persistent storage)

- `GET /api/db/books` - Get all books
- `GET /api/db/books/{id}` - Get a specific book
- `POST /api/db/books` - Create a new book
- `PUT /api/db/books/{id}` - Update a book
- `DELETE /api/db/books/{id}` - Delete a book

### Shelves API

- `GET /api/shelves` - Get all shelves
- `GET /api/shelves/{id}/books` - Get all books in a shelf
- `POST /api/shelves/{shelfId}/books/{bookId}` - Add a book to a shelf
- `DELETE /api/shelves/{shelfId}/books/{bookId}` - Remove a book from a shelf

## Setup

```bash
# Install dependencies
go mod download

# Run the server
go run .
```

## Testing

```bash
# Run all tests
go test -v ./...
```

## Project Structure

- `main.go`: Main application, API routes
- `models.go`: Data models and database access functions
- `handlers.go`: HTTP request handlers
- `database.go`: Database setup and SQL logging
- `utils.go`: Utility functions

## Implementation Details

- RESTful API with JSON responses
- Clean modular code structure
- SQLite database with proper schema
- Predefined shelves
- Book rating system