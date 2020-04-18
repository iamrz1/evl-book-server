package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"evl-book-server/config"
	"evl-book-server/db"
	"evl-book-server/routes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

const (
	username   = "test"
	password   = "test"
	admin      = "admin"
	bookID     = 1
	bookName   = "A Book"
	authorID   = 1
	authorName = "An Author"
	loanOne    = 1
	loanTwo    = 2
)

var (
	token      = ""
	adminToken = ""
)

func TestUserSignUP(t *testing.T) {
	user := config.UserCredentials{
		Username:    username,
		Name:        "dummy_user",
		Password:    password,
		UserData:    config.UserData{},
		LoanIDArray: nil,
	}
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(user)
	req, err := http.NewRequest(http.MethodPost, "http://localhost:3000/api/signup", buf)
	if err != nil {
		t.Error(err.Error())
	}
	getSingleOKResponse(t, req)
}

func TestAdminLogin(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://localhost:3000/api/login", nil)
	if err != nil {
		t.Error(err.Error())
	}
	req.Header.Set("Authorization", "Base "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", admin, admin))))
	adminToken = getSingleOKResponse(t, req)
}

func TestUserLogin(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "http://localhost:3000/api/login", nil)
	if err != nil {
		t.Error(err.Error())
	}
	req.Header.Set("Authorization", "Base "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))))
	token = getSingleOKResponse(t, req)
}

func TestCreateAuthor(t *testing.T) {
	author := &config.Author{
		ID:         bookID,
		AuthorName: authorName,
	}
	authorBytes, err := json.Marshal(author)
	url := "http://localhost:3000/api/admin/author/create"

	if err != nil {
		t.Error(err.Error())
		return
	}

	req1, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(authorBytes))
	if err != nil {
		t.Error(err.Error())
	}
	req1.Header.Set("Authorization", "Bearer "+adminToken)

	req2, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(authorBytes))
	if err != nil {
		t.Error(err.Error())
	}
	req2.Header.Set("Authorization", "Bearer "+token)
	requests := []*http.Request{req1, req2}
	statusOutArr := []int{http.StatusOK, http.StatusUnauthorized}
	getMultiPleResponse(t, requests, statusOutArr)
}

func TestGetAuthor(t *testing.T) {
	url := "http://localhost:3000/api/author"
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", url, authorID), nil)
	if err != nil {
		t.Error(err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)

	getSingleOKResponse(t, req)
}

func TestGetAllAuthors(t *testing.T) {
	url := "http://localhost:3000/api/authors"
	req1, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Error(err.Error())
	}

	req1.Header.Set("Authorization", "Bearer "+adminToken)

	req2, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}

	req2.Header.Set("Authorization", "Bearer "+token)

	req3, err := http.NewRequest(http.MethodGet, "http://localhost:3000/api/author", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	req3.Header.Set("Authorization", "Bearer "+token)
	requests := []*http.Request{req1, req2, req3}
	statusOutArr := []int{http.StatusOK, http.StatusOK, http.StatusNotFound}
	getMultiPleResponse(t, requests, statusOutArr)
}

func TestCreateBook(t *testing.T) {
	book := &config.Book{
		ID:       bookID,
		BookName: bookName,
		AuthorID: authorID,
	}
	bookBytes, err := json.Marshal(book)

	if err != nil {
		t.Error(err.Error())
		return
	}
	url := "http://localhost:3000/api/admin/book/create"

	req1, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bookBytes))
	if err != nil {
		t.Error(err.Error())
	}
	req1.Header.Set("Authorization", "Bearer "+adminToken)

	req2, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bookBytes))
	if err != nil {
		t.Error(err.Error())
	}
	req2.Header.Set("Authorization", "Bearer "+token)
	requests := []*http.Request{req1, req2}
	statusOutArr := []int{http.StatusOK, http.StatusUnauthorized}
	getMultiPleResponse(t, requests, statusOutArr)
}

func TestCreateLoan(t *testing.T) {
	url := "http://localhost:3000/api/loan/request/"

	req1, err := http.NewRequest(http.MethodPost, url+strconv.Itoa(loanOne), nil)
	if err != nil {
		t.Error(err.Error())
	}
	req1.Header.Set("Authorization", "Bearer "+token)

	req2, err := http.NewRequest(http.MethodPost, url+strconv.Itoa(loanTwo), nil)
	if err != nil {
		t.Error(err.Error())
	}
	req2.Header.Set("Authorization", "Bearer "+token)

	requests := []*http.Request{req1, req2}
	statusOutArr := []int{http.StatusOK, http.StatusOK}
	getMultiPleResponse(t, requests, statusOutArr)
}

