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
	ReservedWords = []string{"http", "https", "www", "ftp", "admin", ".com", ".io", ".net", "login"}
)

func ValidateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := strings.TrimSpace(strings.ToLower(vars[auth.UsernameKey]))
	//validate user credentials
	for _, word := range ReservedWords {
		if username == word {
			w.Header().Set(ValidUserName, FalseString)
			w.Header().Set(ErrorLogKey, "username "+username+" is a reserved word")
			w.Write([]byte(fmt.Sprintln(FalseString)))
			return
		}
	}

	if match, _ := regexp.MatchString("^([a-z])+([a-z0-9])*$", username); !match {
		w.Header().Set(ValidUserName, FalseString)
		w.Header().Set(ErrorLogKey, UnsupportedCharacterErr)
		w.Write([]byte(fmt.Sprintln(FalseString)))
		return
	}

	_, err := db.GetSingleValue(username)
	if err != nil {
		if err.Error() == db.RedisNilErr {
			w.Header().Set(ValidUserName, TrueString)
			w.Write([]byte(TrueString))
			return
		}
		w.Header().Set(ValidUserName, FalseString)
		w.Header().Set(ErrorLogKey, "db err "+err.Error())
		w.Write([]byte(FalseString))
		return
	}

	w.Header().Set(ValidUserName, FalseString)
	w.Write([]byte(FalseString))
}
