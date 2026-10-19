package http

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=25"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=32"`
}

type LoginRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8,max=32"`
	ClientIP  string `json:"client_ip" binding:"required"`
	UserAgent string `json:"user_agent" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string
	ClientIP     string
	UserAgent    string
}
