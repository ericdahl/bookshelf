const { createApp, ref, reactive, computed, onMounted, nextTick } = Vue;

const API_BASE_URL = '/api';

// Simple reusable Book Card component
const BookCard = {
    props: ['book'],
    emits: ['edit'],
    template: `
        <div class="book-cover">
            <img v-if="book.cover_url" :src="book.cover_url" :alt="'Cover of ' + book.title" @error="onImgError">
            <div v-else class="placeholder">No Cover</div>
        </div>
        <div class="book-details">
            <h4>{{ book.title }}</h4>
            <p>by {{ book.author || 'Unknown Author' }}</p>
            <p v-if="book.isbn">ISBN: {{ book.isbn }}</p>
            <p v-if="book.rating" class="rating">Rating: {{ '★'.repeat(book.rating) }}{{ '☆'.repeat(10 - book.rating) }} ({{ book.rating }}/10)</p>
            <p v-if="book.comments" class="comments">Comments: {{ book.comments }}</p>
            <div class="book-actions">
                <button class="edit-button secondary outline" @click.stop="$emit('edit', book)">Edit</button>
                <!-- <button class="delete-button contrast outline" @click.stop="deleteBook(book.id)">Delete</button> -->
            </div>
        </div>
    `,
    methods: {
        onImgError(event) {
            // Replace broken image with placeholder
            event.target.outerHTML = '<div class="placeholder">No Cover</div>';
        }
    }
};


