package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/examples/api/layered/internal/models"
)

// userUpdater represents a type capable of updating a user and
// returning it or an error.
type userUpdater interface {
	UpdateUser(ctx context.Context, id uint64, patch models.User) (models.User, error)
}

// @Summary		Update User
// @Description	Update User by ID
// @Tags			user
// @Accept			json
// @Produce		json
// @Param			id		path		string		true	"User ID"
// @Param			request	body		UserRequest	true	"User to Create"
// @Success		200		{object}	models.User
// @Failure		400		{object}	string
// @Failure		404		{object}	string
// @Failure		500		{object}	string
// @Router			/user/{id}  [PUT]
func HandleUpdateUser(logger *slog.Logger, userUpdater userUpdater) http.HandlerFunc {
	const name = "handlers.HandleUpdateUser"
	logger = logger.With(slog.String("func", name))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), name)
		defer span.End()

		// Read id from path parameters
		idStr := r.PathValue("id")

		// Convert the ID from string to int
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to parse id from url",
				slog.String("id", idStr),
				slog.String("error", err.Error()),
			)

			_ = encodeResponse(w, http.StatusBadRequest, ProblemDetail{
				Title:  "Invalid ID",
				Status: http.StatusBadRequest,
				Detail: "The provided ID is not a valid integer.",
			})

			return
		}

		// Request validation
		request, problems, err := decodeValid[UserRequest](r)
		if err != nil && len(problems) == 0 {
			logger.ErrorContext(
				ctx,
				"failed to decode request",
				slog.String("error", err.Error()))

			_ = encodeResponse(w, http.StatusInternalServerError, NewInternalServerError())

			return
		}
		if len(problems) > 0 {
			logger.ErrorContext(
				ctx,
				"Validation error",
				slog.String("Validation errors: ", fmt.Sprintf("%#v", problems)),
			)

			NewValidationBadRequest(problems)
		}

		modelRequest := models.User{
			Name:     request.Name,
			Email:    request.Email,
			Password: request.Password,
		}

		// Update the user
		user, err := userUpdater.UpdateUser(ctx, uint64(id), modelRequest)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to update user",
				slog.String("error", err.Error()),
			)

			_ = encodeResponse(w, http.StatusInternalServerError, NewInternalServerError())

			return
		}

		// Encode the response model as JSON
		_ = encodeResponse(w, http.StatusOK, UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		})
	}
}
