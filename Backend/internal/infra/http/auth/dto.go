package auth

//easyjson:json
type RegRequestDto struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResp struct {
	Token string `json:"token"`
}
