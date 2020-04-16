package config

type Book struct {
	ID          int    `json:"book_id"`
	BookName    string `json:"book_name"`
	AuthorID    int    `json:"author_id"`
	AddToCount  int    `json:"add_count"`
	TotalCount  int
	OnLoanCount int
}

type Author struct {
	ID         int    `json:"author_id"`
	AuthorName string `json:"author_name"`
	//AuthorName string   `json:"author_name"`
	AuthoredBooks []AuthoredBookUnit
}

type AuthoredBookUnit struct {
	BookID int `json:"book_id"`
}

type BookLoan struct {
	LoanRequests []BookLoanUnit `json:"loan_requests"`
}

type BookLoanUnit struct {
	BookID   int    `json:"book_id"`
	username string `json:"username"`
}
