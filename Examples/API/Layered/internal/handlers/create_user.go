package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/codes"

	"example.com/examples/api/layered/internal/middleware"
	"example.com/examples/api/layered/internal/models"
)

// userCreator represents a type capable of reading a user from storage and
// returning it or an error.
type userCreator interface {
	CreateUser(ctx context.Context, user models.User) (models.User, error)
}

// HandleCreateUser handles the creation of a new user.
//
// @Summary		Create User
// @Description	Creates a User
// @Tags		user
// @Accept		json
// @Produce		json
// @Param		request	body		UserRequest	true	"User to Create"
// @Success		200		{object}	uint
// @Failure		400		{object}	string
// @Failure		404		{object}	string
// @Failure		500		{object}	string
// @Router		/user  [POST]
func HandleCreateUser(logger *slog.Logger, userCreator userCreator) http.HandlerFunc {
	const name = "handlers.HandleCreateUser"
	logger = logger.With(slog.String("func", name))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), name)
		defer span.End()

		// Request validation
		request, problems, err := decodeValid[UserRequest](r)
		if err != nil && len(problems) == 0 {
			logger.ErrorContext(
				ctx,
				"failed to decode request",
				slog.String("error", err.Error()),
			)
			// otel set error info
			span.SetStatus(codes.Error, "decoding request failed")
			span.RecordError(err)

			_ = encodeResponseJSON(
				w, http.StatusInternalServerError, ProblemDetail{
					Title:   "Bad Request",
					Status:  400,
					Detail:  "Invalid request body.",
					TraceID: middleware.GetTraceID(ctx),
				},
			)

			return
		}
		if len(problems) > 0 {
			validationError := "validation failed"
			logger.ErrorContext(
				ctx,
				validationError,
				slog.Any("validation_errors", problems),
			)
			span.SetStatus(codes.Error, validationError)
			span.RecordError(errors.New(validationError))

			_ = encodeResponseJSON(w, http.StatusBadRequest, NewValidationBadRequest(ctx, problems))

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
			span.SetStatus(codes.Error, "creating user failed")
			span.RecordError(err)

			_ = encodeResponseJSON(w, http.StatusInternalServerError, NewInternalServerError(ctx))

			return
		}

		// Convert our models.User domain model into a response model.
		_ = encodeResponseJSON(
			w, http.StatusCreated, UserResponse{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
			},
		)
	}
}
