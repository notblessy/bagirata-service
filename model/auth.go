package model

// LoginRequest is the body for POST /login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is the body for POST /register
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// UserResponse is the user object in auth responses (no password)
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

// AuthResponse is the response for login and register
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
