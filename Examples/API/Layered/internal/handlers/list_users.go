package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"example.com/examples/api/layered/models"
)

// userReader represents a type capable of reading a user from storage and
// returning it or an error.
type usersLister interface {
	ListUsers(ctx context.Context, name string) ([]models.User, error)
}

// listUsersResponse represents the response for listing users.
type listUsersResponse struct {
	Users []UserResponse
}

// @Summary		List Users
// @Description	List All Users
// @Tags			user
// @Accept			json
// @Produce		json
// @Param			name	query		string	false	"query param"
// @Success		200		{array}		models.User
// @Failure		400		{object}	string
// @Failure		404		{object}	string
// @Failure		500		{object}	string
// @Router			/user  [GET]
func HandleListUsers(logger *slog.Logger, usersLister usersLister) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		name := r.URL.Query().Get("name")

		// Read the user
		users, err := usersLister.ListUsers(ctx, name)
		if err != nil {
			logger.ErrorContext(
				r.Context(),
				"failed to list users",
				slog.String("error", err.Error()),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Convert our models.User domain model into a response model.
		response := listUsersResponse{
			Users: []UserResponse{},
		}

		for _, user := range users {
			newUser := UserResponse{
				ID:       user.ID,
				Name:     user.Name,
				Email:    user.Email,
				Password: user.Password,
			}
			response.Users = append(response.Users, newUser)
		}

		// Encode the response model as JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.ErrorContext(
				r.Context(),
				"failed to encode response",
				slog.String("error", err.Error()))

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}
