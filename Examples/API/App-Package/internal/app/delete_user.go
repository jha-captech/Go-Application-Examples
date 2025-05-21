package app

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
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
func deleteUser(logger *slog.Logger, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Read id from path parameters
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.ErrorContext(ctx, "failed to parse id from url", slog.String("id", idStr), slog.String("error", err.Error()))
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		// Delete user from db
		logger.DebugContext(ctx, "Deleting user", "id", id)

		result, err := db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
		if err != nil {
			logger.ErrorContext(ctx, "failed to delete user", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			logger.ErrorContext(ctx, "failed to get rows affected", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if rowsAffected == 0 {
			http.Error(w, "User Not Found", http.StatusNotFound)
			return
		}

		// Respond with no content
		w.WriteHeader(http.StatusNoContent)
	})
}
