package model

import (
	"testing"
)

func TestBookStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status BookStatus
		want   bool
	}{
		{
			name:   "Want to Read status",
			status: StatusWantToRead,
			want:   true,
		},
		{
			name:   "Currently Reading status",
			status: StatusCurrentlyReading,
			want:   true,
		},
		{
			name:   "Read status",
			status: StatusRead,
			want:   true,
		},
		{
			name:   "Invalid status",
			status: "Invalid Status",
			want:   false,
		},
		{
			name:   "Empty status",
			status: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.IsValid(); got != tt.want {
				t.Errorf("BookStatus.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBook_Validate(t *testing.T) {
	tests := []struct {
		name    string
		book    Book
		wantErr bool
		errMsg  string
	}{
		{
			name: "Valid book with no rating",
			book: Book{
				Title:  "Test Book",
				Author: "Test Author",
				Status: StatusRead,
			},
			wantErr: false,
		},
		{
			name: "Valid book with valid rating",
			book: Book{
				Title:  "Test Book",
				Author: "Test Author",
				Status: StatusRead,
				Rating: intPtr(8),
			},
			wantErr: false,
		},
		{
			name: "Book with rating below minimum",
			book: Book{
				Title:  "Test Book",
				Author: "Test Author",
				Status: StatusRead,
				Rating: intPtr(0),
			},
			wantErr: true,
			errMsg:  "rating must be between 1 and 10",
		},
		{
			name: "Book with rating above maximum",
			book: Book{
				Title:  "Test Book",
				Author: "Test Author",
				Status: StatusRead,
				Rating: intPtr(11),
			},
			wantErr: true,
			errMsg:  "rating must be between 1 and 10",
		},
		{
			name: "Book with invalid status",
			book: Book{
				Title:  "Test Book",
				Author: "Test Author",
				Status: "Invalid Status",
			},
			wantErr: true,
			errMsg:  "invalid status provided",
		},
		{
			name: "Book with empty status",
			book: Book{
				Title:  "Test Book",
				Author: "Test Author",
				Status: "",
			},
			wantErr: true,
			errMsg:  "invalid status provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.book.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Book.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("Book.Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "Simple message",
			message: "test error",
			want:    "test error",
		},
		{
			name:    "Empty message",
			message: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ValidationError{
				Message: tt.message,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("ValidationError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to get pointer to int
func intPtr(i int) *int {
	return &i
}