package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
)

// @Summary		Delete User
// @Description	Delete a user by ID
// @Tags			user
// @Produce		json
// @Param			id	path	string	true	"User ID"
// @Success		204	{string}	string	""
// @Failure		400	{object}	string
// @Failure		404	{object}	string
// @Failure		500	{object}	string
// @Router			/user/{id} [DELETE]
func deleteUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Read id from path parameters
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to parse id from url",
				slog.String("id", idStr),
				slog.String("error", err.Error()),
			)
			encodeErr := encodeResponse(w, http.StatusBadRequest, ProblemDetail{
				Title:  "Invalid ID",
				Status: http.StatusBadRequest,
				Detail: "The provided ID is not a valid integer.",
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

		// Delete user from db
		logger.DebugContext(ctx, "Deleting user", "id", id)

		result, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
		if err != nil {
			logger.ErrorContext(ctx, "failed to delete user", slog.String("error", err.Error()))
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

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to get rows affected",
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

		if rowsAffected == 0 {
			encodeErr := encodeResponse(w, http.StatusNotFound, ProblemDetail{
				Title:  "User Not Found",
				Status: http.StatusNotFound,
				Detail: fmt.Sprintf("User with ID %d not found", id),
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

		// Respond with no content
		w.WriteHeader(http.StatusNoContent)
	}
}
