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
    const fullViewButton = document.getElementById('full-view');
    const compactViewButton = document.getElementById('compact-view');
    const shelvesContainer = document.querySelector('.shelves-container');

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
        
        // Initialize shelf sorting
        initShelfSorting();
    }
    
    // Initialize sorting functionality for each shelf
    function initShelfSorting() {
        const sortSelects = document.querySelectorAll('.sort-select');
        
        sortSelects.forEach(select => {
            // Set default sort value
            select.value = 'title';
            
            // Add change event listener
            select.addEventListener('change', function() {
                const status = this.getAttribute('data-status');
                const sortBy = this.value;
                sortShelfBooks(status, sortBy);
            });
        });
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
                
                // Group books by status
                const booksByStatus = {
                    'Want to Read': [],
                    'Currently Reading': [],
                    'Read': []
                };
                
                books.forEach(book => {
                    if (booksByStatus[book.status]) {
                        booksByStatus[book.status].push(book);
                    }
                });
                
                // Sort each shelf's books by title (default) and add to shelf
                Object.keys(booksByStatus).forEach(status => {
                    const sortSelect = document.querySelector(`.sort-select[data-status="${status}"]`);
                    const sortBy = sortSelect ? sortSelect.value : 'title';
                    
                    const sortedBooks = sortBooks(booksByStatus[status], sortBy);
                    sortedBooks.forEach(book => {
                        addBookToShelf(book);
                    });
                });
                
                hideLoading();
            })
            .catch(error => {
                console.error('Error loading books:', error);
                hideLoading();
                alert('Failed to load books. Please try again.');
            });
    }
    
    // Sort an array of books by the given criteria
    function sortBooks(books, sortBy) {
        return [...books].sort((a, b) => {
            switch (sortBy) {
                case 'title':
                    return a.title.localeCompare(b.title);
                case 'author':
                    return a.author.localeCompare(b.author);
                case 'rating':
                    // Handle null ratings (null ratings go at the end)
                    if (a.rating === null && b.rating === null) return 0;
                    if (a.rating === null) return 1;
                    if (b.rating === null) return -1;
                    // Sort by rating in descending order (higher ratings first)
                    return b.rating - a.rating;
                default:
                    return 0;
            }
        });
    }
    
    // Sort books in a specific shelf
    function sortShelfBooks(status, sortBy) {
        // Get all books from this shelf
        const shelf = document.querySelector(`.books-container[data-status="${status}"]`);
        if (!shelf) return;
        
        // Get all book cards in this shelf
        const bookCards = Array.from(shelf.querySelectorAll('.book-card'));
        
        // Map to objects with data for sorting
        const booksData = bookCards.map(card => {
            const id = card.dataset.id;
            const title = card.querySelector('.book-title').textContent;
            const author = card.querySelector('.book-author').textContent;
            
            // Parse rating if it exists
            let rating = null;
            const ratingElem = card.querySelector('.book-rating');
            if (ratingElem) {
                const ratingMatch = ratingElem.textContent.match(/(\d+)/);
                if (ratingMatch) {
                    rating = parseInt(ratingMatch[1], 10);
                }
            }
            
            return { element: card, id, title, author, rating };
        });
        
        // Sort books
        const sortedBooks = [...booksData].sort((a, b) => {
            switch (sortBy) {
                case 'title':
                    return a.title.localeCompare(b.title);
                case 'author':
                    return a.author.localeCompare(b.author);
                case 'rating':
                    // Handle null ratings (null ratings go at the end)
                    if (a.rating === null && b.rating === null) return 0;
                    if (a.rating === null) return 1;
                    if (b.rating === null) return -1;
                    // Sort by rating in descending order (higher ratings first)
                    return b.rating - a.rating;
                default:
                    return 0;
            }
        });
        
        // Remove all books from shelf
        bookCards.forEach(card => card.remove());
        
        // Add back in sorted order
        sortedBooks.forEach(book => {
            shelf.appendChild(book.element);
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
            searchInput.value = ''; // Clear search input
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
        
        // View toggle buttons
        fullViewButton.addEventListener('click', () => {
            setViewMode('full');
        });
        
        compactViewButton.addEventListener('click', () => {
            setViewMode('compact');
        });
        
        // Load saved view preference
        loadViewPreference();
    }
    
    // Set the view mode (full or compact)
    function setViewMode(mode) {
        if (mode === 'compact') {
            shelvesContainer.classList.add('compact-mode');
            fullViewButton.classList.remove('active');
            compactViewButton.classList.add('active');
            // Save preference
            localStorage.setItem('bookshelfViewMode', 'compact');
            
            // Add table headers to each shelf
            document.querySelectorAll('.books-container').forEach(container => {
                // Remove existing headers if any
                const existingHeader = container.querySelector('.books-container-header');
                if (existingHeader) {
                    existingHeader.remove();
                }
                
                // Create new header
                const header = document.createElement('div');
                header.className = 'books-container-header';
                
                const headerRow = document.createElement('div');
                headerRow.className = 'header-row';
                
                // Add column headers
                const titleHeader = document.createElement('div');
                titleHeader.className = 'header-cell';
                titleHeader.textContent = 'Title';
                
                const authorHeader = document.createElement('div');
                authorHeader.className = 'header-cell';
                authorHeader.textContent = 'Author';
                
                const seriesHeader = document.createElement('div');
                seriesHeader.className = 'header-cell';
                seriesHeader.textContent = 'Series';
                
                const ratingHeader = document.createElement('div');
                ratingHeader.className = 'header-cell';
                ratingHeader.textContent = 'Rating';
                
                // Assemble header
                headerRow.appendChild(titleHeader);
                headerRow.appendChild(authorHeader);
                headerRow.appendChild(seriesHeader);
                headerRow.appendChild(ratingHeader);
                header.appendChild(headerRow);
                
                // Insert header at the beginning of the container
                container.insertBefore(header, container.firstChild);
                
                // Convert existing book cards to tabular format
                container.querySelectorAll('.book-card').forEach(convertBookCardToTableRow);
            });
        } else {
            shelvesContainer.classList.remove('compact-mode');
            fullViewButton.classList.add('active');
            compactViewButton.classList.remove('active');
            // Save preference
            localStorage.setItem('bookshelfViewMode', 'full');
            
            // Remove table headers
            document.querySelectorAll('.books-container-header').forEach(header => {
                header.remove();
            });
            
            // Restore original book card structure if needed
            document.querySelectorAll('.book-card').forEach(card => {
                // Make sure book-info is displayed
                const infoDiv = card.querySelector('.book-info');
                if (infoDiv) {
                    infoDiv.style.removeProperty('display');
                }
                
                // Remove any table cells if they exist
                const cells = card.querySelectorAll('.cell-title, .cell-author, .cell-series, .cell-rating');
                cells.forEach(cell => cell.remove());
            });
        }
    }
    
    // Convert a book card to a table row format
    function convertBookCardToTableRow(card) {
        // If cells already exist, just return
        if (card.querySelector('.cell-title')) {
            return;
        }
        
        // Get book data from existing elements
        const title = card.querySelector('.book-title').textContent;
        const author = card.querySelector('.book-author').textContent;
        
        // Create table cells
        const titleCell = document.createElement('div');
        titleCell.className = 'cell-title';
        titleCell.innerHTML = `<div class="book-title">${title}</div>`;
        
        const authorCell = document.createElement('div');
        authorCell.className = 'cell-author';
        authorCell.innerHTML = `<div class="book-author">${author}</div>`;
        
        const seriesCell = document.createElement('div');
        seriesCell.className = 'cell-series';
        const seriesElement = card.querySelector('.book-series');
        if (seriesElement) {
            seriesCell.innerHTML = `<div class="book-series">${seriesElement.textContent}</div>`;
        } else {
            seriesCell.innerHTML = `<div class="book-series">-</div>`;
        }
        
        const ratingCell = document.createElement('div');
        ratingCell.className = 'cell-rating';
        const ratingElement = card.querySelector('.book-rating');
        if (ratingElement) {
            ratingCell.innerHTML = `<div class="book-rating">${ratingElement.textContent.replace('Rating: ', '')}</div>`;
        } else {
            ratingCell.innerHTML = `<div class="book-rating">-</div>`;
        }
        
        // Add cells to the card
        card.appendChild(titleCell);
        card.appendChild(authorCell);
        card.appendChild(seriesCell);
        card.appendChild(ratingCell);
    }
    
    // Load saved view preference
    function loadViewPreference() {
        const savedMode = localStorage.getItem('bookshelfViewMode');
        if (savedMode === 'compact') {
            setViewMode('compact');
        } else {
            setViewMode('full'); // Default to full view
        }
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
                },
                // Disable drag and drop in compact mode by checking the view mode
                disabled: function() {
                    return shelvesContainer.classList.contains('compact-mode');
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
                    // Show number of results
                    resultsContainer.innerHTML = `<p class="search-count">${books.length} books found for "${query}"</p>`;
                    
                    // Create a container for the book cards
                    const booksGrid = document.createElement('div');
                    booksGrid.className = 'search-results-grid';
                    
                    // Add each book to the grid
                    books.forEach(book => {
                        booksGrid.appendChild(createSearchResultCard(book));
                    });
                    
                    resultsContainer.appendChild(booksGrid);
                }
                
                // Automatically scroll to the search results
                searchResults.classList.remove('hidden');
                searchResults.scrollIntoView({ behavior: 'smooth' });
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
            const bookCard = createBookCard(book);
            shelf.appendChild(bookCard);
            
            // If in compact mode, convert the card to table row format
            if (shelvesContainer.classList.contains('compact-mode')) {
                convertBookCardToTableRow(bookCard);
            }
        }
    }

    // Create a book card element
    function createBookCard(book) {
        const card = document.createElement('div');
        card.className = 'book-card';
        card.dataset.id = book.id;
        
        const coverUrl = book.cover_url || 'https://via.placeholder.com/150x200?text=No+Cover';
        const ratingHtml = book.rating ? `<p class="book-rating">Rating: ${book.rating}/10</p>` : '';
        
        // Prepare series info display if available
        let seriesHtml = '';
        if (book.series && book.series_index) {
            seriesHtml = `<p class="book-series">${book.series} Book ${book.series_index}</p>`;
        } else if (book.series) {
            seriesHtml = `<p class="book-series">${book.series}</p>`;
        }
        
        card.innerHTML = `
            <div class="book-cover">
                <img src="${coverUrl}" alt="${book.title} cover">
            </div>
            <div class="book-info">
                <h3 class="book-title">${book.title}</h3>
                <p class="book-author">${book.author}</p>
                ${seriesHtml}
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
        
        // Set button text based on whether book is already in a shelf
        let buttonText = "Add to Shelf";
        let buttonClass = "add-book";
        if (book.existing_shelf) {
            buttonText = `Shelf: ${book.existing_shelf}`;
            buttonClass = "add-book book-exists";
        }

        card.innerHTML = `
            <div class="book-cover">
                <img src="${coverUrl}" alt="${book.title} cover">
            </div>
            <div class="book-info">
                <h3 class="book-title">${book.title}</h3>
                <p class="book-author">${book.author}</p>
                <button class="${buttonClass}">${buttonText}</button>
            </div>
        `;
        
        // Add event listener to the add button (only if it's not already in a shelf)
        const addButton = card.querySelector('.add-book');
        if (!book.existing_shelf) {
            addButton.addEventListener('click', () => {
                addBook(book, addButton);
            });
        }
        
        return card;
    }

    // Add a new book to the shelf
    function addBook(book, buttonElement) {
        // If book already exists in a shelf, just show a notification
        if (book.existing_shelf) {
            alert(`This book is already in your "${book.existing_shelf}" shelf`);
            return;
        }
        
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
            
            // Update button to show it was added
            if (buttonElement) {
                buttonElement.textContent = "Added ✓";
                buttonElement.disabled = true;
                buttonElement.classList.add("book-added");
            }
            
            // Keep search results open for adding more books
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
        
        // Update OpenLibrary link
        const openLibraryLink = document.getElementById('detail-openlibrary-link').querySelector('a');
        if (book.open_library_id) {
            // Check if the ID is in the format OL12345M or if it's a full path like /works/OL12345M
            let olid = book.open_library_id;
            if (olid.startsWith('/')) {
                // Extract just the ID part
                const parts = olid.split('/');
                olid = parts[parts.length - 1];
            }
            
            // Set the URL based on the format of the ID
            let url;
            if (olid.startsWith('OL') && olid.endsWith('M')) {
                // It's an edition ID (starts with OL and ends with M)
                url = `https://openlibrary.org/books/${olid}`;
            } else if (olid.startsWith('OL') && olid.endsWith('W')) {
                // It's a works ID (starts with OL and ends with W)
                url = `https://openlibrary.org/works/${olid}`;
            } else {
                // Default to works path 
                url = `https://openlibrary.org/works/${olid}`;
            }
            
            openLibraryLink.href = url;
            openLibraryLink.parentElement.style.display = 'block';
        } else {
            openLibraryLink.parentElement.style.display = 'none';
        }
        
        // Update rating UI
        updateRatingUI(book.rating || 0);
        currentRating = book.rating || null;
        
        // Update comments
        document.getElementById('book-comments').value = book.comments || '';
        
        // Update series information
        document.getElementById('book-series').value = book.series || '';
        document.getElementById('book-series-index').value = book.series_index || '';
        
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
        const series = document.getElementById('book-series').value.trim();
        let seriesIndex = document.getElementById('book-series-index').value;
        
        // Convert seriesIndex to a number if it's not empty
        if (seriesIndex) {
            seriesIndex = parseInt(seriesIndex);
            // Validate series index
            if (isNaN(seriesIndex) || seriesIndex <= 0) {
                hideLoading();
                alert('Series index must be a positive number');
                return;
            }
        } else {
            seriesIndex = null;
        }
        
        // Validate that series index requires series name
        if (seriesIndex && !series) {
            hideLoading();
            alert('Cannot add a series index without a series name');
            return;
        }
        
        fetch(API.BOOK_DETAILS(currentBook.id), {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                rating: currentRating,
                comments: comments || null,
                series: series || null,
                series_index: seriesIndex
            })
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to update book details');
            }
            
            // Update the current book object
            currentBook.rating = currentRating;
            currentBook.comments = comments || null;
            currentBook.series = series || null;
            currentBook.series_index = seriesIndex;
            
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
            // Update rating if needed
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
            
            // Update series info if needed
            const seriesElement = bookCard.querySelector('.book-series');
            if (book.series) {
                const seriesText = book.series_index 
                    ? `${book.series} Book ${book.series_index}` 
                    : book.series;
                    
                if (seriesElement) {
                    seriesElement.textContent = seriesText;
                } else {
                    const bookInfo = bookCard.querySelector('.book-info');
                    const authorElement = bookInfo.querySelector('.book-author');
                    
                    const seriesP = document.createElement('p');
                    seriesP.className = 'book-series';
                    seriesP.textContent = seriesText;
                    
                    // Insert after author element
                    if (authorElement.nextSibling) {
                        bookInfo.insertBefore(seriesP, authorElement.nextSibling);
                    } else {
                        bookInfo.appendChild(seriesP);
                    }
                }
            } else if (seriesElement) {
                seriesElement.remove();
            }
        }
    }
});