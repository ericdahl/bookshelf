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
// --- Debounce Helper ---
// ... (debounce function remains the same) ...

// --- Global Variables / State ---
// let selectedBook = null; // No longer needed, selection triggers immediate add

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

    // Remove form submit listener as adding happens on result click
    // addBookForm.addEventListener('submit', handleAddBook);

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
        // Store data directly on the element for handleResultSelection to retrieve
        li.dataset.olid = book.open_library_id;
        li.dataset.isbn = book.isbn;
        li.dataset.title = book.title;
        li.dataset.author = book.author || '';
        li.dataset.coverUrl = book.cover_url || '';
        // No individual click listener needed here anymore due to event delegation
        // li.addEventListener('click', handleResultSelection);
        ul.appendChild(li);
    });
    searchResultsDiv.appendChild(ul); // Add the list to the DOM
    searchResultsDiv.style.pointerEvents = 'auto'; // Ensure clicks are enabled after displaying results
}

// Use event delegation on the results container
function setupAddBookForm() {
    // ... existing setup ...
    const searchResultsDiv = document.getElementById('search-results');
    // ... existing setup ...

    // Use event delegation on the results container
    searchResultsDiv.addEventListener('click', handleResultSelection); // Add event listener to the container
}


// Renamed and modified to directly add the book
async function handleResultSelection(event) {
    // Ensure the click is on an LI element within the results list
    const selectedLi = event.target.closest('li');
    if (!selectedLi || !selectedLi.parentElement.classList.contains('search-results-list')) {
        return; // Click was not on a list item
    }

    // Disable further clicks while processing
    const searchResultsDiv = document.getElementById('search-results');
    searchResultsDiv.style.pointerEvents = 'none'; // Prevent multiple adds
    selectedLi.style.fontWeight = 'bold'; // Indicate processing

    const bookToAdd = {
        open_library_id: selectedLi.dataset.olid,
        isbn: selectedLi.dataset.isbn,
        title: selectedLi.dataset.title,
        author: selectedLi.dataset.author || '', // Ensure author is string even if empty

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

    // --- Add Rating and Comments Display/Editing ---
    const detailsDiv = document.createElement('div');
    detailsDiv.classList.add('book-details');
    detailsDiv.style.marginTop = '10px'; // Add some spacing

    // Rating Display and Input
    const ratingLabel = document.createElement('label');
    ratingLabel.textContent = 'Rating (1-10): ';
    ratingLabel.htmlFor = `rating-${book.id}`;
    const ratingInput = document.createElement('input');
    ratingInput.type = 'number';
    ratingInput.id = `rating-${book.id}`;
    ratingInput.name = 'rating';
    ratingInput.min = '1';
    ratingInput.max = '10';
    ratingInput.value = book.rating !== null ? book.rating : ''; // Handle null rating
    ratingInput.style.width = '50px'; // Small input field
    ratingLabel.appendChild(ratingInput);
    detailsDiv.appendChild(ratingLabel);
    detailsDiv.appendChild(document.createElement('br')); // Line break

    // Comments Display and Input
    const commentsLabel = document.createElement('label');
    commentsLabel.textContent = 'Comments: ';
    commentsLabel.htmlFor = `comments-${book.id}`;
    const commentsInput = document.createElement('textarea');
    commentsInput.id = `comments-${book.id}`;
    commentsInput.name = 'comments';
    commentsInput.rows = 2; // Small textarea
    commentsInput.style.width = '95%'; // Adjust width
    commentsInput.style.marginTop = '5px';
    commentsInput.value = book.comments !== null ? book.comments : ''; // Handle null comments
    detailsDiv.appendChild(commentsLabel);
    detailsDiv.appendChild(commentsInput);
    detailsDiv.appendChild(document.createElement('br')); // Line break

    // Save Button
    const saveButton = document.createElement('button');
    saveButton.textContent = 'Save Details';
    saveButton.classList.add('save-details-button');
    saveButton.style.marginTop = '5px';
    saveButton.addEventListener('click', () => handleUpdateDetails(book.id));
    detailsDiv.appendChild(saveButton);

    div.appendChild(detailsDiv);
    // ---------------------------------------------

    // Add drag start listener
    div.addEventListener('dragstart', handleDragStart);

    return div;
}

// async function handleAddBook(event) { ... } // No longer needed as form submission isn't used for adding

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

// --- Handle Updating Book Details (Rating/Comments) ---
async function handleUpdateDetails(bookId) {
    console.log(`Saving details for book ID: ${bookId}`);
    const ratingInput = document.getElementById(`rating-${bookId}`);
    const commentsInput = document.getElementById(`comments-${bookId}`);

    // Get values, treat empty string as null for backend
    const ratingValue = ratingInput.value.trim();
    const commentsValue = commentsInput.value.trim();

    let rating = null;
    if (ratingValue !== '') {
        rating = parseInt(ratingValue, 10);
        if (isNaN(rating) || rating < 1 || rating > 10) {
            alert("Rating must be a number between 1 and 10.");
            ratingInput.focus(); // Focus the invalid input
            return;
        }
    }

    const comments = commentsValue !== '' ? commentsValue : null;

    const payload = {
        rating: rating,
        comments: comments,
    };

    console.log("Payload for update:", payload);

    try {
        const response = await fetch(`/api/books/${bookId}/details`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload),
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
        }

        console.log(`Details updated successfully for book ID: ${bookId}`);
        // Optional: Provide visual feedback (e.g., briefly change button text/color)
        const saveButton = ratingInput.closest('.book-card').querySelector('.save-details-button');
        if (saveButton) {
            const originalText = saveButton.textContent;
            saveButton.textContent = 'Saved!';
            saveButton.disabled = true;
            setTimeout(() => {
                saveButton.textContent = originalText;
                saveButton.disabled = false;
            }, 1500); // Revert after 1.5 seconds
        }
        // No need to reload all books, the UI inputs already reflect the change.
        // If we displayed rating/comments separately from inputs, we'd update those here.

    } catch (error) {
        console.error('Error updating book details:', error);
        alert(`Error updating details: ${error.message}`);
    }
}


// --- Add models constants to JS for status ---
// This avoids hardcoding strings in multiple places
const models = {
    StatusWantToRead: "Want to Read",
    StatusCurrentlyReading: "Currently Reading",
    StatusRead: "Read",
};
