package routes

import (
	"encoding/json"
	"evl-book-server/auth"
	"evl-book-server/db"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func ImageUploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("uploading image")
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("token not received")
	}

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, _, err := r.FormFile(FileID)
	if err != nil {
		fmt.Println("error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Create a temporary file within our server directory that follows
	// a particular naming pattern

	username := r.Header.Get(auth.UsernameKey)
	tempFile, err := ioutil.TempFile("image-server", fmt.Sprintf("%s_*.png", username))
	if err != nil {
		fmt.Fprintf(w, "%s\n", err.Error())
		fmt.Println(err)
		return
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "%s\n", err.Error())
		return
	}
	// write this byte array to our temporary file
	_, err = tempFile.Write(fileBytes)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "%s\n", err.Error())
		return
	}
	// save the link to users profile
	userKey := UserPrefix + username
	user, err := getUserByKey(userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s\n", err.Error())
		return
	}

	user.UserData.ProfilePicURL = tempFile.Name()
	userBytes, err := json.Marshal(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "%s\n", err.Error())
		return
	}
	_ = db.SetJsonValues(userKey, userBytes)

	fmt.Fprintf(w, "successfully Uploaded image\n")
}
