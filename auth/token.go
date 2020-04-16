package auth

import (
	"evl-book-server/config"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"time"
)

const (
	AuthorizedKey = "authorized"
	UsernameKey   = "username"
	ExpirationKey = "exp"
	AdminKey      = "admin"
)

func GenerateJWT(user config.UserCredentials) (string, error) {
	appCfg := config.App()
	if appCfg.Key == "" {
		log.Panicln("Server needs a key to generate tokens")
	}
	var mySigningKey = []byte(appCfg.Key)
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims[AuthorizedKey] = true
	claims[UsernameKey] = user.Username
	claims[AdminKey] = user.UserData.IsAdmin
	claims[ExpirationKey] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		_ = fmt.Errorf("something went wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}
