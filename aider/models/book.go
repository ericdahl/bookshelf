package models

// Book represents a book in the bookshelf.
type Book struct {
	ID            int64   `json:"id"`
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	OpenLibraryID string  `json:"open_library_id"` // e.g., "OL7353617M"
	ISBN          string  `json:"isbn"`            // ISBN-10 or ISBN-13 (Required via DB constraint)
	Status        string  `json:"status"`          // "Want to Read", "Currently Reading", "Read"
	Rating        *int    `json:"rating"`          // Pointer to allow null, 1-10
	Comments      *string `json:"comments"`        // Pointer to allow null
	CoverURL      *string `json:"cover_url"`       // Pointer to allow null
}

// Constants for book statuses
const (
	StatusWantToRead      = "Want to Read"
	StatusCurrentlyReading = "Currently Reading"
	StatusRead            = "Read"
)
