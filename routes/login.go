package routes

import (
	"encoding/base64"
	"encoding/json"
	"evl-book-server/auth"
	"evl-book-server/config"
	"net/http"
	"strings"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	user := getCredentials(r)
	if user.Username == "" || user.Password == ""{
		w.WriteHeader(http.StatusForbidden)
		return
	}

	//validate user credentials
	if strings.ToLower(user.Username) != "rezoan" {
		if user.Password != "abc123" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	}

	// TODO: use isAuthenticated() to validate user credentials

	if !isAuthenticated(user.Username,user.Password){
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//Generate token
	token, err := auth.GenerateJWT(user.Username)
	if err != nil {
		return
	}

	//create a token instance using the token string
	JsonResponse(token, w)

}

func JsonResponse(response interface{}, w http.ResponseWriter) {
	json, err :=  json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func getCredentials(r *http.Request) config.UserCredentials {
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
	credSplitArray := strings.SplitN(credential,":",2)
	// check with database

	return config.UserCredentials{
		Username: credSplitArray[0],
		Password: credSplitArray[1],
	}
}

func isAuthenticated(username, password string) bool {
	authenticated := true
	// TODO: validate against record in db

	return authenticated
}