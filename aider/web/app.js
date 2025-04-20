// --- Debounce Helper ---
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// --- Global Variables / State ---
let selectedBook = null; // Store details of the selected book from search

document.addEventListener('DOMContentLoaded', () => {
    loadBooks();
    setupAddBookForm();
});

function setupAddBookForm() {
    const addBookForm = document.getElementById('add-book-form');
    const searchInput = document.getElementById('search-book');
    const searchResultsDiv = document.getElementById('search-results');
    const addBookButton = document.getElementById('add-book-button');

    if (!addBookForm || !searchInput || !searchResultsDiv || !addBookButton) {
        console.error("Add book form elements not found");
        return;
    }

    addBookForm.addEventListener('submit', handleAddBook);

    // Debounced search function
    const debouncedSearch = debounce(async () => {
        const query = searchInput.value.trim();
        if (query.length < 3) { // Only search if query is reasonably long
            searchResultsDiv.innerHTML = ''; // Clear results
            return;
        }
        await fetchSearchResults(query);
    }, 300); // 300ms delay

    searchInput.addEventListener('input', debouncedSearch);

    // Clear results if input is cleared
    searchInput.addEventListener('input', () => {
        if (searchInput.value.trim() === '') {
            searchResultsDiv.innerHTML = '';
            clearSelection();
        }
    });
}


async function fetchSearchResults(query) {
    console.log("Searching Open Library for:", query);
    const searchResultsDiv = document.getElementById('search-results');
    searchResultsDiv.innerHTML = '<i>Searching...</i>'; // Provide feedback

    try {
        const response = await fetch(`/api/search?q=${encodeURIComponent(query)}`);
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const results = await response.json();
        displaySearchResults(results);
    } catch (error) {
        console.error('Error fetching search results:', error);
        searchResultsDiv.innerHTML = '<i style="color: red;">Error fetching results.</i>';
    }
}

function displaySearchResults(results) {
    const searchResultsDiv = document.getElementById('search-results');
    searchResultsDiv.innerHTML = ''; // Clear previous results or 'Searching...' message

    if (!results || results.length === 0) {
        searchResultsDiv.innerHTML = '<i>No results found with ISBN.</i>';
        return;
    }

    const ul = document.createElement('ul');
    ul.classList.add('search-results-list'); // Add class for styling

    results.forEach(book => {
        const li = document.createElement('li');
        li.textContent = `${book.title} by ${book.author || 'Unknown Author'} (ISBN: ${book.isbn})`;
        li.dataset.olid = book.open_library_id;
        li.dataset.isbn = book.isbn;
        li.dataset.title = book.title;
        li.dataset.author = book.author || '';
        li.dataset.coverUrl = book.cover_url || '';
        li.addEventListener('click', handleResultSelection);
        ul.appendChild(li);
    });
    searchResultsDiv.appendChild(ul);
}

function handleResultSelection(event) {
    const selectedLi = event.target;
    selectedBook = {
        open_library_id: selectedLi.dataset.olid,
        isbn: selectedLi.dataset.isbn,
        title: selectedLi.dataset.title,
        author: selectedLi.dataset.author,
        cover_url: selectedLi.dataset.coverUrl,
    };

    console.log("Selected book:", selectedBook);

    // Populate hidden fields
    document.getElementById('selected-olid').value = selectedBook.open_library_id;
    document.getElementById('selected-isbn').value = selectedBook.isbn;
    document.getElementById('selected-title').value = selectedBook.title;
    document.getElementById('selected-author').value = selectedBook.author;
    document.getElementById('selected-cover-url').value = selectedBook.cover_url;


    // Update display area
    const displayDiv = document.getElementById('selected-book-display');
    displayDiv.textContent = `Selected: ${selectedBook.title} by ${selectedBook.author}`;

    // Enable Add button
    document.getElementById('add-book-button').disabled = false;

    // Clear search input and results
    document.getElementById('search-book').value = '';
    document.getElementById('search-results').innerHTML = '';
}

