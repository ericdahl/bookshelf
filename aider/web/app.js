document.addEventListener('DOMContentLoaded', () => {
    loadBooks();

    // Add event listener for the form (will be implemented later)
    const addBookForm = document.getElementById('add-book-form');
    if (addBookForm) {
        addBookForm.addEventListener('submit', handleAddBook);
    } else {
        console.error("Add book form not found");
    }
});

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

    // Add placeholders for rating/comments later
    // if (book.rating) { ... }
    // if (book.comments) { ... }

    // Add drag start listener (will be implemented later)
    // Add drag start listener
    div.addEventListener('dragstart', handleDragStart);

    return div;
}

async function handleAddBook(event) {
    event.preventDefault(); // Prevent default form submission
    console.log("Add book form submitted (implementation pending).");

    const titleInput = document.getElementById('title');
    const authorInput = document.getElementById('author');

    const newBook = {
        title: titleInput.value,
        author: authorInput.value,
        status: 'Want to Read', // Default status
        // Add other fields like open_library_id later
    };

    console.log("New book data:", newBook);

    // Implement POST request to /api/books
    try {
        const response = await fetch('/api/books', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(newBook),
        });
        if (!response.ok) {
             const errorText = await response.text(); // Get error details from backend
             throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
        }
        // Clear form and reload books
        titleInput.value = '';
        authorInput.value = '';
        loadBooks(); // Reload to show the new book
    } catch (error) {
        console.error('Error adding book:', error);
        alert(`Error adding book: ${error.message}`); // Show error to user
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
