package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"example.com/examples/api/layered/internal/models"
	"go.opentelemetry.io/otel"
)

const name = "example.com/examples/api/layered/internal/handlers"

var tracer = otel.Tracer(name)

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
		_, span := tracer.Start(r.Context(), "createUserHandler")
		defer span.End()
		ctx := r.Context()

		// Request validation
		request, problems, err := decodeValid[UserRequest](r)
		if err != nil && len(problems) == 0 {
			logger.ErrorContext(
				ctx,
				"failed to decode request",
				slog.String("error", err.Error()),
			)

			encodeErr := encodeResponse(w, http.StatusInternalServerError, ProblemDetail{
				Title:  "Bad Request",
				Status: 400,
				Detail: "Invalid request body.",
			})
			if encodeErr != nil {
				logger.ErrorContext(
					ctx,
					"failed to encode response",
					slog.String("error", encodeErr.Error()),
				)
			}

			return
		}
		if len(problems) > 0 {
			logger.ErrorContext(
				ctx,
				"Validation error",
				slog.String("Validation errors: ", fmt.Sprintf("%#v", problems)),
			)

			encodeErr := encodeResponse(w, http.StatusBadRequest, NewValidationBadRequest(problems))
			if encodeErr != nil {
				logger.ErrorContext(
					ctx,
					"failed to encode response",
					slog.String("error", encodeErr.Error()),
				)
			}

			return
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

			encodeErr := encodeResponse(w, http.StatusInternalServerError, NewInternalServerError())
			if encodeErr != nil {
				logger.ErrorContext(
					ctx,
					"failed to encode response",
					slog.String("error", encodeErr.Error()),
				)
			}
			return
		}

		// Convert our models.User domain model into a response model.
		encodeErr := encodeResponse(w, http.StatusCreated, UserResponse{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			Password: user.Password,
		})
		if encodeErr != nil {
			logger.ErrorContext(
				ctx,
				"failed to encode response",
				slog.String("error", encodeErr.Error()),
			)
		}
	}
}
