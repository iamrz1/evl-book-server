package routes

import (
	"net/http"
)

func HomePageHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World. This is a placeholder for actual homepage from URL: " + r.URL.String()))
}
