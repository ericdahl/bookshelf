document.addEventListener('DOMContentLoaded', () => {
    const bookshelfColumns = document.getElementById('bookshelf-columns');
    const searchForm = document.getElementById('search-form');
    const searchResultsContainer = document.getElementById('search-results');
    const searchQueryInput = document.getElementById('search-query');
    const searchStatus = document.getElementById('search-status');
    const bookshelfStatus = document.getElementById('bookshelf-status');
    const editModal = document.getElementById('edit-modal');
    const editForm = document.getElementById('edit-form');
    const editStatus = document.getElementById('edit-status');

    const API_BASE_URL = '/api'; // Assuming API is served from the same origin

    // --- Utility Functions ---
    function setStatusMessage(element, message, isError = false) {
        if (!element) return;
        element.textContent = message;
        element.className = isError ? 'error' : '';
        // Optionally clear message after a delay
        // setTimeout(() => { element.textContent = ''; element.className = ''; }, 5000);
    }

    // --- API Fetch Functions ---
    async function fetchBooks() {
        setStatusMessage(bookshelfStatus, 'Loading books...');
        try {
            const response = await fetch(`${API_BASE_URL}/books`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const books = await response.json();
            renderBookshelf(books);
            setStatusMessage(bookshelfStatus, 'Books loaded.');
        } catch (error) {
            console.error('Error fetching books:', error);
            setStatusMessage(bookshelfStatus, `Error loading books: ${error.message}`, true);
        }
    }

    async function searchOpenLibrary(query) {
        setStatusMessage(searchStatus, 'Searching...');
        searchResultsContainer.innerHTML = ''; // Clear previous results
        try {
            const response = await fetch(`${API_BASE_URL}/search?q=${encodeURIComponent(query)}`);
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: 'Unknown error occurred' }));
                throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
            }
            const results = await response.json();
            renderSearchResults(results);
            setStatusMessage(searchStatus, results.length > 0 ? `${results.length} results found.` : 'No results found.');
        } catch (error) {
            console.error('Error searching Open Library:', error);
            setStatusMessage(searchStatus, `Search failed: ${error.message}`, true);
        }
    }

    async function addBookToShelf(bookData) {
        setStatusMessage(searchStatus, `Adding "${bookData.title}"...`);
        try {
            const response = await fetch(`${API_BASE_URL}/books`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(bookData),
            });
            if (!response.ok) {
                 const errorData = await response.json().catch(() => ({ error: 'Failed to add book' }));
                 throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
            }
            const newBook = await response.json();
            addBookToDOM(newBook); // Add to the correct column
            setStatusMessage(searchStatus, `"${newBook.title}" added successfully!`);
            searchResultsContainer.innerHTML = ''; // Clear search results after adding
            searchQueryInput.value = ''; // Clear search input
        } catch (error) {
            console.error('Error adding book:', error);
            setStatusMessage(searchStatus, `Error adding book: ${error.message}`, true);
        }
    }

    async function updateBookStatus(bookId, newStatus) {
        setStatusMessage(bookshelfStatus, 'Updating status...');
        try {
            const response = await fetch(`${API_BASE_URL}/books/${bookId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ status: newStatus }),
            });
            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: 'Failed to update status' }));
                throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
            }
            // No need to re-fetch all books, the drop was successful visually.
            // Optionally update the book card's data attribute if needed elsewhere.
            const bookCard = document.querySelector(`.book-card[data-id="${bookId}"]`);
            if (bookCard) {
                bookCard.dataset.status = newStatus;
            }
            setStatusMessage(bookshelfStatus, 'Book status updated.');
        } catch (error) {
            console.error('Error updating book status:', error);
            setStatusMessage(bookshelfStatus, `Error updating status: ${error.message}`, true);
            // TODO: Optionally revert the drag/drop visually on error
            fetchBooks(); // Re-fetch to ensure consistency on error
        }
    }

     async function updateBookDetails(bookId, rating, comments) {
        setStatusMessage(editStatus, 'Saving changes...');
        try {
            // Prepare payload: only include fields that have values.
            // Send null if a field should be cleared.
            const payload = {};
            if (rating !== undefined) { // Check if rating was provided (could be empty string from form)
                payload.rating = rating === '' ? null : parseInt(rating, 10);
            }
             if (comments !== undefined) { // Check if comments were provided
                payload.comments = comments === '' ? null : comments;
            }

            // Basic validation before sending
            if (payload.rating !== null && (isNaN(payload.rating) || payload.rating < 1 || payload.rating > 10)) {
                 throw new Error("Rating must be a number between 1 and 10, or empty.");
            }


            const response = await fetch(`${API_BASE_URL}/books/${bookId}/details`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ error: 'Failed to save details' }));
                throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
            }

            setStatusMessage(editStatus, 'Changes saved successfully!');
            // Update the specific book card in the DOM
            fetchBooks(); // Easiest way to reflect changes for now
            closeModal(editModal); // Close modal on success

        } catch (error) {
            console.error('Error updating book details:', error);
            setStatusMessage(editStatus, `Error saving: ${error.message}`, true);
        }
    }


    // --- DOM Manipulation ---
    function createBookCard(book) {
        const card = document.createElement('article');
        card.className = 'book-card';
        card.dataset.id = book.id;
        card.dataset.status = book.status;
        card.dataset.title = book.title; // Store for modal
        card.dataset.rating = book.rating ?? ''; // Store for modal
        card.dataset.comments = book.comments ?? ''; // Store for modal

        // Cover Image
        const coverDiv = document.createElement('div');
        coverDiv.className = 'book-cover';
        if (book.cover_url) {
            const img = document.createElement('img');
            img.src = book.cover_url;
            img.alt = `Cover of ${book.title}`;
            img.onerror = () => { // Handle broken image links
                 coverDiv.innerHTML = `<div class="placeholder">No Cover</div>`;
            }
            coverDiv.appendChild(img);
        } else {
             coverDiv.innerHTML = `<div class="placeholder">No Cover</div>`;
        }

        // Book Details
        const detailsDiv = document.createElement('div');
        detailsDiv.className = 'book-details';
        detailsDiv.innerHTML = `
            <h4>${book.title}</h4>
            <p>by ${book.author || 'Unknown Author'}</p>
            ${book.isbn ? `<p>ISBN: ${book.isbn}</p>` : ''}
            ${book.rating ? `<p class="rating">Rating: ${'★'.repeat(book.rating)}${'☆'.repeat(10 - book.rating)} (${book.rating}/10)</p>` : ''}
            ${book.comments ? `<p class="comments">Comments: ${book.comments}</p>` : ''}
            <div class="book-actions">
                <button class="edit-button secondary outline" data-id="${book.id}">Edit</button>
                <!-- <button class="delete-button contrast outline" data-id="${book.id}">Delete</button> -->
            </div>
        `;

        card.appendChild(coverDiv);
        card.appendChild(detailsDiv);

        // Add event listener for the edit button
        const editButton = card.querySelector('.edit-button');
        if (editButton) {
            editButton.addEventListener('click', (e) => {
                e.stopPropagation(); // Prevent card drag initiation
                openEditModal(book);
            });
        }

        // Add event listener for delete button (future)
        // const deleteButton = card.querySelector('.delete-button');
        // if (deleteButton) { ... }

        return card;
    }

    function addBookToDOM(book) {
        const listIdMap = {
            "Want to Read": "want-to-read-list",
            "Currently Reading": "currently-reading-list",
            "Read": "read-list"
        };
        const listId = listIdMap[book.status];
        if (listId) {
            const listElement = document.getElementById(listId);
            const bookCard = createBookCard(book);
            listElement.appendChild(bookCard);
        } else {
            console.warn(`Unknown status "${book.status}" for book ID ${book.id}`);
        }
    }

    function renderBookshelf(books) {
        // Clear existing books
        document.querySelectorAll('.book-list').forEach(list => list.innerHTML = '');
        // Populate lists
        books.forEach(book => addBookToDOM(book));
        // Re-initialize sortable after rendering
        initializeSortable();
    }

    function renderSearchResults(results) {
        searchResultsContainer.innerHTML = ''; // Clear previous
        if (!results || results.length === 0) {
            searchResultsContainer.innerHTML = '<p>No books found matching your query.</p>';
            return;
        }

        results.forEach(book => {
            const item = document.createElement('div');
            item.className = 'search-result-item';

            // Basic info display
            item.innerHTML = `
                ${book.cover_url ? `<img src="${book.cover_url}" alt="Cover">` : '<div class="placeholder" style="width:40px; height:60px; font-size:0.7em;">No Cover</div>'}
                <div class="details">
                    <strong>${book.title}</strong>
                    <span>by ${book.author || 'Unknown Author'}</span>
                    ${book.isbn ? `<span> | ISBN: ${book.isbn}</span>` : ''}
                </div>
                <button class="add-button" data-olid="${book.open_library_id}">Add to Shelf</button>
            `;

            // Add event listener to the button
            const addButton = item.querySelector('.add-button');
            addButton.addEventListener('click', () => {
                // Prepare data to send to POST /api/books
                const bookData = {
                    title: book.title,
                    author: book.author || 'Unknown Author',
                    open_library_id: book.open_library_id,
                    isbn: book.isbn || null, // Send null if undefined/empty
                    cover_url: book.cover_url || null,
                    // Status will be defaulted by backend if not sent, or we can set it here:
                    // status: "Want to Read"
                };
                addBookToShelf(bookData);
            });

            searchResultsContainer.appendChild(item);
        });
    }

    // --- Drag and Drop (SortableJS) ---
    function initializeSortable() {
        const lists = document.querySelectorAll('.book-list');
        lists.forEach(list => {
            new Sortable(list, {
                group: 'bookshelf', // Set same group name for all lists
                animation: 150,
                ghostClass: 'sortable-ghost', // Class name for the drop placeholder
                chosenClass: 'sortable-chosen', // Class name for the chosen item
                onEnd: function (evt) {
                    // evt.to   -> Target list element
                    // evt.from -> Source list element
                    // evt.item -> Dragged element (the book card)
                    // evt.newIndex -> Index in target list
                    // evt.oldIndex -> Index in source list

                    const bookId = evt.item.dataset.id;
                    const newStatus = evt.to.dataset.status; // Get status from target list's data attribute

                    if (bookId && newStatus && evt.from !== evt.to) {
                        console.log(`Book ID ${bookId} moved to ${newStatus}`);
                        updateBookStatus(bookId, newStatus);
                    }
                },
            });
        });
    }

    // --- Modal Handling (Pico.css style) ---
    const openModal = (modal) => {
        if (modal) {
            modal.showModal(); // Use native dialog method
            modal.setAttribute('aria-hidden', 'false');
            document.body.style.overflow = 'hidden'; // Prevent background scrolling
        }
    };

    const closeModal = (modal) => {
        if (modal) {
            modal.close(); // Use native dialog method
            modal.setAttribute('aria-hidden', 'true');
            document.body.style.overflow = ''; // Restore background scrolling
        }
    };

    // Global function for Pico's data-target dismissal (can be called from HTML)
    window.toggleModal = (event) => {
        event.preventDefault();
        const modal = document.getElementById(event.currentTarget.dataset.target);
        if (modal) {
            if (modal.hasAttribute('open')) {
                 closeModal(modal);
            } else {
                 openModal(modal);
            }
        }
    };

     // Close modal on clicking the backdrop (for native <dialog>)
    editModal.addEventListener('click', (event) => {
        if (event.target === editModal) {
            closeModal(editModal);
        }
    });


    function openEditModal(book) {
        const modalTitle = document.getElementById('modal-book-title');
        const bookIdInput = document.getElementById('edit-book-id');
        const ratingInput = document.getElementById('edit-rating');
        const commentsInput = document.getElementById('edit-comments');

        modalTitle.textContent = book.title;
        bookIdInput.value = book.id;
        // Use ?? '' to handle null/undefined -> empty string for form fields
        ratingInput.value = book.rating ?? '';
        commentsInput.value = book.comments ?? '';
        setStatusMessage(editStatus, ''); // Clear previous status

        openModal(editModal);
    }


    // --- Event Listeners ---
    searchForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const query = searchQueryInput.value.trim();
        if (query) {
            searchOpenLibrary(query);
        } else {
             setStatusMessage(searchStatus, 'Please enter a search term.', true);
        }
    });

     editForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const bookId = document.getElementById('edit-book-id').value;
        const rating = document.getElementById('edit-rating').value; // Get value as string
        const comments = document.getElementById('edit-comments').value; // Get value as string

        if (bookId) {
            updateBookDetails(bookId, rating, comments);
        } else {
             setStatusMessage(editStatus, 'Error: Book ID missing.', true);
        }
    });


    // --- Initial Load ---
    fetchBooks();
    // Note: SortableJS initialization is called within renderBookshelf after books are loaded.

});