function clearSelection() {
    selectedBook = null;
    document.getElementById('selected-olid').value = '';
    document.getElementById('selected-isbn').value = '';
    document.getElementById('selected-title').value = '';
    document.getElementById('selected-author').value = '';
    document.getElementById('selected-cover-url').value = '';
    document.getElementById('selected-book-display').textContent = '';
    document.getElementById('add-book-button').disabled = true;
}


async function loadBooks() {
    console.log("Loading books...");
    try {
        const response = await fetch('/api/books');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        const books = await response.json();
        console.log("Books received:", books);
        displayBooks(books);
    } catch (error) {
        console.error('Error loading books:', error);
        // Display error message to the user?
    }
}

function displayBooks(books) {
    // Get column elements
    const wantToReadCol = document.getElementById('want-to-read');
    const currentlyReadingCol = document.getElementById('currently-reading');
    const readCol = document.getElementById('read');

    // Clear existing books (simple approach)
    wantToReadCol.innerHTML = '<h2>Want to Read</h2>';
    currentlyReadingCol.innerHTML = '<h2>Currently Reading</h2>';
    readCol.innerHTML = '<h2>Read</h2>';

    if (!books || books.length === 0) {
        console.log("No books to display.");
        // Optionally display a message in each column
        return;
    }

    books.forEach(book => {
        const bookElement = createBookElement(book);
        switch (book.status) {
            case 'Want to Read':
                wantToReadCol.appendChild(bookElement);
                break;
            case 'Currently Reading':
                currentlyReadingCol.appendChild(bookElement);
                break;
            case 'Read':
                readCol.appendChild(bookElement);
                break;
            default:
                console.warn(`Unknown status for book ${book.id}: ${book.status}`);
        }
    });

    // Initialize drag and drop after books are displayed
    initDragAndDrop();
}

function createBookElement(book) {
    const div = document.createElement('div');
    div.classList.add('book-card');
    div.setAttribute('draggable', 'true'); // Make it draggable
    div.dataset.bookId = book.id; // Store book ID for later use

    const title = document.createElement('h3');
    title.textContent = book.title;
    div.appendChild(title);

    if (book.author) {
        const author = document.createElement('p');
        author.textContent = `By: ${book.author}`;
        div.appendChild(author);
    }

    if (book.isbn) {
        const isbn = document.createElement('p');
        isbn.textContent = `ISBN: ${book.isbn}`;
        isbn.style.fontSize = '0.8em'; // Smaller text for ISBN
        div.appendChild(isbn);
    }

    if (book.cover_url) {
        const cover = document.createElement('img');
        cover.src = book.cover_url;
        cover.alt = `Cover of ${book.title}`;
        cover.style.maxWidth = '50px'; // Simple styling
        cover.style.float = 'right';
        div.insertBefore(cover, title); // Insert cover before title
    }


    // Add placeholders for rating/comments later
    // if (book.rating) { ... }
    // if (book.comments) { ... }

    // Add drag start listener
    div.addEventListener('dragstart', handleDragStart);

    return div;
}

async function handleAddBook(event) {
    event.preventDefault(); // Prevent default form submission

    if (!selectedBook || !selectedBook.open_library_id || !selectedBook.isbn || !selectedBook.title) {
        alert("Please search and select a book first.");
        return;
    }

    console.log("Adding selected book:", selectedBook);

    // Construct the book object to send to the backend
    const bookToAdd = {
        open_library_id: selectedBook.open_library_id,
        isbn: selectedBook.isbn,
        title: selectedBook.title,
        author: selectedBook.author,
        cover_url: selectedBook.cover_url,
        status: models.StatusWantToRead, // Default status
        // Add rating and comments here later when those fields exist
    };

    // Implement POST request to /api/books
    try {
        const response = await fetch('/api/books', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(bookToAdd),
        });
        if (!response.ok) {
             const errorText = await response.text(); // Get error details from backend
             throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
        }
        // Clear selection and reload books
        clearSelection();
        loadBooks(); // Reload to show the new book
    } catch (error) {
        console.error('Error adding book:', error);
        alert(`Error adding book: ${error.message}`); // Show error to user
        // Optionally re-enable button or provide other feedback
    }
}

