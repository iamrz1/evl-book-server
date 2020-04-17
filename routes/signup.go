package routes

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"evl-book-server/auth"
	"evl-book-server/config"
	"evl-book-server/db"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// AddUserHandler lets users sign up using
// a unique username and password, name field is optional
func AddUserHandler(w http.ResponseWriter, r *http.Request) {
	// assuming that we will receive json as signup form
	user := getJsonCredentials(r)
	if user.Username == "" || user.Password == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	ok, err := ValidateUsername(r, user.Username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !ok {
		w.Header().Set(ValidUserName, FalseString)
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("invalid username"))
		return
	}

	// beyond this block, the user's credentials are acceptable.
	// process and save them in db
	user.UserData = config.UserData{
		IsAdmin:       false,
		ProfilePicURL: "",
	}
	userBytes, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = db.SetJsonValues(UserPrefix+user.Username, userBytes)
	_, _ = w.Write([]byte("signed up successfully"))
}

func getJsonCredentials(r *http.Request) config.UserCredentials {
	cred := config.UserCredentials{}
	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		// If there is something wrong with the request body, return a  nil structure
		return config.UserCredentials{}
	}
	cred.Password = GetMD5Hash(cred.Password)

	return cred
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

//  ValidateUsername validates any input against a set of predefined
// conditions to validate username, using an existing api endpoint
func ValidateUsername(r *http.Request, username string) (bool, error) {
	resp, err := http.Get(fmt.Sprintf("%s://%s/api/validate/username/%s", config.App().Scheme, r.Host, username))
	if err != nil {
		return false, err
	}
	if resp.Header.Get(ValidUserName) == TrueString {
		return true, nil
	}
	return false, nil
}

// UpdateInfoHandler updates Name or Password, but not username or userinfo{}
func UpdateInfoHandler(w http.ResponseWriter, r *http.Request) {
	// assuming that we will receive json as signup form
	username := r.Header.Get(auth.UsernameKey)
	user := getJsonCredentials(r)
	if user.Username != "" && user.Username != username{
		_, _ = w.Write([]byte("changing username is not allowed"))
		w.WriteHeader(http.StatusForbidden)
		return
	}
	userKey := strings.ToLower(UserPrefix + username)

	savedUser, err := getUserByKey(userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user.Name == ""{
		user.Name = savedUser.Name
	}
	if user.Password == GetMD5Hash(""){
		user.Password = savedUser.Password
	}

	if savedUser.Name == user.Name && savedUser.Password == user.Password{
		_, _ = w.Write([]byte("no changes were made"))
		return
	}

	savedUser.Name = user.Name
	savedUser.Password = user.Password

	// save this update in db
	userBytes, err := json.Marshal(savedUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_ = db.SetJsonValues(userKey, userBytes)
	_, _ = w.Write([]byte("profile updated"))
}

func getUserByKey(userKey string) (config.UserCredentials, error) {
	userBytes, err := db.GetByteValues(strings.ToLower(userKey))
	if err != nil {
		log.Println("could not find user by key")
		return config.UserCredentials{}, err
	}
	user := config.UserCredentials{}
	if err := json.Unmarshal(userBytes, &user); err != nil {
		log.Println("could not find user by key")
		return config.UserCredentials{}, err
	}
	return user, nil
}
