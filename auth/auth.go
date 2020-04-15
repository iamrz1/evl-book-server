package auth

import (
	"evl-book-server/config"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"net/http"
	"strings"
	"time"
)

var (
	Store *sessions.CookieStore
)

type Auth struct{}

func (*Auth) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	t := time.Now()
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Need Bearer authorization! Generate token using your username and password here: http://localhost:<port>/api/login\n"))
		return
	}
	token, err:= jwt.Parse(strings.Split(authHeader, " ")[1], func(token *jwt.Token) (interface{}, error) {
		return []byte(config.App().Key), nil
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("token not valid"))
		return
	}
	claimMap := token.Claims.(jwt.MapClaims)
	for key, value := range claimMap{
		r.Header.Add(key,fmt.Sprintf("%v",value))
	}
	next(w, r)
	fmt.Printf("Execution time: %s \n", time.Now().Sub(t).String())
}
