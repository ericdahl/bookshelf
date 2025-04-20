const app = Vue.createApp({
    data() {
        return {
            view: 'books',
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
            currentShelf: null,
            selectedShelves: [],
            booksInShelf: {},
            addBookModal: null,
            deleteBookModal: null
        }
    },
    computed: {
        filteredBooks() {
            let books = this.currentShelf ? 
                (this.booksInShelf[this.currentShelf.id] || []) : 
                this.books;

            if (this.searchTerm) {
                const search = this.searchTerm.toLowerCase();
                return books.filter(book => 
                    book.title.toLowerCase().includes(search) || 
                    book.author.toLowerCase().includes(search)
                );
            }
            return books;
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
                    alert('Failed to load books. Please try again later.');
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
                alert('Failed to load shelves. Please try again later.');
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
        
        showBooks() {
            this.currentShelf = null;
            this.view = 'books';
        },
        
        showShelfBooks(shelf) {
            this.currentShelf = shelf;
            this.view = 'books';
            this.fetchBooksInShelf(shelf.id);
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
                let successMessage;
                
                if (this.editingBook) {
                    // Update existing book
                    response = await axios.put(`/api/db/books/${this.newBook.id}`, this.newBook);
                    
                    // Update the book in our local array
                    const index = this.books.findIndex(b => b.id === this.newBook.id);
                    if (index !== -1) {
                        this.books[index] = response.data;
                    }
                    
                    successMessage = 'Book updated successfully!';
                } else {
                    // Create new book
                    response = await axios.post('/api/db/books', this.newBook);
                    
                    // Add the new book to our local array
                    this.books.push(response.data);
                    
                    successMessage = 'Book added successfully!';
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
                alert(successMessage);
                
            } catch (error) {
                console.error('Error saving book:', error);
                alert('Failed to save book. Please try again.');
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
                alert('Book deleted successfully!');
                
            } catch (error) {
                console.error('Error deleting book:', error);
                alert('Failed to delete book. Please try again.');
            }
        },
        
        async addBookToShelf(book, shelf) {
            try {
                await axios.post(`/api/shelves/${shelf.id}/books/${book.id}`);
                alert(`Added "${book.title}" to "${shelf.name}" shelf!`);
                
                // Refresh the shelf data
                this.fetchBooksInShelf(shelf.id);
                
            } catch (error) {
                console.error('Error adding book to shelf:', error);
                alert('Failed to add book to shelf. Please try again.');
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
    }
});

// Fix for Vue 3 compatibility with older code
app.config.globalProperties.$set = function(obj, key, value) {
    if (Array.isArray(obj)) {
        obj.splice(key, 1, value);
        return value;
    }
    obj[key] = value;
    return value;
};

app.mount('#app');