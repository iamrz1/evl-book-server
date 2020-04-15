package config

type UserCredentials struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	UserData UserData `json:"user_data"`
}

type UserData struct {
	Name          string `json:"name"`
	ProfilePicURL string `json:profile_picture_url`
	IsAdmin       bool   `json:"admin"`
}
