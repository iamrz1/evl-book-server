package config

type Book struct {
	ID          int    `json:"book_id"`
	BookName    string `json:"book_name"`
	AuthorID    int    `json:"author_id"`
	AddCount    int    `json:"add_count"`
	TotalCount  int
	OnLoanCount int
}

type Author struct {
	ID              int    `json:"author_id"`
	AuthorName      string `json:"author_name"`
	AuthoredBookIDs []int
}

type Loan struct {
	ID       int
	BookID   int
	Username string
	Approved bool
}
