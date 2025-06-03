package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/codes"

	"example.com/examples/api/layered/internal/models"
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

// HandleListUsers handles the listing of all users.
//
// @Summary		List Users
// @Description	List All Users
// @Tags		user
// @Accept		json
// @Produce		json
// @Param		name	query		string	false	"query param"
// @Success		200		{array}		models.User
// @Failure		400		{object}	string
// @Failure		404		{object}	string
// @Failure		500		{object}	string
// @Router		/user  [GET]
func HandleListUsers(logger *slog.Logger, usersLister usersLister) http.HandlerFunc {
	const name = "handlers.HandleListUsers"
	logger = logger.With(slog.String("func", name))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), name)
		defer span.End()

		name := r.URL.Query().Get("name")

		// Read the user
		users, err := usersLister.ListUsers(ctx, name)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to list users",
				slog.String("error", err.Error()),
			)
			span.SetStatus(codes.Error, "listing users failed")
			span.RecordError(err)

			_ = encodeResponseJSON(w, http.StatusInternalServerError, NewInternalServerError(ctx))

			return
		}

		// Convert our models.User domain model into a response model.
		response := listUsersResponse{
			Users: []UserResponse{},
		}

		for _, user := range users {
			newUser := UserResponse{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
			}
			response.Users = append(response.Users, newUser)
		}

		// Encode the response model as JSON
		_ = encodeResponseJSON(w, http.StatusOK, response)
	}
}