const app = createApp({
    components: {
        BookCard
    },
    setup() {
        // --- Reactive State ---
        const books = ref([]); // Array of all books from the backend
        const searchQuery = ref('');
        const searchResults = ref([]);
        const searchAttempted = ref(false); // To know if a search was run
        const currentEditBook = ref(null); // Holds the book object being edited
        const editFormData = reactive({ // Form data for the edit modal
            rating: null,
            comments: ''
        });
        const editModalRef = ref(null); // Template ref for the dialog element

        const loading = reactive({
            books: false,
            search: false,
            add: null, // Store OLID of book being added
            edit: false,
            statusUpdate: false,
        });

        const status = reactive({ // For user feedback messages
            bookshelf: { message: '', isError: false },
            search: { message: '', isError: false },
            edit: { message: '', isError: false },
        });

        // --- Utility Functions ---
        function setStatusMessage(area, message, isError = false, clearAfter = 5000) {
            if (status[area]) {
                status[area].message = message;
                status[area].isError = isError;
                if (clearAfter > 0) {
                    setTimeout(() => {
                        if (status[area].message === message) { // Clear only if message hasn't changed
                             status[area].message = '';
                             status[area].isError = false;
                        }
                    }, clearAfter);
                }
            }
        }

        // --- API Functions ---
        async function fetchBooks() {
            loading.books = true;
            setStatusMessage('bookshelf', '');
            try {
                const response = await fetch(`${API_BASE_URL}/books`);
                if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
                books.value = await response.json();
                setStatusMessage('bookshelf', 'Books loaded.');
            } catch (error) {
                console.error('Error fetching books:', error);
                setStatusMessage('bookshelf', `Error loading books: ${error.message}`, true, 0); // Don't auto-clear error
            } finally {
                loading.books = false;
            }
        }

        async function searchBooks() {
            if (!searchQuery.value.trim()) {
                setStatusMessage('search', 'Please enter a search term.', true);
                return;
            }
            loading.search = true;
            searchAttempted.value = true;
            searchResults.value = [];
            setStatusMessage('search', 'Searching...');
            try {
                const response = await fetch(`${API_BASE_URL}/search?q=${encodeURIComponent(searchQuery.value)}`);
                if (!response.ok) {
                    const errorData = await response.json().catch(() => ({ error: 'Search request failed' }));
                    throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
                }
                searchResults.value = await response.json();
                 setStatusMessage('search', searchResults.value.length > 0 ? `${searchResults.value.length} results found.` : 'No results found.');
            } catch (error) {
                console.error('Error searching Open Library:', error);
                setStatusMessage('search', `Search failed: ${error.message}`, true, 0);
            } finally {
                loading.search = false;
            }
        }

        async function addBook(bookData) {
            loading.add = bookData.open_library_id; // Indicate which book is being added
            setStatusMessage('search', `Adding "${bookData.title}"...`);
            try {
                // Prepare data for the backend API
                const payload = {
                    title: bookData.title,
                    author: bookData.author || 'Unknown Author',
                    open_library_id: bookData.open_library_id,
                    isbn: bookData.isbn || null,
                    cover_url: bookData.cover_url || null,
                    // Status defaults to "Want to Read" on the backend
                };
                const response = await fetch(`${API_BASE_URL}/books`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload),
                });
                 if (!response.ok) {
                    const errorData = await response.json().catch(() => ({ error: 'Failed to add book' }));
                    throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
                }
                const newBook = await response.json();
                // Add to local state immediately for responsiveness
                books.value.push(newBook);
                // Optionally sort books again if needed: books.value.sort((a, b) => a.title.localeCompare(b.title));
                setStatusMessage('search', `"${newBook.title}" added successfully!`);
                searchResults.value = []; // Clear search results
                searchQuery.value = ''; // Clear search input
                searchAttempted.value = false; // Reset search attempted flag
            } catch (error) {
                console.error('Error adding book:', error);
                // Check for unique constraint error (basic check)
                if (error.message && error.message.toLowerCase().includes('unique constraint')) {
                     setStatusMessage('search', `Error: Book "${bookData.title}" is already on your shelf.`, true, 0);
                } else {
                     setStatusMessage('search', `Error adding book: ${error.message}`, true, 0);
                }
            } finally {
                loading.add = null;
            }
        }

        async function updateBookStatus(bookId, newStatus) {
            // Find the book in the local state to update it optimistically
            const bookIndex = books.value.findIndex(b => b.id === parseInt(bookId, 10));
            if (bookIndex === -1) return; // Should not happen

            const originalStatus = books.value[bookIndex].status;
            // Optimistic update
            books.value[bookIndex].status = newStatus;

            loading.statusUpdate = true;
            setStatusMessage('bookshelf', 'Updating status...');

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
                setStatusMessage('bookshelf', 'Book status updated.');
            } catch (error) {
                console.error('Error updating book status:', error);
                setStatusMessage('bookshelf', `Error updating status: ${error.message}`, true, 0);
                // Revert optimistic update on error
                books.value[bookIndex].status = originalStatus;
            } finally {
                 loading.statusUpdate = false;
            }
        }

        async function updateBookDetails(bookId, rating, comments) {
            loading.edit = true;
            setStatusMessage('edit', 'Saving changes...');
            try {
                const payload = {};
                // Handle empty string from input as null for rating
                payload.rating = (rating === '' || rating === null || isNaN(parseInt(rating, 10))) ? null : parseInt(rating, 10);
                // Handle empty string from input as null for comments
                payload.comments = (comments === '' || comments === null) ? null : comments;

                // Validation (redundant with backend but good for UX)
                if (payload.rating !== null && (payload.rating < 1 || payload.rating > 10)) {
                    throw new Error("Rating must be between 1 and 10, or empty.");
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

                // Update local data on success
                const bookIndex = books.value.findIndex(b => b.id === bookId);
                if (bookIndex !== -1) {
                    books.value[bookIndex].rating = payload.rating;
                    books.value[bookIndex].comments = payload.comments;
                }

                setStatusMessage('edit', 'Changes saved successfully!');
                closeEditModal(); // Close modal on success

            } catch (error) {
                console.error('Error updating book details:', error);
                setStatusMessage('edit', `Error saving: ${error.message}`, true, 0);
            } finally {
                loading.edit = false;
            }
        }

        // --- Computed Properties ---
        const wantToReadBooks = computed(() => books.value.filter(b => b.status === 'Want to Read'));
        const currentlyReadingBooks = computed(() => books.value.filter(b => b.status === 'Currently Reading'));
        const readBooks = computed(() => books.value.filter(b => b.status === 'Read'));

        // --- Modal Logic ---
        function openEditModal(book) {
            currentEditBook.value = book;
            // Reset form data based on the book being edited
            // Use nullish coalescing for potentially null values
            editFormData.rating = book.rating ?? null; // Keep null if no rating
            editFormData.comments = book.comments ?? ''; // Use empty string for textarea if null
            setStatusMessage('edit', ''); // Clear previous status
            // Use the template ref to access the dialog element
            if (editModalRef.value) {
                editModalRef.value.showModal();
                 // Optional: focus first input
                 nextTick(() => {
                    const ratingInput = editModalRef.value.querySelector('#edit-rating');
                    if(ratingInput) ratingInput.focus();
                 });
            }
        }

        function closeEditModal() {
            if (editModalRef.value) {
                editModalRef.value.close();
            }
            currentEditBook.value = null; // Clear the book being edited
        }

        function saveEdit() {
            if (currentEditBook.value) {
                updateBookDetails(currentEditBook.value.id, editFormData.rating, editFormData.comments);
            }
        }

        // --- SortableJS Integration ---
        function initializeSortable() {
             // Ensure Sortable is loaded
            if (typeof Sortable === 'undefined') {
                console.error("SortableJS not loaded!");
                return;
            }

            const lists = document.querySelectorAll('.book-list');
            lists.forEach(list => {
                // Destroy previous instance if exists (important for reactivity)
                if (list.sortableInstance) {
                    list.sortableInstance.destroy();
                }
                // Create new instance
                list.sortableInstance = new Sortable(list, {
                    group: 'bookshelf',
                    animation: 150,
                    ghostClass: 'sortable-ghost',
                    chosenClass: 'sortable-chosen',
                    draggable: '.book-card', // Specify draggable elements
                    onEnd: (evt) => {
                        const bookId = evt.item.dataset.id;
                        const newStatus = evt.to.dataset.status;
                        const oldStatus = evt.from.dataset.status;

                        // Only trigger update if moved to a *different* list
                        if (bookId && newStatus && oldStatus !== newStatus) {
                            console.log(`Book ID ${bookId} moved from ${oldStatus} to ${newStatus}`);
                            // Call the Vue method to handle the API update and state change
                            updateBookStatus(bookId, newStatus);
                        } else {
                             console.log(`Book ID ${bookId} moved within the same list or data missing.`);
                             // If moved within the same list, Vue's computed properties handle the visual order change automatically
                             // if the underlying `books` array order were changed, but SortableJS handles the DOM directly here.
                             // For visual consistency after same-list drag, could force re-render or manually reorder `books.value`
                             // but it's often not necessary unless strict order persistence is required.
                        }
                    },
                });
            });
        }


        // --- Lifecycle Hooks ---
        onMounted(async () => {
            await fetchBooks();
            // Initialize SortableJS after the DOM is updated by Vue
            await nextTick(); // Wait for Vue to render the fetched books
            initializeSortable();

             // Add event listener for closing modal via backdrop click
            if (editModalRef.value) {
                editModalRef.value.addEventListener('click', (event) => {
                    if (event.target === editModalRef.value) {
                        closeEditModal();
                    }
                });
            }
        });

        // --- Return values for template ---
        return {
            books,
            searchQuery,
            searchResults,
            searchAttempted,
            currentEditBook,
            editFormData,
            editModalRef,
            loading,
            status,
            fetchBooks,
            searchBooks,
            addBook,
            updateBookStatus, // Although called internally by SortableJS, expose if needed
            openEditModal,
            closeEditModal,
            saveEdit,
            wantToReadBooks,
            currentlyReadingBooks,
            readBooks,
        };
    }
});

// Mount the app
app.mount('#app');
