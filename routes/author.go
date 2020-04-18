package routes

import (
	"encoding/json"
	"errors"
	"evl-book-server/config"
	"evl-book-server/db"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// AuthorCreateHandler creates a new author using the given JSON
func AuthorCreateHandler(w http.ResponseWriter, r *http.Request) {
	// assuming that we will receive json as signup form
	author := getAuthorDetails(r)

	// check for inconsistencies
	validAuthor, err := validateAuthorCreate(author)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// save author to db
	authorKey := AuthorPrefix + strconv.Itoa(author.ID)
	authorBytes, err := json.Marshal(validAuthor)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = db.SetJsonValues(authorKey, authorBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("author added successfully"))
}

// AuthorUpdateHandler updates author info using the given JSON
func AuthorUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// assuming that we will receive json
	author := getAuthorDetails(r)

	// check for inconsistencies
	validAuthor, err := validateAuthorUpdate(author)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	authorKey := AuthorPrefix + strconv.Itoa(author.ID)
	authorBytes, err := json.Marshal(validAuthor)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = db.SetJsonValues(authorKey, authorBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("author added successfully"))
}

// AuthorDeleteHandler deletes an author by ID
func AuthorDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	authorID := vars["id"]

	authorKey := AuthorPrefix + authorID

	err := db.RemoveByKey(authorKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("author deleted successfully"))
}

// GetAuthorHandler returns an author's info by authorID
func GetAuthorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	authorID := vars["id"]

	authorKey := AuthorPrefix + authorID

	author, err := db.GetByteValues(authorKey)
	if err != nil {
		if err.Error() == db.RedisNilErr {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write(author)
}

// GetAllAuthorsHandler returns an array of all authors' info
func GetAllAuthorsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/jsonResponse")
	authorKeys, err := db.ScanKeysByPrefix(AuthorPrefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(authorKeys) == 0 {
		_, _ = w.Write([]byte("no author has been added yet"))
		return
	}
	resultString := "["
	for _, authorKey := range authorKeys {
		author, err := db.GetSingleValue(authorKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if resultString != "[" {
			resultString += ","
		}
		resultString += author
	}
	resultString += "]"
	_, _ = w.Write([]byte(resultString))
}

func getAuthorDetails(r *http.Request) config.Author {
	author := config.Author{}
	err := json.NewDecoder(r.Body).Decode(&author)
	if err != nil {
		// If there is something wrong with the request body
		return config.Author{}
	}
	return author
}

func validateAuthorCreate(author config.Author) (config.Author, error) {
	if author.AuthorName == "" || author.ID == 0 {
		return config.Author{}, errors.New("author name or ID is missing")
	}

	authorKey := AuthorPrefix + strconv.Itoa(author.ID)
	ok, err := isAuthorExistInDB(authorKey)
	if err != nil {
		return config.Author{}, err
	}
	if !ok {
		return author, nil
	}
	// author already exists
	return config.Author{}, errors.New("author already exists")
}

func isAuthorExistInDB(key string) (bool, error) {
	_, err := db.GetSingleValue(key)
	if err != nil && err.Error() != db.RedisNilErr {
		return false, err
	}
	if err != nil && err.Error() == db.RedisNilErr {
		//author doesnt already exist
		return false, nil
	}
	// author exists
	return true, nil
}

func validateAuthorUpdate(author config.Author) (config.Author, error) {
	if author.AuthorName == "" || author.ID == 0 {
		return config.Author{}, errors.New("author name or ID is missing")
	}
	authorKey := AuthorPrefix + strconv.Itoa(author.ID)
	ok, err := isAuthorExistInDB(authorKey)
	if err != nil {
		return config.Author{}, err
	}
	if !ok {
		// author is new
		return author, errors.New("author does not exist")
	}
	// author is old
	return author, nil
}
