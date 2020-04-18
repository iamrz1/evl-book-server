package routes

import (
	"encoding/json"
	"errors"
	"evl-book-server/config"
	"evl-book-server/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

const (
	AuthorPrefix = "author_"
	BookPrefix   = "book_"
	UserPrefix   = "user_"
	LoanPrefix   = "loan_"
)

// BookCreateHandler creates a new book using the given JSON
func BookCreateHandler(w http.ResponseWriter, r *http.Request) {
	// assuming that we will receive json as signup form
	book := getBookDetails(r)

	// check for inconsistencies
	validBook, err := ValidateBookCreate(book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err.Error() == db.RedisNilErr {
			http.Error(w, "author doesn't exist", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	bookKey := BookPrefix + strconv.Itoa(book.ID)
	bookBytes, err := json.Marshal(validBook)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = db.SetJsonValues(bookKey, bookBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("book added successfully"))
}

// BookUpdateHandler updates a book's info using the given JSON
func BookUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// assuming that we will receive json as signup form
	book := getBookDetails(r)

	// check for inconsistencies
	validBook, err := ValidateBookUpdate(book)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bookKey := BookPrefix + strconv.Itoa(book.ID)
	bookBytes, err := json.Marshal(validBook)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = db.SetJsonValues(bookKey, bookBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("book added successfully"))
}

// BookDeleteHandler deletes a book by the given ID
func BookDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["id"]

	bookKey := BookPrefix + bookID

	err := db.RemoveByKey(bookKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("book deleted successfully"))
}

// GetBookHandler returns a book's info by bookID
func GetBookHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["id"]

	bookKey := BookPrefix + bookID

	book, err := db.GetByteValues(bookKey)
	if err != nil {
		if err.Error() == db.RedisNilErr {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(book)
}

// GetAllBooksHandler returns an array of all books' info
func GetAllBooksHandler(w http.ResponseWriter, _ *http.Request) {

	bookKeys, err := db.ScanKeysByPrefix(BookPrefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(bookKeys) == 0 {
		_, _ = w.Write([]byte("no book has been added yet"))
		return
	}
	resultString := "["
	for _, bookKey := range bookKeys {
		book, err := db.GetSingleValue(bookKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if resultString != "[" {
			resultString += ","
		}
		resultString += book
	}
	resultString += "]"
	_, _ = w.Write([]byte(resultString))
}

func getBookDetails(r *http.Request) config.Book {
	book := config.Book{}
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		// If there is something wrong with the request body, return a nil structure
		return config.Book{}
	}
	return book
}

func ValidateBookCreate(book config.Book) (config.Book, error) {
	if book.BookName == "" || book.ID == 0 {
		return config.Book{}, errors.New("book name or ID is missing")
	}
	log.Println("book author id=", book.AuthorID)
	author := config.Author{}
	authorKey := AuthorPrefix + strconv.Itoa(book.AuthorID)
	var err error
	err = nil
	if book.AuthorID != 0 {
		author, err = getAuthorByKeyFromDB(authorKey)
		if err != nil {
			return config.Book{}, err
		}
		author.AuthoredBookIDs = append(author.AuthoredBookIDs, book.ID)
	}

	bookKey := BookPrefix + strconv.Itoa(book.ID)
	ok, err := isBookExistInDB(bookKey)
	if err != nil {
		return config.Book{}, err
	}
	if !ok {
		// book is new
		book.TotalCount = book.AddCount
		if book.TotalCount == 0 {
			book.TotalCount = 1
		}
		book.AddCount = 0
		book.OnLoanCount = 0
		 if author.ID != 0 {
			 authorBytes, err := json.Marshal(author)
			 if err != nil {
				 return config.Book{}, err
			 }
			 _ = db.SetJsonValues(authorKey,authorBytes)
		 }

		return book, nil
	}
	// book is old
	return book, errors.New("book already exists")
}

func isBookExistInDB(key string) (bool, error) {
	_, err := db.GetSingleValue(key)
	if err != nil {
		if err.Error() != db.RedisNilErr {
			return false, err
		} else {
			//book doesnt already exist
			return false, nil
		}
	}
	// book exists
	return true, nil
}

func ValidateBookUpdate(book config.Book) (config.Book, error) {
	if book.BookName == "" || book.ID == 0 {
		return config.Book{}, errors.New("book name or ID is missing")
	}
	log.Println("book author id=", book.AuthorID)
	bookKey := BookPrefix + strconv.Itoa(book.ID)
	ok, err := isBookExistInDB(bookKey)
	if err != nil {
		return config.Book{}, err
	}
	if !ok {
		// book is new
		return book, errors.New("book does not exist")
	}

	// book is old
	savedBook := config.Book{}
	savedBookByte, _ := db.GetByteValues(bookKey)

	if err := json.Unmarshal(savedBookByte, &savedBook); err != nil {
		return config.Book{}, err
	}

	if book.AuthorID != 0 {
		// look for the author id, if author exists, ad this book to his collection
		// else return nil
		log.Println("yet to implement author")
	}
	if savedBook.AuthorID != book.AuthorID {
		if book.AuthorID != 0 {
			authorKey := AuthorPrefix + strconv.Itoa(book.AuthorID)
			author, err := getAuthorByKeyFromDB(authorKey)
			if err != nil {
				return config.Book{}, err
			}
			author.AuthoredBookIDs = append(author.AuthoredBookIDs, book.ID)
			authorByte, err := json.Marshal(author)
			if err != nil {
				return config.Book{}, err
			}
			_ = db.SetJsonValues(authorKey, authorByte)
		}
		if savedBook.AuthorID != 0 {
			authorKey := AuthorPrefix + strconv.Itoa(savedBook.AuthorID)
			author, err := getAuthorByKeyFromDB(authorKey)
			// we dont have to block update for any error here
			if err == nil {
				author.AuthoredBookIDs = append(author.AuthoredBookIDs, book.ID)
				author.AuthoredBookIDs = RemoveElementFromArray(author.AuthoredBookIDs, savedBook.AuthorID)
				authorByte, err := json.Marshal(author)
				if err != nil {
					return config.Book{}, err
				}
				_ = db.SetJsonValues(authorKey, authorByte)
			}

		}
	}

	book.TotalCount = savedBook.TotalCount
	book.OnLoanCount = savedBook.OnLoanCount
	if book.AddCount > 0 {
		book.TotalCount += book.AddCount
	}

	return book, nil
}

func getAuthorByKeyFromDB(authorKey string) (config.Author, error) {
	authorByte, err := db.GetByteValues(authorKey)
	if err != nil {
		return config.Author{}, err
	}
	author := config.Author{}

	if err := json.Unmarshal(authorByte, &author); err != nil {
		return config.Author{}, err
	}

	return author, nil
}
