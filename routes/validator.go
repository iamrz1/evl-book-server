package routes

import (
	"evl-book-server/auth"
	"evl-book-server/db"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"regexp"
	"strings"
)

const (
	ValidUserName           = "isUsernameValid"
	ErrorLogKey             = "log"
	TrueString              = "true"
	FalseString             = "false"
	UnsupportedCharacterErr = "input contains unsupported character(s)"
)

var (
	ReservedWords = []string{"http", "https", "www", "ftp", "admin", ".com", ".io", ".net", "login", "book_", "author_", "loan_", "user_"}
)

// ValidateUser endpoint is used to validate any potential username
//that users want to obtain for themselves using a predefined
//set of rules and existing usernames in database
func ValidateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := strings.TrimSpace(strings.ToLower(vars[auth.UsernameKey]))
	//validate user credentials
	for _, word := range ReservedWords {
		if username == word {
			w.Header().Set(ValidUserName, FalseString)
			w.Header().Set(ErrorLogKey, "username "+username+" is a reserved word")
			_, _ = w.Write([]byte(fmt.Sprintln(FalseString)))
			return
		}
	}

	if match, _ := regexp.MatchString("^([a-z])+([a-z0-9])*$", username); !match {
		w.Header().Set(ValidUserName, FalseString)
		w.Header().Set(ErrorLogKey, UnsupportedCharacterErr)
		_, _ = w.Write([]byte(fmt.Sprintln(FalseString)))
		return
	}

	_, err := db.GetSingleValue(username)
	if err != nil {
		if err.Error() == db.RedisNilErr {
			w.Header().Set(ValidUserName, TrueString)
			_, _ = w.Write([]byte(TrueString))
			return
		}
		w.Header().Set(ValidUserName, FalseString)
		w.Header().Set(ErrorLogKey, "db err "+err.Error())
		_, _ = w.Write([]byte(FalseString))
		return
	}

	w.Header().Set(ValidUserName, FalseString)
	_, _ = w.Write([]byte(FalseString))
}
