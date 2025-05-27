package handlers

// UserRequest represents the request for creating a user.
type UserRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=50"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=30"`
}
