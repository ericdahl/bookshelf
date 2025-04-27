document.addEventListener('DOMContentLoaded', function() {
    // API endpoints
    const API = {
        BOOKS: '/api/books',
        SEARCH: '/api/books/search',
        BOOK_STATUS: (id) => `/api/books/${id}`,
        BOOK_DETAILS: (id) => `/api/books/${id}/details`,
        DELETE_BOOK: (id) => `/api/books/${id}`
    };

    // DOM Elements
    const searchInput = document.getElementById('search-input');
    const searchButton = document.getElementById('search-button');
    const searchResults = document.getElementById('search-results');
    const closeSearch = document.getElementById('close-search');
    const resultsContainer = document.querySelector('.results-container');
    const shelves = document.querySelectorAll('.books-container');
    const bookDetails = document.getElementById('book-details');
    const closeDetails = document.getElementById('close-details');
    const saveDetails = document.getElementById('save-details');
    const deleteBookButton = document.getElementById('delete-book');
    const loadingOverlay = document.getElementById('loading-overlay');
    const ratingStars = document.querySelectorAll('.stars i');

    // Current book being viewed/edited
    let currentBook = null;
    let currentRating = null;

    // Initialize the application
    initApp();

    // Initialize the application
    function initApp() {
        // Load all books from the server
        loadBooks();

        // Set up event listeners
        setupEventListeners();

        // Initialize drag and drop
        initDragAndDrop();
    }

    // Load all books from the server
    function loadBooks() {
        showLoading();
        fetch(API.BOOKS)
            .then(response => response.json())
            .then(books => {
                // Clear existing books from shelves
                document.querySelectorAll('.books-container').forEach(shelf => {
                    shelf.innerHTML = '';
                });
                
                // Add books to their respective shelves
                books.forEach(book => {
                    addBookToShelf(book);
                });
                hideLoading();
            })
            .catch(error => {
                console.error('Error loading books:', error);
                hideLoading();
                alert('Failed to load books. Please try again.');
            });
    }

    // Set up event listeners
    function setupEventListeners() {
        // Search
        searchButton.addEventListener('click', searchBooks);
        searchInput.addEventListener('keypress', e => {
            if (e.key === 'Enter') {
                searchBooks();
            }
        });
        
        // Close search results
        closeSearch.addEventListener('click', () => {
            searchResults.classList.add('hidden');
        });

        // Book details
        closeDetails.addEventListener('click', () => {
            bookDetails.classList.add('hidden');
        });
        
        // Rating stars
        ratingStars.forEach(star => {
            star.addEventListener('click', function() {
                const rating = parseInt(this.dataset.rating);
                updateRatingUI(rating);
                currentRating = rating;
            });
            
            // Add hover effect
            star.addEventListener('mouseenter', function() {
                const rating = parseInt(this.dataset.rating);
                previewRating(rating);
            });
        });
        
        // Reset rating preview on mouseleave
        document.querySelector('.stars').addEventListener('mouseleave', function() {
            updateRatingUI(currentRating || 0);
        });
        
        // Save book details
        saveDetails.addEventListener('click', saveBookDetails);
        
        // Delete book
        deleteBookButton.addEventListener('click', deleteBook);
    }

    // Initialize drag and drop
    function initDragAndDrop() {
        shelves.forEach(shelf => {
            new Sortable(shelf, {
                group: 'books',
                animation: 150,
                ghostClass: 'sortable-ghost',
                dragClass: 'sortable-drag',
                onEnd: function(evt) {
                    const bookId = evt.item.dataset.id;
                    const newStatus = evt.to.dataset.status;
                    
                    // Update the book status on the server
                    updateBookStatus(bookId, newStatus);
                }
            });
        });
    }

    // Search for books using the Open Library API
    function searchBooks() {
        const query = searchInput.value.trim();
        
        if (query === '') {
            return;
        }
        
        showLoading();
        fetch(`${API.SEARCH}?q=${encodeURIComponent(query)}`)
            .then(response => response.json())
            .then(books => {
                resultsContainer.innerHTML = '';
                
                if (books.length === 0) {
                    resultsContainer.innerHTML = '<p>No books found. Try a different search term.</p>';
                } else {
                    books.forEach(book => {
                        resultsContainer.appendChild(createSearchResultCard(book));
                    });
                }
                
                searchResults.classList.remove('hidden');
                hideLoading();
            })
            .catch(error => {
                console.error('Error searching books:', error);
                hideLoading();
                alert('Failed to search books. Please try again.');
            });
    }

    // Add a book to the appropriate shelf
    function addBookToShelf(book) {
        const shelf = document.querySelector(`.books-container[data-status="${book.status}"]`);
        if (shelf) {
            shelf.appendChild(createBookCard(book));
        }
    }

    // Create a book card element
    function createBookCard(book) {
        const card = document.createElement('div');
        card.className = 'book-card';
        card.dataset.id = book.id;
        
        const coverUrl = book.cover_url || 'https://via.placeholder.com/150x200?text=No+Cover';
        const ratingHtml = book.rating ? `<p class="book-rating">Rating: ${book.rating}/10</p>` : '';
        
        card.innerHTML = `
            <div class="book-cover">
                <img src="${coverUrl}" alt="${book.title} cover">
            </div>
            <div class="book-info">
                <h3 class="book-title">${book.title}</h3>
                <p class="book-author">${book.author}</p>
                ${ratingHtml}
            </div>
        `;
        
        // Add click event to open book details
        card.addEventListener('click', () => {
            showBookDetails(book);
        });
        
        return card;
    }

    // Create a search result card
    function createSearchResultCard(book) {
        const card = document.createElement('div');
        card.className = 'book-card search-result';
        
        const coverUrl = book.cover_url || 'https://via.placeholder.com/150x200?text=No+Cover';
        
        card.innerHTML = `
            <div class="book-cover">
                <img src="${coverUrl}" alt="${book.title} cover">
            </div>
            <div class="book-info">
                <h3 class="book-title">${book.title}</h3>
                <p class="book-author">${book.author}</p>
                <button class="add-book">Add to Shelf</button>
            </div>
        `;
        
        // Add event listener to the add button
        const addButton = card.querySelector('.add-book');
        addButton.addEventListener('click', () => {
            addBook(book);
        });
        
        return card;
    }

    // Add a new book to the shelf
    function addBook(book) {
        showLoading();
        
        const newBook = {
            title: book.title,
            author: book.author,
            open_library_id: book.open_library_id,
            isbn: book.isbn || '',
            status: 'Want to Read',
            cover_url: book.cover_url || null
        };
        
        fetch(API.BOOKS, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(newBook)
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to add book');
            }
            return response.json();
        })
        .then(addedBook => {
            // Add the book to the shelf
            addBookToShelf(addedBook);
            // Close the search results
            searchResults.classList.add('hidden');
            hideLoading();
        })
        .catch(error => {
            console.error('Error adding book:', error);
            hideLoading();
            alert('Failed to add book. Please try again.');
        });
    }

    // Update a book's status
    function updateBookStatus(bookId, newStatus) {
        showLoading();
        
        fetch(API.BOOK_STATUS(bookId), {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ status: newStatus })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to update book status');
            }
            hideLoading();
        })
        .catch(error => {
            console.error('Error updating book status:', error);
            hideLoading();
            alert('Failed to update book status. Please try again.');
            // Reload books to reset the UI to the server state
            loadBooks();
        });
    }

    // Show book details
    function showBookDetails(book) {
        currentBook = book;
        
        // Update the UI with book details
        document.getElementById('detail-title').textContent = book.title;
        document.getElementById('detail-author').textContent = book.author;
        document.getElementById('detail-cover').src = book.cover_url || 'https://via.placeholder.com/150x200?text=No+Cover';
        
        // Update rating UI
        updateRatingUI(book.rating || 0);
        currentRating = book.rating || null;
        
        // Update comments
        document.getElementById('book-comments').value = book.comments || '';
        
        // Show the details popup
        bookDetails.classList.remove('hidden');
    }

    // Preview rating on hover
    function previewRating(rating) {
        ratingStars.forEach((star, index) => {
            if (index < rating) {
                star.className = 'fas fa-star';
            } else {
                star.className = 'far fa-star';
            }
        });
        
        // Update the rating value text
        const ratingText = rating > 0 ? rating.toString() : "None";
        document.getElementById('rating-value').textContent = ratingText;
    }
    
    // Update rating UI
    function updateRatingUI(rating) {
        ratingStars.forEach((star, index) => {
            if (index < rating) {
                star.className = 'fas fa-star';
            } else {
                star.className = 'far fa-star';
            }
        });
        
        // Update the rating value text
        const ratingText = rating > 0 ? rating.toString() : "None";
        document.getElementById('rating-value').textContent = ratingText;
    }

    // Save book details
    function saveBookDetails() {
        if (!currentBook) return;
        
        showLoading();
        
        const comments = document.getElementById('book-comments').value.trim();
        
        fetch(API.BOOK_DETAILS(currentBook.id), {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                rating: currentRating,
                comments: comments || null
            })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to update book details');
            }
            
            // Update the current book object
            currentBook.rating = currentRating;
            currentBook.comments = comments || null;
            
            // Update the book card in the shelf
            updateBookCardInShelf(currentBook);
            
            // Close the details popup
            bookDetails.classList.add('hidden');
            hideLoading();
        })
        .catch(error => {
            console.error('Error updating book details:', error);
            hideLoading();
            alert('Failed to update book details. Please try again.');
        });
    }

    // Delete a book
    function deleteBook() {
        if (!currentBook || !confirm('Are you sure you want to delete this book?')) return;
        
        showLoading();
        
        fetch(API.DELETE_BOOK(currentBook.id), {
            method: 'DELETE'
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to delete book');
            }
            
            // Close the details popup
            bookDetails.classList.add('hidden');
            
            // Reload books to update the UI
            loadBooks();
        })
        .catch(error => {
            console.error('Error deleting book:', error);
            hideLoading();
            alert('Failed to delete book. Please try again.');
        });
    }

    // Show loading overlay
    function showLoading() {
        loadingOverlay.classList.remove('hidden');
    }

    // Hide loading overlay
    function hideLoading() {
        loadingOverlay.classList.add('hidden');
    }

    // Update book card in shelf
    function updateBookCardInShelf(book) {
        const bookCard = document.querySelector(`.book-card[data-id="${book.id}"]`);
        if (bookCard) {
            const ratingElement = bookCard.querySelector('.book-rating');
            
            if (book.rating) {
                if (ratingElement) {
                    ratingElement.textContent = `Rating: ${book.rating}/10`;
                } else {
                    const bookInfo = bookCard.querySelector('.book-info');
                    const ratingP = document.createElement('p');
                    ratingP.className = 'book-rating';
                    ratingP.textContent = `Rating: ${book.rating}/10`;
                    bookInfo.appendChild(ratingP);
                }
            } else if (ratingElement) {
                ratingElement.remove();
            }
        }
    }
});