func TestGetAllAndUserSpecificPendingLoans(t *testing.T) {
	allUrl := "http://localhost:3000/api/admin/loans"
	pendingUrl := "http://localhost:3000/api/admin/loans/pending"
	pendingUserUrl := "http://localhost:3000/api/loans/pending"
	req1, err := http.NewRequest(http.MethodGet, allUrl, nil)
	if err != nil {
		t.Error(err.Error())
	}

	req1.Header.Set("Authorization", "Bearer "+adminToken)

	req2, err := http.NewRequest(http.MethodGet, pendingUrl, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}

	req2.Header.Set("Authorization", "Bearer "+token)

	req2u, err := http.NewRequest(http.MethodGet, pendingUserUrl, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}

	req2u.Header.Set("Authorization", "Bearer "+token)

	url := "http://localhost:3000/api/admin/loans/approve/" + strconv.Itoa(loanOne)
	req3, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	req3.Header.Set("Authorization", "Bearer "+token)
	requests := []*http.Request{req1, req2, req2u, req3}
	statusOutArr := []int{http.StatusOK, http.StatusUnauthorized, http.StatusOK, http.StatusUnauthorized}
	getMultiPleResponse(t, requests, statusOutArr)
}

func TestAcceptLoan(t *testing.T) {
	url := "http://localhost:3000/api/admin/loans/approve/" + strconv.Itoa(loanOne)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Error(err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)

	getSingleOKResponse(t, req)
}

func TestRejectLoan(t *testing.T) {
	url := "http://localhost:3000/api/admin/loans/decline/" + strconv.Itoa(loanTwo)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Error(err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)

	getSingleOKResponse(t, req)
}

func TestReturnedLoan(t *testing.T) {
	url := "http://localhost:3000/api/admin/loans/returned/" + strconv.Itoa(loanOne)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Error(err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)

	getSingleOKResponse(t, req)
}

func TestGetBook(t *testing.T) {
	url := "http://localhost:3000/api/book"
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%d", url, bookID), nil)
	if err != nil {
		t.Error(err.Error())
	}

	req.Header.Set("Authorization", "Bearer "+adminToken)

	getSingleOKResponse(t, req)
}

func TestGetAllBooks(t *testing.T) {
	req1, err := http.NewRequest(http.MethodGet, "http://localhost:3000/api/books", nil)
	if err != nil {
		t.Error(err.Error())
	}

	req1.Header.Set("Authorization", "Bearer "+adminToken)

	req2, err := http.NewRequest(http.MethodGet, "http://localhost:3000/api/books", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}

	req2.Header.Set("Authorization", "Bearer "+token)

	req3, err := http.NewRequest(http.MethodGet, "http://localhost:3000/api/book", nil)
	if err != nil {
		t.Error(err.Error())
		return
	}
	req3.Header.Set("Authorization", "Bearer "+token)
	requests := []*http.Request{req1, req2, req3}
	statusOutArr := []int{http.StatusOK, http.StatusOK, http.StatusNotFound}
	getMultiPleResponse(t, requests, statusOutArr)
}

func TestDeleteBook(t *testing.T) {
	url := fmt.Sprintf("http://localhost:3000/api/admin/book/delete/%d", bookID)
	req1, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Error(err.Error())
	}
	req1.Header.Set("Authorization", "Bearer "+token)

	req2, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Error(err.Error())
	}
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	requests := []*http.Request{req1, req2}
	statusOutArr := []int{http.StatusUnauthorized, http.StatusOK}
	getMultiPleResponse(t, requests, statusOutArr)
}

func TestDeleteAuthor(t *testing.T) {
	url := fmt.Sprintf("http://localhost:3000/api/admin/author/delete/%d", bookID)
	req1, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Error(err.Error())
	}
	req1.Header.Set("Authorization", "Bearer "+token)

	req2, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Error(err.Error())
	}
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	requests := []*http.Request{req1, req2}
	statusOutArr := []int{http.StatusUnauthorized, http.StatusOK}
	getMultiPleResponse(t, requests, statusOutArr)
}

func TestDBDeleteUser(t *testing.T) {
	url := fmt.Sprintf("http://localhost:3000/api/admin/author/delete/%d", bookID)
	req1, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Error(err.Error())
	}
	req1.Header.Set("Authorization", "Bearer "+token)

	req2, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		t.Error(err.Error())
	}
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	requests := []*http.Request{req1, req2}
	statusOutArr := []int{http.StatusUnauthorized, http.StatusOK}
	getMultiPleResponse(t, requests, statusOutArr)
}

func TestRedis(t *testing.T) {
	db.InitRedis()
	redis := db.GetClient()
	result, err := redis.Ping().Result()
	log.Println(result)
	if err == nil && result != strings.ToUpper("pong") {
		t.Error("FAIL")
	}
	err = db.RemoveByKey(routes.UserPrefix + username)
	if err != nil {
		t.Error("FAIL")
	}
}

func getSingleOKResponse(t *testing.T, req *http.Request) string {

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Println("err:", err.Error())
	}

	//Check the status code is what we expect.
	if http.StatusOK != res.StatusCode {
		t.Error("Handler returned wrong status code: got ", res.StatusCode, "want ", http.StatusOK)
	}

	bytes := make([]byte, res.ContentLength)
	res.Body.Read(bytes)

	return strings.ReplaceAll(string(bytes), "\"", "")
}

func getMultiPleResponse(t *testing.T, requests []*http.Request, statusOutArr []int) {

	for i, req := range requests {
		//Now the response request pair is served via http
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Println("err:", err.Error())
		}

		//Check the status code is what we expect.
		if statusOutArr[i] != res.StatusCode {
			t.Error(i, "Handler returned wrong status code: got ", res.StatusCode, "want ", statusOutArr[i])
		}
	}

}
