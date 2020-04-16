package db

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"evl-book-server/config"
	"log"
)

func AddDefaultAdmin() {
	// ensuring default admin account
	user := config.UserCredentials{
		Username: "admin",
		Password: GetMD5Hash("admin"),
		UserData: config.UserData{
			IsAdmin:       true,
			Name:          "",
			ProfilePicURL: "",
		},
	}

	// beyond this block, the user's credentials are acceptable.
	// process and save them in db
	userBytes, err := json.Marshal(user)
	if err != nil {
		log.Println("error encoding admin data")
		return
	}
	_, err = GetSingleValue("user_" + user.Username)
	if err != nil {
		if err.Error() == RedisNilErr {
			SetJsonValues("user_"+user.Username, userBytes)
		} else {
			log.Println("error getting admin data")
		}
	}

}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
