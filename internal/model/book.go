package model

// BookStatus represents the reading status of a book.
type BookStatus string

const (
	StatusWantToRead     BookStatus = "Want to Read"
	StatusCurrentlyReading BookStatus = "Currently Reading"
	StatusRead           BookStatus = "Read"
)

// IsValid checks if the status string is one of the predefined valid statuses.
func (s BookStatus) IsValid() bool {
	switch s {
	case StatusWantToRead, StatusCurrentlyReading, StatusRead:
		return true
	default:
		return false
	}
}

// BookType represents the type of book (paper book or audiobook).
type BookType string

const (
	TypeBook      BookType = "book"
	TypeAudiobook BookType = "audiobook"
)

// IsValid checks if the type is one of the predefined valid types.
func (t BookType) IsValid() bool {
	switch t {
	case TypeBook, TypeAudiobook:
		return true
	default:
		return false
	}
}

// Book represents a book entry in the bookshelf.
type Book struct {
	ID            int64      `json:"id"`
	Title         string     `json:"title"`
	Author        string     `json:"author"`
	OpenLibraryID string     `json:"open_library_id"` // e.g., OL7353617M
	ISBN          string     `json:"isbn,omitempty"`  // Optional, but useful
	Status        BookStatus `json:"status"`
	Type          BookType   `json:"type"`            // "book" or "audiobook"
	Rating        *int       `json:"rating,omitempty"`   // Pointer to allow null, 1-10
	Comments      *string    `json:"comments,omitempty"` // Pointer to allow null
	CoverURL      *string    `json:"cover_url,omitempty"` // URL for the book cover image
	Series        *string    `json:"series,omitempty"`    // Name of the series (optional)
	SeriesIndex   *int       `json:"series_index,omitempty"` // Position in the series (optional)
}

// Validate checks the book data for validity.
// Checks Rating range, Status and Type values.
func (b *Book) Validate() error {
	if b.Rating != nil && (*b.Rating < 1 || *b.Rating > 10) {
		// Consider using a custom error type or fmt.Errorf
		return &ValidationError{"rating must be between 1 and 10"}
	}
	if !b.Status.IsValid() {
		return &ValidationError{"invalid status provided"}
	}
	if b.Type == "" {
		// Default to "book" if not specified
		b.Type = TypeBook
	} else if !b.Type.IsValid() {
		return &ValidationError{"invalid type provided, must be 'book' or 'audiobook'"}
	}
	// Add other validations as needed (e.g., Title required)
	return nil
}

// ValidationError represents an error during model validation.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
