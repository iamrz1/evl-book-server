package routes

import (
	"evl-book-server/auth"
	"fmt"
	"net/http"
)

func HomePageHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Hello World. This is a placeholder for URL: " + r.URL.String() + "\n"))
	_, _ = w.Write([]byte(fmt.Sprintf("User : %s \n", r.Header.Get(auth.UsernameKey))))
	_, _ = w.Write([]byte(fmt.Sprintf("Admin : %s \n", r.Header.Get(auth.AdminKey))))

}
