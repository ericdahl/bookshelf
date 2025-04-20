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

    // Initialize drag and drop after books are displayed (will be added later)
    // initDragAndDrop();
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
    // div.addEventListener('dragstart', handleDragStart);

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

    // TODO: Implement POST request to /api/books
    // try {
    //     const response = await fetch('/api/books', {
    //         method: 'POST',
    //         headers: {
    //             'Content-Type': 'application/json',
    //         },
    //         body: JSON.stringify(newBook),
    //     });
    //     if (!response.ok) {
    //         throw new Error(`HTTP error! status: ${response.status}`);
    //     }
    //     // Clear form and reload books
    //     titleInput.value = '';
    //     authorInput.value = '';
    //     loadBooks();
    // } catch (error) {
    //     console.error('Error adding book:', error);
    // }
}

// --- Drag and Drop Functions (Placeholders) ---

// function initDragAndDrop() {
//     console.log("Initializing drag and drop (implementation pending)...");
//     const columns = document.querySelectorAll('.column');
//     columns.forEach(column => {
//         column.addEventListener('dragover', handleDragOver);
//         column.addEventListener('drop', handleDrop);
//     });
// }

// function handleDragStart(event) {
//     console.log("Drag start:", event.target.dataset.bookId);
//     event.dataTransfer.setData('text/plain', event.target.dataset.bookId);
//     event.dataTransfer.effectAllowed = 'move';
// }

// function handleDragOver(event) {
//     event.preventDefault(); // Necessary to allow drop
//     event.dataTransfer.dropEffect = 'move';
// }

// async function handleDrop(event) {
//     event.preventDefault();
//     const bookId = event.dataTransfer.getData('text/plain');
//     const targetColumn = event.target.closest('.column'); // Find the column element

//     if (!targetColumn || !bookId) {
//         console.error("Drop target is not a valid column or bookId is missing.");
//         return;
//     }

//     const newStatus = getStatusFromColumnId(targetColumn.id);
//     if (!newStatus) {
//         console.error("Could not determine new status from column ID:", targetColumn.id);
//         return;
//     }

//     console.log(`Dropping book ${bookId} into column ${targetColumn.id} (status: ${newStatus})`);

//     // TODO: Implement PUT request to /api/books/{bookId} to update status
//     // try {
//     //     const response = await fetch(`/api/books/${bookId}`, {
//     //         method: 'PUT',
//     //         headers: {
//     //             'Content-Type': 'application/json',
//     //         },
//     //         body: JSON.stringify({ status: newStatus }),
//     //     });
//     //     if (!response.ok) {
//     //         throw new Error(`HTTP error! status: ${response.status}`);
//     //     }
//     //     // Move the element in the UI immediately for responsiveness
//     //     const draggedElement = document.querySelector(`.book-card[data-book-id="${bookId}"]`);
//     //     if (draggedElement) {
//     //         targetColumn.appendChild(draggedElement);
//     //     } else {
//     //         // If element not found (shouldn't happen), reload all books
//     //         loadBooks();
//     //     }
//     // } catch (error) {
//     //     console.error('Error updating book status:', error);
//     //     // Optionally revert the UI change or show an error
//     // }
// }

// function getStatusFromColumnId(columnId) {
//     switch (columnId) {
//         case 'want-to-read': return 'Want to Read';
//         case 'currently-reading': return 'Currently Reading';
//         case 'read': return 'Read';
//         default: return null;
//     }
// }
