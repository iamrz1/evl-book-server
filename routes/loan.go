package routes

import (
	"encoding/json"
	"errors"
	"evl-book-server/auth"
	"evl-book-server/config"
	"evl-book-server/db"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// CreateLoanRequestHandler lets a logged in user to
// request for a book using that books ID.
func CreateLoanRequestHandler(w http.ResponseWriter, r *http.Request) {
	loan := getLoanDetails(r)
	// check for inconsistencies
	validLoan, err := ValidateLoanCreate(loan)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	loanKey := LoanPrefix + strconv.Itoa(loan.ID)
	loanBytes, err := json.Marshal(validLoan)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = db.SetJsonValues(loanKey, loanBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Add loan to user's loanArray
	err = addLoanIDToUsersLoanIDArray(r.Header.Get(auth.UsernameKey), loan.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("loan request created"))
}

func addLoanIDToUsersLoanIDArray(username string, loanID int) error {
	userKey := strings.ToLower(UserPrefix + username)

	user, err := getUserByKey(userKey)
	if err != nil {
		return err
	}

	user.LoanIDArray = append(user.LoanIDArray, loanID)

	updatedUserBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}

	_ = db.SetJsonValues(userKey, updatedUserBytes)

	return nil
}

// ApproveLoanRequestHandler is used by the admin to
// approve a loan request by loan ID
func ApproveLoanRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	loanKey := LoanPrefix + vars["id"]
	loan := getLoanByKeyFromDB(loanKey)
	if loan.Approved == true {
		w.WriteHeader(http.StatusForbidden)
		http.Error(w, "loan has been approved already", http.StatusForbidden)
		return
	}
	//add approved flag
	loan.Approved = true

	//increment onloan in books by one
	err := updateBookLoanCountByID(loan.BookID, 1)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		if err.Error() == db.RedisNilErr {
			http.Error(w, "loan doesn't exist", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	//now update the load and finish the approval process
	loanBytes, err := json.Marshal(loan)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = db.SetJsonValues(loanKey, loanBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("loan approved successfully"))
}

// DeclineLoanRequestHandler declines loan. If it is a pending loan,
// It removes loan request from database and remove it;s id from user's pending list
func DeclineLoanRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loanKey := LoanPrefix + vars["id"]
	loan := getLoanByKeyFromDB(loanKey)
	if loan.Approved == true {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("can not decline request, loan has already been approved"))
		return
	}

	loanID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("loan id is must be an integer"))
		return
	}
	//remove loan from user's end
	err = removeLoanIDFromUser(UserPrefix+loan.Username, loanID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("loan id could not be removed from user's loan array"))
		return
	}

	//now delete the loan and finish the decline process
	err = db.RemoveByKey(loanKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("loan declined successfully"))
}

// ReturnedBookHandler takes loaned item back. If it is an approved loan,
// It removes loan request from database and remove it's id from user's pending list
func ReturnedBookHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loanKey := LoanPrefix + vars["id"]
	loan := getLoanByKeyFromDB(loanKey)

	if loan.Approved == false {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("can not accept return request, loan has not been approved yet"))
		return
	}

	loanID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("loan id is must be an integer"))
		return
	}
	//remove loan from user's end
	err = removeLoanIDFromUser(UserPrefix+loan.Username, loanID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("loan id could not be removed from user's loan array"))
		return
	}

	// decrement onloan count in book by one
	err = updateBookLoanCountByID(loan.BookID, -1)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	//now delete the loan and finish the decline process
	err = db.RemoveByKey(loanKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte("return confirmed successfully"))
}

