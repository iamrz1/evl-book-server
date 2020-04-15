package routes

import (
	"encoding/base64"
	"encoding/json"
	"evl-book-server/auth"
	"evl-book-server/config"
	"evl-book-server/db"
	"net/http"
	"strings"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	user := getBasicAuthCredentials(r)
	if user.Username == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// use db to verify credentials
	ok, user, err := UserAuthentication(user.Username, user.Password)
	if err != nil {
		if err.Error() == db.RedisNilErr{
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("user doesn't exist"))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//Generate token
	token, err := auth.GenerateJWT(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//create a token instance using the token string
	JsonResponse(token, w)

}

func JsonResponse(response interface{}, w http.ResponseWriter) {
	json, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func getBasicAuthCredentials(r *http.Request) config.UserCredentials {
	//Authorization in Header has the encoder-base and user-credentials encrypted in it
	encryptedAuthArr := strings.Split(r.Header.Get("Authorization"), " ")
	if len(encryptedAuthArr) != 2 {
		//log.Fatal("request body doesnt have proper authorization format")
		return config.UserCredentials{}
	}

	byteCredStr, err := base64.StdEncoding.DecodeString(encryptedAuthArr[1])
	if err != nil {
		//log.Fatal("couldn't decode credentials to string")
		return config.UserCredentials{}
	}
	credential := string(byteCredStr)
	//log.Println("credential = ", credential)
	credSplitArray := strings.SplitN(credential, ":", 2)
	// check with database

	return config.UserCredentials{
		Username: credSplitArray[0],
		Password: credSplitArray[1],
	}
}

func UserAuthentication(username, password string) (bool, config.UserCredentials, error) {
	userDetails, err := db.GetByteValues(username)
	if err != nil {
		return false,config.UserCredentials{}, err
	}

	user := config.UserCredentials{}
	if err := json.Unmarshal(userDetails, &user); err != nil {
		return false,config.UserCredentials{}, err
	}
	return GetMD5Hash(password) == user.Password,user ,nil
}
