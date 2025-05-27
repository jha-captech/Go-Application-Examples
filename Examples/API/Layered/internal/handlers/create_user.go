package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"example.com/examples/api/layered/internal/models"
)

// userCreator represents a type capable of reading a user from storage and
// returning it or an error.
type userCreator interface {
	CreateUser(ctx context.Context, user models.User) (models.User, error)
}

// @Summary		Create User
// @Description	Creates a User
// @Tags			user
// @Accept			json
// @Produce		json
// @Param			request	body		UserRequest	true	"User to Create"
// @Success		200		{object}	uint
// @Failure		400		{object}	string
// @Failure		404		{object}	string
// @Failure		500		{object}	string
// @Router			/user  [POST]
func HandleCreateUser(logger *slog.Logger, userCreator userCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Request validation
		request, problems, err := decodeValid[UserRequest](r)
		if err != nil && len(problems) == 0 {
			logger.ErrorContext(
				ctx,
				"failed to decode request",
				slog.String("error", err.Error()),
			)

			// replace with json decoder helper with error from decoder
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		if len(problems) > 0 {
			logger.ErrorContext(
				ctx,
				"Validation error",
				slog.String("Validation errors: ", fmt.Sprintf("%#v", problems)),
			)
		}

		modelRequest := models.User{
			Name:     request.Name,
			Email:    request.Email,
			Password: request.Password,
		}

		// Read the user
		user, err := userCreator.CreateUser(ctx, modelRequest)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to create user",
				slog.String("error", err.Error()),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Convert our models.User domain model into a response model.
		encodeResponse(w, logger, http.StatusOK, UserResponse{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			Password: user.Password,
		})
	}
}
