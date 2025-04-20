package main

import (
	"time"
)

// DbBook represents a book as stored in the database
type DbBook struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	Rating    float64   `json:"rating"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Shelf represents a book shelf or collection
type Shelf struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// GetAllBooks retrieves all books from the database
func GetAllBooks() ([]DbBook, error) {
	LogQuery("SELECT id, title, author, rating, created_at, updated_at FROM books")
	rows, err := DB.Query("SELECT id, title, author, rating, created_at, updated_at FROM books")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := []DbBook{}
	for rows.Next() {
		var book DbBook
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Rating, &book.CreatedAt, &book.UpdatedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}

// GetBookByID retrieves a book by its ID
func GetBookByID(id string) (DbBook, error) {
	var book DbBook
	LogQuery("SELECT id, title, author, rating, created_at, updated_at FROM books WHERE id = ?", id)
	err := DB.QueryRow("SELECT id, title, author, rating, created_at, updated_at FROM books WHERE id = ?", id).
		Scan(&book.ID, &book.Title, &book.Author, &book.Rating, &book.CreatedAt, &book.UpdatedAt)
	if err != nil {
		return DbBook{}, err
	}

	return book, nil
}

// AddBook adds a new book to the database
func AddBook(book DbBook) error {
	now := time.Now()
	book.CreatedAt = now
	book.UpdatedAt = now

	LogQuery("INSERT INTO books (id, title, author, rating, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		book.ID, book.Title, book.Author, book.Rating, book.CreatedAt, book.UpdatedAt)
	_, err := DB.Exec("INSERT INTO books (id, title, author, rating, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		book.ID, book.Title, book.Author, book.Rating, book.CreatedAt, book.UpdatedAt)
	return err
}

// UpdateBook updates an existing book
func UpdateBook(book DbBook) error {
	book.UpdatedAt = time.Now()

	LogQuery("UPDATE books SET title = ?, author = ?, rating = ?, updated_at = ? WHERE id = ?",
		book.Title, book.Author, book.Rating, book.UpdatedAt, book.ID)
	_, err := DB.Exec("UPDATE books SET title = ?, author = ?, rating = ?, updated_at = ? WHERE id = ?",
		book.Title, book.Author, book.Rating, book.UpdatedAt, book.ID)
	return err
}

// DeleteBook removes a book from the database
func DeleteBook(id string) error {
	// First delete book-shelf associations
	LogQuery("DELETE FROM book_shelves WHERE book_id = ?", id)
	_, err := DB.Exec("DELETE FROM book_shelves WHERE book_id = ?", id)
	if err != nil {
		return err
	}

	// Then delete the book
	LogQuery("DELETE FROM books WHERE id = ?", id)
	_, err = DB.Exec("DELETE FROM books WHERE id = ?", id)
	return err
}

// GetShelves retrieves all shelves
func GetShelves() ([]Shelf, error) {
	LogQuery("SELECT id, name, created_at FROM shelves")
	rows, err := DB.Query("SELECT id, name, created_at FROM shelves")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shelves := []Shelf{}
	for rows.Next() {
		var shelf Shelf
		err := rows.Scan(&shelf.ID, &shelf.Name, &shelf.CreatedAt)
		if err != nil {
			return nil, err
		}
		shelves = append(shelves, shelf)
	}

	return shelves, nil
}

// AddBookToShelf adds a book to a specific shelf
func AddBookToShelf(bookID, shelfID string) error {
	now := time.Now()

	LogQuery("INSERT OR REPLACE INTO book_shelves (book_id, shelf_id, added_at) VALUES (?, ?, ?)",
		bookID, shelfID, now)
	_, err := DB.Exec("INSERT OR REPLACE INTO book_shelves (book_id, shelf_id, added_at) VALUES (?, ?, ?)",
		bookID, shelfID, now)
	return err
}

// RemoveBookFromShelf removes a book from a specific shelf
func RemoveBookFromShelf(bookID, shelfID string) error {
	LogQuery("DELETE FROM book_shelves WHERE book_id = ? AND shelf_id = ?", bookID, shelfID)
	_, err := DB.Exec("DELETE FROM book_shelves WHERE book_id = ? AND shelf_id = ?", bookID, shelfID)
	return err
}

// GetBooksInShelf retrieves all books in a specific shelf
func GetBooksInShelf(shelfID string) ([]DbBook, error) {
	LogQuery(`
		SELECT b.id, b.title, b.author, b.rating, b.created_at, b.updated_at
		FROM books b
		JOIN book_shelves bs ON b.id = bs.book_id
		WHERE bs.shelf_id = ?
	`, shelfID)

	rows, err := DB.Query(`
		SELECT b.id, b.title, b.author, b.rating, b.created_at, b.updated_at
		FROM books b
		JOIN book_shelves bs ON b.id = bs.book_id
		WHERE bs.shelf_id = ?
	`, shelfID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := []DbBook{}
	for rows.Next() {
		var book DbBook
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Rating, &book.CreatedAt, &book.UpdatedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}

	return books, nil
}
