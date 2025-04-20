const app = Vue.createApp({
    data() {
        return {
            books: [],
            shelves: [],
            searchTerm: '',
            newBook: {
                title: '',
                author: '',
                rating: 0
            },
            editingBook: null,
            bookToDelete: null,
            currentCategory: null, // 'shelf' or 'author' or null
            currentFilter: null,  // specific shelf or author
            sortField: 'title',  // default sort by title
            sortOrder: 'asc',
            selectedShelves: [],
            booksInShelf: {},
            addBookModal: null,
            deleteBookModal: null,
            toastMessage: '',
            showToast: false
        }
    },
    computed: {
        pageTitle() {
            if (!this.currentCategory) return 'All Books';
            if (this.currentCategory === 'shelf' && this.currentFilter) {
                return `Books in "${this.currentFilter.name}" Shelf`;
            }
            if (this.currentCategory === 'author' && this.currentFilter) {
                return `Books by ${this.currentFilter}`;
            }
            return this.currentCategory === 'shelf' ? 'Books by Shelf' : 'Books by Author';
        },
        
        // Get unique list of authors from books
        authors() {
            const authorSet = new Set();
            this.books.forEach(book => authorSet.add(book.author));
            return Array.from(authorSet).sort();
        },
        
        // Sort shelves to put "Want to Read" first
        sortedShelves() {
            return [...this.shelves].sort((a, b) => {
                // "Want to Read" first
                if (a.name === "Want to Read") return -1;
                if (b.name === "Want to Read") return 1;
                
                // Then "Currently Reading"
                if (a.name === "Currently Reading") return -1;
                if (b.name === "Currently Reading") return 1;
                
                // Then "Read"
                if (a.name === "Read") return -1;
                if (b.name === "Read") return 1;
                
                // Then alphabetical
                return a.name.localeCompare(b.name);
            });
        },
        
        // Filter books based on search term
        filteredBooks() {
            if (this.searchTerm) {
                const search = this.searchTerm.toLowerCase();
                return this.books.filter(book => 
                    book.title.toLowerCase().includes(search) || 
                    book.author.toLowerCase().includes(search)
                );
            }
            return this.books;
        },
        
        // Sort the filtered books
        sortedFilteredBooks() {
            return [...this.filteredBooks].sort((a, b) => {
                let valueA, valueB;
                
                if (this.sortField === 'title') {
                    valueA = a.title.toLowerCase();
                    valueB = b.title.toLowerCase();
                } else if (this.sortField === 'author') {
                    valueA = a.author.toLowerCase();
                    valueB = b.author.toLowerCase();
                } else if (this.sortField === 'created_at') {
                    // Convert to date objects if they're strings
                    valueA = a.created_at ? new Date(a.created_at) : new Date(0);
                    valueB = b.created_at ? new Date(b.created_at) : new Date(0);
                } else {
                    valueA = a[this.sortField];
                    valueB = b[this.sortField];
                }
                
                if (this.sortOrder === 'asc') {
                    return valueA > valueB ? 1 : -1;
                } else {
                    return valueA < valueB ? 1 : -1;
                }
            });
        },
        
        // Group books by category (shelf or author)
        groupedBooks() {
            if (!this.currentCategory) return {};
            
            const result = {};
            
            if (this.currentCategory === 'shelf') {
                // Initialize with all shelves, even empty ones
                this.shelves.forEach(shelf => {
                    result[shelf.name] = [];
                });
                
                // Add books to their shelves
                for (const shelfId in this.booksInShelf) {
                    const shelf = this.shelves.find(s => s.id === shelfId);
                    if (shelf) {
                        // Filter books if there's a search term
                        let shelfBooks = this.booksInShelf[shelfId];
                        if (this.searchTerm) {
                            const search = this.searchTerm.toLowerCase();
                            shelfBooks = shelfBooks.filter(book => 
                                book.title.toLowerCase().includes(search) || 
                                book.author.toLowerCase().includes(search)
                            );
                        }
                        
                        // Sort the books
                        shelfBooks = [...shelfBooks].sort((a, b) => {
                            let valueA, valueB;
                            if (this.sortField === 'title') {
                                valueA = a.title.toLowerCase();
                                valueB = b.title.toLowerCase();
                            } else if (this.sortField === 'author') {
                                valueA = a.author.toLowerCase();
                                valueB = b.author.toLowerCase();
                            } else if (this.sortField === 'created_at') {
                                valueA = a.created_at ? new Date(a.created_at) : new Date(0);
                                valueB = b.created_at ? new Date(b.created_at) : new Date(0);
                            } else {
                                valueA = a[this.sortField];
                                valueB = b[this.sortField];
                            }
                            
                            if (this.sortOrder === 'asc') {
                                return valueA > valueB ? 1 : -1;
                            } else {
                                return valueA < valueB ? 1 : -1;
                            }
                        });
                        
                        result[shelf.name] = shelfBooks;
                    }
                }
                
                // If filtering by a specific shelf, only return that one
                if (this.currentFilter) {
                    const filteredResult = {};
                    filteredResult[this.currentFilter.name] = result[this.currentFilter.name];
                    return filteredResult;
                }
                
                // Reorder keys to have "Want to Read" first
                const orderedResult = {};
                if (result["Want to Read"]) orderedResult["Want to Read"] = result["Want to Read"];
                if (result["Currently Reading"]) orderedResult["Currently Reading"] = result["Currently Reading"];
                if (result["Read"]) orderedResult["Read"] = result["Read"];
                
                // Add the rest alphabetically
                Object.keys(result)
                    .filter(key => !["Want to Read", "Currently Reading", "Read"].includes(key))
                    .sort()
                    .forEach(key => {
                        orderedResult[key] = result[key];
                    });
                
                return orderedResult;
                
            } else if (this.currentCategory === 'author') {
                // Group by author
                let filteredBooks = this.filteredBooks;
                
                // If filtering by a specific author, only include that one
                if (this.currentFilter) {
                    filteredBooks = filteredBooks.filter(book => book.author === this.currentFilter);
                }
                
                // Group books by author
                filteredBooks.forEach(book => {
                    if (!result[book.author]) {
                        result[book.author] = [];
                    }
                    result[book.author].push(book);
                });
                
                // Sort books within each author group
                for (const author in result) {
                    result[author].sort((a, b) => {
                        let valueA, valueB;
                        if (this.sortField === 'title') {
                            valueA = a.title.toLowerCase();
                            valueB = b.title.toLowerCase();
                        } else if (this.sortField === 'created_at') {
                            valueA = a.created_at ? new Date(a.created_at) : new Date(0);
                            valueB = b.created_at ? new Date(b.created_at) : new Date(0);
                        } else {
                            valueA = a[this.sortField];
                            valueB = b[this.sortField];
                        }
                        
                        if (this.sortOrder === 'asc') {
                            return valueA > valueB ? 1 : -1;
                        } else {
                            return valueA < valueB ? 1 : -1;
                        }
                    });
                }
                
                // Sort authors alphabetically
                const sortedResult = {};
                Object.keys(result)
                    .sort()
                    .forEach(author => {
                        sortedResult[author] = result[author];
                    });
                
                return sortedResult;
            }
            
            return {};
        },
        
        // Total number of books (for empty state)
        totalBooks() {
            return this.books.length;
        }
    },
    methods: {
        async fetchBooks() {
            try {
                const response = await axios.get('/api/db/books');
                this.books = response.data;
            } catch (error) {
                console.error('Error fetching books:', error);
                // Fallback to in-memory API if database API fails
                try {
                    const fallbackResponse = await axios.get('/api/books');
                    this.books = fallbackResponse.data;
                } catch (fallbackError) {
                    console.error('Error fetching from fallback API:', fallbackError);
                }
            }
        },
        
        async fetchShelves() {
            try {
                const response = await axios.get('/api/shelves');
                this.shelves = response.data;
                
                // Pre-fetch books for each shelf
                for (const shelf of this.shelves) {
                    this.fetchBooksInShelf(shelf.id);
                }
            } catch (error) {
                console.error('Error fetching shelves:', error);
            }
        },
        
        async fetchBooksInShelf(shelfId) {
            try {
                const response = await axios.get(`/api/shelves/${shelfId}/books`);
                this.$set(this.booksInShelf, shelfId, response.data);
            } catch (error) {
                console.error(`Error fetching books in shelf ${shelfId}:`, error);
            }
        },
        
        showAllBooks() {
            this.currentCategory = null;
            this.currentFilter = null;
        },
        
        showBooksByShelf(shelf) {
            this.currentCategory = 'shelf';
            this.currentFilter = shelf;
            this.fetchBooksInShelf(shelf.id);
        },
        
        showBooksByShelves() {
            this.currentCategory = 'shelf';
            this.currentFilter = null;
            // Make sure all shelves are fetched
            for (const shelf of this.shelves) {
                this.fetchBooksInShelf(shelf.id);
            }
        },
        
        showBooksByAuthor(author) {
            this.currentCategory = 'author';
            this.currentFilter = author;
        },
        
        showBooksByAuthors() {
            this.currentCategory = 'author';
            this.currentFilter = null;
        },
        
        sortBooks(field) {
            if (this.sortField === field) {
                // Toggle sort order if clicking the same field
                this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortField = field;
                this.sortOrder = 'asc'; // Default to ascending for new field
            }
        },
        
        showAddBookModal() {
            this.editingBook = null;
            this.newBook = {
                title: '',
                author: '',
                rating: 0
            };
            this.selectedShelves = [];
            this.addBookModal.show();
        },
        
        showEditBookModal(book) {
            this.editingBook = book;
            this.newBook = {
                id: book.id,
                title: book.title,
                author: book.author,
                rating: book.rating
            };
            
            // Find which shelves this book is in
            this.selectedShelves = [];
            for (const shelfId in this.booksInShelf) {
                const books = this.booksInShelf[shelfId];
                if (books.some(b => b.id === book.id)) {
                    this.selectedShelves.push(shelfId);
                }
            }
            
            this.addBookModal.show();
        },
        
        async saveBook() {
            try {
                let response;
                
                if (this.editingBook) {
                    // Update existing book
                    response = await axios.put(`/api/db/books/${this.newBook.id}`, this.newBook);
                    
                    // Update the book in our local array
                    const index = this.books.findIndex(b => b.id === this.newBook.id);
                    if (index !== -1) {
                        this.books[index] = response.data;
                    }
                } else {
                    // Create new book
                    response = await axios.post('/api/db/books', this.newBook);
                    
                    // Add the new book to our local array
                    this.books.push(response.data);
                }
                
                // Handle shelves
                const bookId = response.data.id;
                
                // First, we need to remove the book from all shelves
                if (this.editingBook) {
                    for (const shelfId in this.booksInShelf) {
                        const books = this.booksInShelf[shelfId];
                        if (books.some(b => b.id === bookId)) {
                            if (!this.selectedShelves.includes(shelfId)) {
                                // The book should be removed from this shelf
                                await axios.delete(`/api/shelves/${shelfId}/books/${bookId}`);
                            }
                        }
                    }
                }
                
                // Then add it to the selected shelves
                for (const shelfId of this.selectedShelves) {
                    await axios.post(`/api/shelves/${shelfId}/books/${bookId}`);
                }
                
                // Refresh data
                this.fetchBooks();
                for (const shelf of this.shelves) {
                    this.fetchBooksInShelf(shelf.id);
                }
                
                this.addBookModal.hide();
            } catch (error) {
                console.error('Error saving book:', error);
            }
        },
        
        confirmDeleteBook(book) {
            this.bookToDelete = book;
            this.deleteBookModal.show();
        },
        
        async deleteBook() {
            if (!this.bookToDelete) return;
            
            try {
                await axios.delete(`/api/db/books/${this.bookToDelete.id}`);
                
                // Remove the book from our local array
                this.books = this.books.filter(b => b.id !== this.bookToDelete.id);
                
                // Remove the book from all shelves in our local data
                for (const shelfId in this.booksInShelf) {
                    this.booksInShelf[shelfId] = this.booksInShelf[shelfId].filter(
                        b => b.id !== this.bookToDelete.id
                    );
                }
                
                this.deleteBookModal.hide();
            } catch (error) {
                console.error('Error deleting book:', error);
            }
        },
        
        async addBookToShelf(book, shelf) {
            try {
                await axios.post(`/api/shelves/${shelf.id}/books/${book.id}`);
                
                // Refresh the shelf data
                this.fetchBooksInShelf(shelf.id);
            } catch (error) {
                console.error('Error adding book to shelf:', error);
            }
        }
    },
    mounted() {
        // Initialize Bootstrap modals
        this.addBookModal = new bootstrap.Modal(document.getElementById('addBookModal'));
        this.deleteBookModal = new bootstrap.Modal(document.getElementById('deleteBookModal'));
        
        // Fetch initial data
        this.fetchBooks();
        this.fetchShelves();
        
        // Default view: show books by shelf
        this.currentCategory = 'shelf';
    }
});

// Helper function for Vue 3 compatibility
app.config.globalProperties.$set = function(obj, key, value) {
    if (Array.isArray(obj)) {
        obj.splice(key, 1, value);
        return value;
    }
    obj[key] = value;
    return value;
};

app.mount('#app');