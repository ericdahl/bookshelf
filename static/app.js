document.addEventListener('DOMContentLoaded', function() {
  const form = document.getElementById('book-form');
  const titleInput = document.getElementById('title');
  const authorInput = document.getElementById('author');
  const shelfInput = document.getElementById('shelf');
  const shelvesDatalist = document.getElementById('shelves');
  const booksContainer = document.getElementById('books-container');
  let shelves = [];

  function fetchShelves() {
    return fetch('/api/shelves')
      .then(response => response.json())
      .then(data => {
        shelves = data;
        populateShelvesDatalist();
      });
  }

  function populateShelvesDatalist() {
    shelvesDatalist.innerHTML = '';
    shelves.forEach(shelf => {
      const option = document.createElement('option');
      option.value = shelf;
      shelvesDatalist.appendChild(option);
    });
  }

  function fetchBooks() {
    return fetch('/api/books')
      .then(response => response.json())
      .then(data => {
        renderBooks(data);
      });
  }

  function renderBooks(books) {
    booksContainer.innerHTML = '';
    const booksByShelf = {};
    books.forEach(book => {
      if (!booksByShelf[book.shelf]) {
        booksByShelf[book.shelf] = [];
      }
      booksByShelf[book.shelf].push(book);
    });

    Object.keys(booksByShelf).forEach(shelf => {
      const section = document.createElement('div');
      section.className = 'mb-4';
      const header = document.createElement('h3');
      header.textContent = shelf;
      section.appendChild(header);

      const list = document.createElement('ul');
      list.className = 'list-group';

      booksByShelf[shelf].forEach(book => {
        const item = document.createElement('li');
        item.className = 'list-group-item d-flex justify-content-between align-items-center';

        const info = document.createElement('div');
        info.innerHTML = '<strong>' + book.title + '</strong> by ' + (book.author || 'Unknown');
        item.appendChild(info);

        const controls = document.createElement('div');

        const select = document.createElement('select');
        select.className = 'form-select d-inline-block me-2';
        select.style.width = 'auto';
        shelves.forEach(s => {
          const option = document.createElement('option');
          option.value = s;
          option.textContent = s;
          if (s === book.shelf) {
            option.selected = true;
          }
          select.appendChild(option);
        });
        select.addEventListener('change', function() {
          fetch('/api/books/' + book.id, {
            method: 'PUT',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
              title: book.title,
              author: book.author,
              shelf: select.value
            })
          }).then(function() {
            init();
          });
        });
        controls.appendChild(select);

        const delBtn = document.createElement('button');
        delBtn.className = 'btn btn-danger btn-sm';
        delBtn.textContent = 'Delete';
        delBtn.addEventListener('click', function() {
          if (confirm('Delete this book?')) {
            fetch('/api/books/' + book.id, { method: 'DELETE' })
              .then(function() {
                init();
              });
          }
        });
        controls.appendChild(delBtn);

        item.appendChild(controls);
        list.appendChild(item);
      });

      section.appendChild(list);
      booksContainer.appendChild(section);
    });
  }

  form.addEventListener('submit', function(e) {
    e.preventDefault();
    const data = {
      title: titleInput.value.trim(),
      author: authorInput.value.trim(),
      shelf: shelfInput.value.trim()
    };
    if (!data.title || !data.shelf) {
      alert('Title and shelf are required');
      return;
    }
    fetch('/api/books', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify(data)
    }).then(function() {
      form.reset();
      init();
    });
  });

  function init() {
    fetchShelves().then(function() {
      fetchBooks();
    });
  }

  init();
});