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
)

func BookCreateHandler(w http.ResponseWriter, r *http.Request) {
	// assuming that we will receive json as signup form
	book := getBookDetails(r)

	// check for inconsistencies
	validBook, err := ValidateBookCreate(book)
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

	w.Write([]byte("book added successfully"))
}

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

	w.Write([]byte("book added successfully"))
}

func BookDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bookID := vars["id"]

	bookKey := BookPrefix + bookID

	err := db.RemoveByKey(bookKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("book deleted successfully"))
}

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

	w.Write(book)
}

func GetAllBooksHandler(w http.ResponseWriter, r *http.Request) {

	bookKeys, err := db.ScanKeysByPrefix(BookPrefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(bookKeys) == 0 {
		w.Write([]byte("no book has been added yet"))
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
	w.Write([]byte(resultString))
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
	if book.AuthorID != 0 {
		// look for the author id, if author exists, ad this book to his collection
		// else return nil
		log.Println("yet to implement author")
	}

	bookKey := BookPrefix + strconv.Itoa(book.ID)
	ok, err := isBookExistInDB(bookKey)
	if err != nil {
		return config.Book{}, err
	}
	if !ok {
		// book is new
		book.TotalCount = book.AddToCount
		book.AddToCount = 0
		book.OnLoanCount = 0
		// TODO: add book to author, if author doesn't exist, create one
		return book, nil
	}
	// book is old
	return book, errors.New("book already exists")
}

func isBookExistInDB(key string) (bool, error) {
	_, err := db.GetSingleValue(key)
	if err != nil && err.Error() != db.RedisNilErr {
		return false, err
	} else if err.Error() == db.RedisNilErr {
		//book doesnt already exist
		return false, nil
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
		//TODO: look if the new author exists, if not refuse the update
		//TODO: if new author exists, add this book to his collection and remove from old authors collection
	}

	book.TotalCount = savedBook.TotalCount
	book.OnLoanCount = savedBook.OnLoanCount
	if book.AddToCount > 0 {
		book.TotalCount += book.AddToCount
	}

	return book, nil
}