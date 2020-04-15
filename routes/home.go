package routes

import (
	"evl-book-server/auth"
	"fmt"
	"net/http"
)

func HomePageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World. This is a placeholder for actual homepage from URL: " + r.URL.String()+"\n"))
	w.Write([]byte(fmt.Sprintf("User : %s \n", r.Header.Get(auth.UsernameKey))))
	w.WriteHeader(http.StatusOK)
}
