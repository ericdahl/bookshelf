package main

import (
	"database/sql"
	"time"
)

// DbBook represents a book as stored in the database
type DbBook struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Author        string    `json:"author"`
	Rating        float64   `json:"rating"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	OpenLibraryID string    `json:"open_library_id,omitempty"`
	ISBN          string    `json:"isbn,omitempty"`
	CoverID       string    `json:"cover_id,omitempty"`
	PublishYear   int       `json:"publish_year,omitempty"`
	Publisher     string    `json:"publisher,omitempty"`
	PageCount     int       `json:"page_count,omitempty"`
	Description   string    `json:"description,omitempty"`
}

// Shelf represents a book shelf or collection
type Shelf struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// GetAllBooks retrieves all books from the database
func GetAllBooks() ([]DbBook, error) {
	LogQuery("SELECT id, title, author, rating, created_at, updated_at, open_library_id, isbn, cover_id, publish_year, publisher, page_count, description FROM books")
	rows, err := DB.Query("SELECT id, title, author, rating, created_at, updated_at, open_library_id, isbn, cover_id, publish_year, publisher, page_count, description FROM books")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	books := []DbBook{}
	for rows.Next() {
		var book DbBook
		var openLibraryID, isbn, coverID, publisher, description sql.NullString
		var publishYear, pageCount sql.NullInt64
		
		err := rows.Scan(
			&book.ID, &book.Title, &book.Author, &book.Rating, 
			&book.CreatedAt, &book.UpdatedAt, &openLibraryID, &isbn,
			&coverID, &publishYear, &publisher, &pageCount, &description,
		)
		if err != nil {
			return nil, err
		}
		
		// Handle nullable fields
		if openLibraryID.Valid {
			book.OpenLibraryID = openLibraryID.String
		}
		if isbn.Valid {
			book.ISBN = isbn.String
		}
		if coverID.Valid {
			book.CoverID = coverID.String
		}
		if publishYear.Valid {
			book.PublishYear = int(publishYear.Int64)
		}
		if publisher.Valid {
			book.Publisher = publisher.String
		}
		if pageCount.Valid {
			book.PageCount = int(pageCount.Int64)
		}
		if description.Valid {
			book.Description = description.String
		}
		
		books = append(books, book)
	}

	return books, nil
}

// GetBookByID retrieves a book by its ID
func GetBookByID(id string) (DbBook, error) {
	var book DbBook
	var openLibraryID, isbn, coverID, publisher, description sql.NullString
	var publishYear, pageCount sql.NullInt64
	
	LogQuery("SELECT id, title, author, rating, created_at, updated_at, open_library_id, isbn, cover_id, publish_year, publisher, page_count, description FROM books WHERE id = ?", id)
	err := DB.QueryRow("SELECT id, title, author, rating, created_at, updated_at, open_library_id, isbn, cover_id, publish_year, publisher, page_count, description FROM books WHERE id = ?", id).
		Scan(
			&book.ID, &book.Title, &book.Author, &book.Rating, 
			&book.CreatedAt, &book.UpdatedAt, &openLibraryID, &isbn,
			&coverID, &publishYear, &publisher, &pageCount, &description,
		)
	if err != nil {
		return DbBook{}, err
	}
	
	// Handle nullable fields
	if openLibraryID.Valid {
		book.OpenLibraryID = openLibraryID.String
	}
	if isbn.Valid {
		book.ISBN = isbn.String
	}
	if coverID.Valid {
		book.CoverID = coverID.String
	}
	if publishYear.Valid {
		book.PublishYear = int(publishYear.Int64)
	}
	if publisher.Valid {
		book.Publisher = publisher.String
	}
	if pageCount.Valid {
		book.PageCount = int(pageCount.Int64)
	}
	if description.Valid {
		book.Description = description.String
	}

	return book, nil
}

// AddBook adds a new book to the database
func AddBook(book DbBook) error {
	now := time.Now()
	book.CreatedAt = now
	book.UpdatedAt = now

	LogQuery(`INSERT INTO books (
		id, title, author, rating, created_at, updated_at, 
		open_library_id, isbn, cover_id, publish_year, publisher, page_count, description
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		book.ID, book.Title, book.Author, book.Rating, book.CreatedAt, book.UpdatedAt,
		book.OpenLibraryID, book.ISBN, book.CoverID, book.PublishYear, book.Publisher, book.PageCount, book.Description)
	
	_, err := DB.Exec(`INSERT INTO books (
		id, title, author, rating, created_at, updated_at, 
		open_library_id, isbn, cover_id, publish_year, publisher, page_count, description
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		book.ID, book.Title, book.Author, book.Rating, book.CreatedAt, book.UpdatedAt,
		book.OpenLibraryID, book.ISBN, book.CoverID, book.PublishYear, book.Publisher, book.PageCount, book.Description)
	return err
}

// UpdateBook updates an existing book
func UpdateBook(book DbBook) error {
	book.UpdatedAt = time.Now()

	LogQuery(`UPDATE books SET 
		title = ?, author = ?, rating = ?, updated_at = ?,
		open_library_id = ?, isbn = ?, cover_id = ?, publish_year = ?,
		publisher = ?, page_count = ?, description = ?
		WHERE id = ?`,
		book.Title, book.Author, book.Rating, book.UpdatedAt,
		book.OpenLibraryID, book.ISBN, book.CoverID, book.PublishYear,
		book.Publisher, book.PageCount, book.Description, book.ID)
	
	_, err := DB.Exec(`UPDATE books SET 
		title = ?, author = ?, rating = ?, updated_at = ?,
		open_library_id = ?, isbn = ?, cover_id = ?, publish_year = ?,
		publisher = ?, page_count = ?, description = ?
		WHERE id = ?`,
		book.Title, book.Author, book.Rating, book.UpdatedAt,
		book.OpenLibraryID, book.ISBN, book.CoverID, book.PublishYear,
		book.Publisher, book.PageCount, book.Description, book.ID)
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
		SELECT b.id, b.title, b.author, b.rating, b.created_at, b.updated_at, 
		b.open_library_id, b.isbn, b.cover_id, b.publish_year, b.publisher, b.page_count, b.description
		FROM books b
		JOIN book_shelves bs ON b.id = bs.book_id
		WHERE bs.shelf_id = ?
	`, shelfID)

	rows, err := DB.Query(`
		SELECT b.id, b.title, b.author, b.rating, b.created_at, b.updated_at,
		b.open_library_id, b.isbn, b.cover_id, b.publish_year, b.publisher, b.page_count, b.description
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
		var openLibraryID, isbn, coverID, publisher, description sql.NullString
		var publishYear, pageCount sql.NullInt64
		
		err := rows.Scan(
			&book.ID, &book.Title, &book.Author, &book.Rating, 
			&book.CreatedAt, &book.UpdatedAt, &openLibraryID, &isbn,
			&coverID, &publishYear, &publisher, &pageCount, &description,
		)
		if err != nil {
			return nil, err
		}
		
		// Handle nullable fields
		if openLibraryID.Valid {
			book.OpenLibraryID = openLibraryID.String
		}
		if isbn.Valid {
			book.ISBN = isbn.String
		}
		if coverID.Valid {
			book.CoverID = coverID.String
		}
		if publishYear.Valid {
			book.PublishYear = int(publishYear.Int64)
		}
		if publisher.Valid {
			book.Publisher = publisher.String
		}
		if pageCount.Valid {
			book.PageCount = int(pageCount.Int64)
		}
		if description.Valid {
			book.Description = description.String
		}
		
		books = append(books, book)
	}

	return books, nil
}
