package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ericdahl/bookshelf/internal/model"
)

// MockBookStore is a mock implementation of the BookStore interface for testing
type MockBookStore struct {
	Books       []model.Book
	LastAddedID int64
	GetBooksErr error
	AddBookErr  error
	GetBookErr  error
	UpdateErr   error
	DeleteErr   error
}

func (m *MockBookStore) GetBooks() ([]model.Book, error) {
	return m.Books, m.GetBooksErr
}

func (m *MockBookStore) AddBook(book *model.Book) (int64, error) {
	if m.AddBookErr != nil {
		return 0, m.AddBookErr
	}
	m.LastAddedID++
	book.ID = m.LastAddedID
	m.Books = append(m.Books, *book)
	return book.ID, nil
}

func (m *MockBookStore) GetBookByID(id int64) (*model.Book, error) {
	if m.GetBookErr != nil {
		return nil, m.GetBookErr
	}
	for _, book := range m.Books {
		if book.ID == id {
			return &book, nil
		}
	}
	return nil, m.GetBookErr
}

func (m *MockBookStore) UpdateBookStatus(id int64, status model.BookStatus) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	for i, book := range m.Books {
		if book.ID == id {
			m.Books[i].Status = status
			return nil
		}
	}
	return nil
}

func (m *MockBookStore) UpdateBookDetails(id int64, rating *int, comments *string) error {
	if m.UpdateErr != nil {
		return m.UpdateErr
	}
	for i, book := range m.Books {
		if book.ID == id {
			m.Books[i].Rating = rating
			m.Books[i].Comments = comments
			return nil
		}
	}
	return nil
}

func (m *MockBookStore) DeleteBook(id int64) error {
	if m.DeleteErr != nil {
		return m.DeleteErr
	}
	for i, book := range m.Books {
		if book.ID == id {
			// Remove the book at index i
			m.Books = append(m.Books[:i], m.Books[i+1:]...)
			return nil
		}
	}
	return nil
}

// TestGetBooksHandlerWithMock tests the GetBooksHandler with a mock store
func TestGetBooksHandlerWithMock(t *testing.T) {
	// Set up mock store with predefined books
	comments := "Test comments"
	rating := 8
	coverURL := "http://example.com/cover.jpg"
	mockStore := &MockBookStore{
		Books: []model.Book{
			{
				ID:            1,
				Title:         "Test Book 1",
				Author:        "Test Author 1",
				OpenLibraryID: "OL123M",
				Status:        model.StatusWantToRead,
				Rating:        &rating,
				Comments:      &comments,
				CoverURL:      &coverURL,
			},
			{
				ID:            2,
				Title:         "Test Book 2",
				Author:        "Test Author 2",
				OpenLibraryID: "OL456M",
				Status:        model.StatusCurrentlyReading,
			},
		},
	}

	// Create handler with mock store
	handler := NewAPIHandler(mockStore)

	// Create request
	req := httptest.NewRequest("GET", "/api/books", nil)
	w := httptest.NewRecorder()

	// Call handler directly
	handler.GetBooksHandler(w, req)

	// Check result
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}