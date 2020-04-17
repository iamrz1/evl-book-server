package config

type UserCredentials struct {
	Username    string   `json:"username"`
	Name        string   `json:"name"`
	Password    string   `json:"password"`
	UserData    UserData `json:"user_data"`
	LoanIDArray []int
}

type UserData struct {
	ProfilePicURL string
	IsAdmin       bool
}

type UserLoanData struct {
	LoanedBooks []LoanedBookUnit `json:"loaned_books"`
}

type LoanedBookUnit struct {
	BookID int `json:"book_id"`
	//AuthorID int `json:"author_id"`
}
