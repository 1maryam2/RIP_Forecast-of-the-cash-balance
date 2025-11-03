package ds

// TokenPair содержит пару access и refresh токенов
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// LoginResponse ответ на авторизацию
type LoginResponse struct {
	TokenPair
	User *Users `json:"user"`
}