// --- Drag and Drop Functions ---

function initDragAndDrop() {
    console.log("Initializing drag and drop...");
    const columns = document.querySelectorAll('.column');
    columns.forEach(column => {
        // Prevent default behavior to allow drop
        column.addEventListener('dragover', handleDragOver);
        // Handle the drop event
        column.addEventListener('drop', handleDrop);
    });
    // Note: dragstart is added in createBookElement
}

function handleDragStart(event) {
    // Check if the dragged element is a book card
    if (event.target.classList.contains('book-card')) {
        const bookId = event.target.dataset.bookId;
        console.log("Drag start:", bookId);
        event.dataTransfer.setData('text/plain', bookId);
        event.dataTransfer.effectAllowed = 'move';
    } else {
        // Prevent dragging non-book elements within columns
        event.preventDefault();
    }
}


function handleDragOver(event) {
    event.preventDefault(); // Necessary to allow drop
    event.dataTransfer.dropEffect = 'move';
    // Optional: Add visual feedback (e.g., highlight drop zone)
    // event.target.closest('.column').classList.add('drag-over');
}

// Optional: Add dragleave listener to remove visual feedback
// document.querySelectorAll('.column').forEach(column => {
//     column.addEventListener('dragleave', (event) => {
//         event.target.closest('.column').classList.remove('drag-over');
//     });
// });


async function handleDrop(event) {
    event.preventDefault();
    // Optional: Remove visual feedback
    // event.target.closest('.column').classList.remove('drag-over');

    const bookId = event.dataTransfer.getData('text/plain');
    const targetColumn = event.target.closest('.column'); // Find the column element

    if (!targetColumn || !bookId) {
        console.error("Drop target is not a valid column or bookId is missing.");
        return;
    }

    // Find the dragged element
    const draggedElement = document.querySelector(`.book-card[data-book-id="${bookId}"]`);
    if (!draggedElement) {
        console.error(`Dragged element for book ID ${bookId} not found.`);
        return; // Should not happen if dragstart worked correctly
    }

    // Prevent dropping onto the same column it came from
    if (targetColumn === draggedElement.parentElement) {
        console.log("Book dropped onto the same column.");
        return;
    }


    const newStatus = getStatusFromColumnId(targetColumn.id);
    if (!newStatus) {
        console.error("Could not determine new status from column ID:", targetColumn.id);
        return;
    }

    console.log(`Dropping book ${bookId} into column ${targetColumn.id} (status: ${newStatus})`);

    // --- Optimistic UI Update ---
    // Move the element in the UI immediately for responsiveness.
    // Append the dragged element to the target column.
    // We need to ensure we append the actual book card, not just text inside it.
    targetColumn.appendChild(draggedElement);
    // --------------------------


    // Implement PUT request to /api/books/{bookId} to update status
    try {
        const response = await fetch(`/api/books/${bookId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ status: newStatus }), // Send only the status
        });
        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
        }
        console.log(`Book ${bookId} status updated successfully to ${newStatus}`);
        // UI is already updated optimistically. If the request fails, we might need to revert.
    } catch (error) {
        console.error('Error updating book status:', error);
        alert(`Error updating book status: ${error.message}`);
        // --- Revert UI on Failure ---
        // If the API call fails, move the element back to its original column
        // This requires knowing the original column, which we can get before the move
        // or simply reload all books for simplicity in this example.
        console.log("Reloading books due to update error...");
        loadBooks(); // Reload to ensure UI consistency after error
        // --------------------------
    }
}

function getStatusFromColumnId(columnId) {
    switch (columnId) {
        case 'want-to-read': return 'Want to Read';
        case 'currently-reading': return 'Currently Reading';
        case 'read': return 'Read';
        default:
            console.error("Unknown column ID:", columnId);
            return null;
    }
}

// --- Add models constants to JS for status ---
// This avoids hardcoding strings in multiple places
const models = {
    StatusWantToRead: "Want to Read",
    StatusCurrentlyReading: "Currently Reading",
    StatusRead: "Read",
};
