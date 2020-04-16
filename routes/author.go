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

func AuthorCreateHandler(w http.ResponseWriter, r *http.Request) {
	// assuming that we will receive json as signup form
	author := getAuthorDetails(r)

	// check for inconsistencies
	validAuthor, err := ValidateAuthorCreate(author)
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

	w.Write([]byte("author added successfully"))
}

func AuthorUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// assuming that we will receive json as signup form
	author := getAuthorDetails(r)

	// check for inconsistencies
	validAuthor, err := ValidateAuthorUpdate(author)
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

	w.Write([]byte("author added successfully"))
}

func AuthorDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	authorID := vars["id"]

	authorKey := AuthorPrefix + authorID

	err := db.RemoveByKey(authorKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte("author deleted successfully"))
}

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

	w.Write(author)
}

func GetAllAuthorsHandler(w http.ResponseWriter, r *http.Request) {

	authorKeys, err := db.ScanKeysByPrefix(AuthorPrefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(authorKeys) == 0 {
		w.Write([]byte("no author has been added yet"))
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
	w.Write([]byte(resultString))
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

func ValidateAuthorCreate(author config.Author) (config.Author, error) {
	if author.AuthorName == "" || author.ID == 0 {
		return config.Author{}, errors.New("author name or ID is missing")
	}

	authorKey := AuthorPrefix + strconv.Itoa(author.ID)
	ok, err := isAuthorExistInDB(authorKey)
	if err != nil {
		return config.Author{}, err
	}
	if !ok {
		// TODO: add author to author, if author doesn't exist, create one
		return author, nil
	}
	// author is old
	return author, errors.New("author already exists")
}

func isAuthorExistInDB(key string) (bool, error) {
	_, err := db.GetSingleValue(key)
	if err != nil && err.Error() != db.RedisNilErr {
		return false, err
	} else if err.Error() == db.RedisNilErr {
		//author doesnt already exist
		return false, nil
	}
	// author exists
	return true, nil
}

func ValidateAuthorUpdate(author config.Author) (config.Author, error) {
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