// GetLoanByIDForThisUserHandler returns a loan by loanID
// if the loan belongs to this user
func GetLoanByIDForThisUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loanID := vars["id"]

	loanKey := LoanPrefix + loanID

	userKey := UserPrefix + strings.ToLower(r.Header.Get(auth.UsernameKey))

	user, err := getUserByKey(userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	loanIDFound := false
	for _, userLoanID := range user.LoanIDArray {
		if strconv.Itoa(userLoanID) == loanID {
			loanIDFound = true
			break
		}
	}
	if !loanIDFound {
		w.WriteHeader(http.StatusNoContent)
		_, _ = w.Write([]byte("you dont have any loan by this id"))
		return
	}

	loanByte, err := db.GetByteValues(loanKey)
	if err == nil {
		_, _ = w.Write(loanByte)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	http.Error(w, err.Error(), http.StatusInternalServerError)
	return
}

// GetAllLoansForThisUserHandler returns all loans
// that belongs to this user
func GetAllLoansForThisUserHandler(w http.ResponseWriter, r *http.Request) {
	userKey := UserPrefix + strings.ToLower(r.Header.Get(auth.UsernameKey))

	user, err := getUserByKey(userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resultString := "["
	for _, loanID := range user.LoanIDArray {
		loanKey := LoanPrefix + strconv.Itoa(loanID)
		loan, err := db.GetSingleValue(loanKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if resultString != "[" {
			resultString += ","
		}
		resultString += loan
	}
	resultString += "]"
	if resultString == "[]" {
		_, _ = w.Write([]byte("no active or pending loans"))
		return
	}
	_, _ = w.Write([]byte(resultString))
}

// GetAllPendingLoansForThisUserHandler returns all pending loans
// that belongs to this user
func GetAllPendingLoansForThisUserHandler(w http.ResponseWriter, r *http.Request) {
	userKey := UserPrefix + strings.ToLower(r.Header.Get(auth.UsernameKey))

	user, err := getUserByKey(userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resultString := "["
	for _, loanID := range user.LoanIDArray {
		loanKey := LoanPrefix + strconv.Itoa(loanID)
		loan := getLoanByKeyFromDB(loanKey)
		if loan.Approved == true {
			continue
		}
		// else if loan is not approved, add to the string of pending loans
		loanString, err := db.GetSingleValue(loanKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if resultString != "[" {
			resultString += ","
		}
		resultString += loanString
	}
	resultString += "]"
	if resultString == "[]" {
		_, _ = w.Write([]byte("no pending loan requests"))
		return
	}
	_, _ = w.Write([]byte(resultString))
}

// GetAllActiveLoansForThisUserHandler returns all
//approved loans that belongs to this user
func GetAllActiveLoansForThisUserHandler(w http.ResponseWriter, r *http.Request) {
	userKey := UserPrefix + strings.ToLower(r.Header.Get(auth.UsernameKey))

	user, err := getUserByKey(userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resultString := "["
	for _, loanID := range user.LoanIDArray {
		loanKey := LoanPrefix + strconv.Itoa(loanID)
		loan := getLoanByKeyFromDB(loanKey)
		if loan.Approved == false {
			continue
		}
		// else if loan is approved, add to the string of approved loans
		loanString, err := db.GetSingleValue(loanKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if resultString != "[" {
			resultString += ","
		}
		resultString += loanString
	}
	resultString += "]"
	if resultString == "[]" {
		_, _ = w.Write([]byte("no active loans"))
		return
	}
	_, _ = w.Write([]byte(resultString))
}

// GetLoanByIDHandler returns a loan by loanID for admin
func GetLoanByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	loanID := vars["id"]

	loanKey := LoanPrefix + loanID

	loanByte, err := db.GetByteValues(loanKey)
	if err == nil {
		_, _ = w.Write(loanByte)
		return
	} else {
		if err.Error() == db.RedisNilErr {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("no loan by this id"))
			return
		}
		// if any other error
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetAllLoansHandler returns to admin a list of all loans
func GetAllLoansHandler(w http.ResponseWriter, _ *http.Request) {
	loanKeys, err := db.ScanKeysByPrefix(LoanPrefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(loanKeys) == 0 {
		_, _ = w.Write([]byte("you dont have any loans"))
		return
	}
	resultString := "["
	for _, loanKey := range loanKeys {
		loan, err := db.GetSingleValue(loanKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if resultString != "[" {
			resultString += ","
		}
		resultString += loan
	}
	resultString += "]"
	if resultString == "[]" {
		_, _ = w.Write([]byte("no active loans or pending loan requests"))
		return
	}
	_, _ = w.Write([]byte(resultString))
}

// GetAllPendingLoansHandler returns to admin a list of all pending loans
func GetAllPendingLoansHandler(w http.ResponseWriter, _ *http.Request) {
	loanKeys, err := db.ScanKeysByPrefix(LoanPrefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(loanKeys) == 0 {
		_, _ = w.Write([]byte("you dont have any loans"))
		return
	}
	resultString := "["
	for _, loanKey := range loanKeys {
		loan := getLoanByKeyFromDB(loanKey)
		if loan.Approved == true {
			continue
		}

		loanString, err := db.GetSingleValue(loanKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if resultString != "[" {
			resultString += ","
		}
		resultString += loanString
	}
	resultString += "]"
	if resultString == "[]" {
		_, _ = w.Write([]byte("no pending loan requests"))
		return
	}
	_, _ = w.Write([]byte(resultString))
}

// GetAllActiveLoansHandler returns to admin a list of all active loans
func GetAllActiveLoansHandler(w http.ResponseWriter, _ *http.Request) {
	loanKeys, err := db.ScanKeysByPrefix(LoanPrefix)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(loanKeys) == 0 {
		_, _ = w.Write([]byte("you dont have any loans"))
		return
	}
	resultString := "["
	for _, loanKey := range loanKeys {
		loan := getLoanByKeyFromDB(loanKey)
		if loan.Approved == false {
			continue
		}

		loanString, err := db.GetSingleValue(loanKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if resultString != "[" {
			resultString += ","
		}
		resultString += loanString
	}
	resultString += "]"
	if resultString == "[]" {
		_, _ = w.Write([]byte("no active loans"))
		return
	}
	_, _ = w.Write([]byte(resultString))
}

// extract info about loan from the incoming request
func getLoanDetails(r *http.Request) config.Loan {
	vars := mux.Vars(r)
	bookID, err := strconv.Atoi(vars["book_id"])
	if err != nil {
		// If there is something wrong with the request body
		return config.Loan{}
	}
	log.Println("bookID=", bookID)
	loan := config.Loan{}
	loan.BookID = bookID
	loanID := 0
	for n := 1; loanID == 0; n++ {
		_, err := db.GetByteValues(LoanPrefix + strconv.Itoa(n))
		if err != nil && err.Error() == db.RedisNilErr {
			loanID = n
		}
	}
	log.Println("loanID=", loanID)
	loan.ID = loanID
	log.Println("username=", r.Header.Get(auth.UsernameKey))
	loan.Username = r.Header.Get(auth.UsernameKey)
	loan.Approved = false

	return loan
}

func getLoanByKeyFromDB(loanKey string) config.Loan {

	loanByte, err := db.GetByteValues(loanKey)
	if err != nil {
		return config.Loan{}
	}
	loan := config.Loan{}

	if err := json.Unmarshal(loanByte, &loan); err != nil {
		return config.Loan{}
	}

	return loan
}

func ValidateLoanCreate(loan config.Loan) (config.Loan, error) {
	if loan.ID == 0 || loan.Username == "" || loan.BookID == 0 {
		return config.Loan{}, errors.New("loanID, username, or BookID is missing")
	}
	return loan, nil
}

func removeLoanIDFromUser(userKey string, loanID int) error {
	user, err := getUserByKey(userKey)
	if err != nil {
		return err
	}

	// remove loan from loan array
	newLoanIDArray := RemoveElementFromArray(user.LoanIDArray, loanID)
	// replace existing array with this new one
	user.LoanIDArray = newLoanIDArray
	//Save the new information in db
	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}
	err = db.SetJsonValues(userKey, userBytes)
	if err != nil {
		return err
	}

	return nil
}

// A helper method that removes a single element from an array
// and returns the modified error
func RemoveElementFromArray(sourceArray []int, element int) []int {

	for i, value := range sourceArray {
		if value == element {
			//sourceArray = append(sourceArray[:i], sourceArray[i+1:]...)
			sourceArray[i] = sourceArray[len(sourceArray)-1]
			return sourceArray[:len(sourceArray)-1]
		}
	}
	return sourceArray
}

func updateBookLoanCountByID(bookID int, count int) error {
	bookKey := BookPrefix + strconv.Itoa(bookID)
	savedBook := config.Book{}
	savedBookByte, err := db.GetByteValues(bookKey)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(savedBookByte, &savedBook); err != nil {
		log.Println("unmarshal error :", err.Error())
		return err
	}
	if savedBook.OnLoanCount+count <= savedBook.TotalCount && savedBook.OnLoanCount+count >= 0 {
		savedBook.OnLoanCount += count
		newBookByte, err := json.Marshal(savedBook)
		if err != nil {
			return err
		}
		_ = db.SetJsonValues(bookKey, newBookByte)
		return nil
	}
	return errors.New("can not loan this book at the moment")
}
