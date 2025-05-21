package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
)

// uerDeleter represents a type capable of deleting a user from storage
type userDeleter interface {
	DeleteUser(ctx context.Context, id uint64) error
}

// @Summary		Delete User
// @Description	Delete User by ID
// @Tags			user
// @Accept			json
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Success		200
// @Failure		400	{object}	string
// @Failure		404	{object}	string
// @Failure		500	{object}	string
// @Router			/user/{id}  [DELETE]
func HandleDeleteUser(logger *slog.Logger, userDeleter userDeleter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Read id from path parameters
		idStr := r.PathValue("id")

		// Convert the ID from string to int
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.ErrorContext(
				r.Context(),
				"failed to parse id from url",
				slog.String("id", idStr),
				slog.String("error", err.Error()),
			)

			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		// Delete the user
		err = userDeleter.DeleteUser(ctx, uint64(id))
		if err != nil {
			logger.ErrorContext(
				r.Context(),
				"failed to read user",
				slog.String("error", err.Error()),
			)

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		// Encode the response model as JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
}
