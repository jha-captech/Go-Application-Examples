package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/examples/api/layered/internal/models"
)

// userReader represents a type capable of reading a user from storage and
// returning it or an error.
type userReader interface {
	ReadUser(ctx context.Context, id uint64) (models.User, error)
}

// @Summary		Read User
// @Description	Read User by ID
// @Tags			user
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"User ID"
// @Success		200	{object}	models.User
// @Failure		400	{object}	string
// @Failure		404	{object}	string
// @Failure		500	{object}	string
// @Router			/user/{id}  [GET]
func HandleReadUser(logger *slog.Logger, userReader userReader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

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

			encodeResponse(w, logger, http.StatusBadRequest, "Invalid ID")
			return
		}

		// Read the user
		user, err := userReader.ReadUser(ctx, uint64(id))
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to read user",
				slog.String("error", err.Error()),
			)

			encodeResponse(w, logger, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		// Encode the response model as JSON
		encodeResponse(w, logger, http.StatusOK, UserResponse{
			ID:       user.ID,
			Name:     user.Name,
			Email:    user.Email,
			Password: user.Password,
		})
	}
}
