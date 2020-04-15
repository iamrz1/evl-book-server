package routes

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"evl-book-server/config"
	"evl-book-server/db"
	"fmt"
	"net/http"
)

func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	// assuming that we will receive json as signup form
	user := getJsonCredentials(r)
	if user.Username == "" || user.Password == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ok, err := ValidateUsername(user.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !ok {
		w.Header().Set(ValidUserName, FalseString)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("invalid username"))
		return
	}

	// beyond this block, the user's credentials are acceptable.
	// process and save them in db
	user.UserData = config.UserData{
		IsAdmin:       false,
		Name:          "",
		ProfilePicURL: "",
	}
	userBytes, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	db.SetJsonValues(user.Username, userBytes)
	w.Write([]byte("signed up successfully"))
}

func getJsonCredentials(r *http.Request) config.UserCredentials {
	cred := config.UserCredentials{}
	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		// If there is something wrong with the request body, return a 400 status
		return config.UserCredentials{}
	}
	// check with database
	cred.Password = GetMD5Hash(cred.Password)
	return cred
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func ValidateUsername(username string) (bool, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/api/validate/username/%s", config.App().Port, username))
	if err != nil {
		return false, err
	}
	if resp.Header.Get(ValidUserName) == TrueString {
		return true, nil
	}
	return false, nil
}
