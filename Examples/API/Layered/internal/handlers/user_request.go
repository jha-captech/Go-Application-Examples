package handlers

import (
	"context"
	"net/mail"
)

// createUserRequest represents the request for creating a user.
type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *UserRequest) Valid(ctx context.Context) map[string]string {
	problems := make(map[string]string)

	if r.Name == "" {
		problems["Name"] = "Name cannot be empty"
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		problems["Email"] = "Email is not in correct format"
	}
	if len(r.Password) <= 8 {
		problems["Password"] = "Password length must be greater than 8 characters "
	}

	return problems
}
