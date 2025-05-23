package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
)

// @Summary		Update User
// @Description	Update user fields by ID
// @Tags			user
// @Accept			json
// @Produce		json
// @Param			id		path	string	true	"User ID"
// @Param			user	body	User	true	"User data"
// @Success		200		{object}	User
// @Failure		400		{object}	string
// @Failure		404		{object}	string
// @Failure		500		{object}	string
// @Router			/user/{id} [PUT]
func updateUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
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
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to read request body",
				slog.String("error", err.Error()),
			)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var user User
		if err := json.Unmarshal(body, &user); err != nil {
			logger.ErrorContext(ctx, "failed to unmarshal user", slog.String("error", err.Error()))
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		logger.InfoContext(ctx, "Updating user",
			slog.Int("id", id),
			slog.String("name", user.Name),
			slog.String("email", user.Email),
		)

		// Update user in db
		query := `
            UPDATE users
            SET name = $1, email = $2, password = $3
            WHERE id = $4
            RETURNING id, name, email, password
        `
		err = db.GetContext(ctx, &user, query, user.Name, user.Email, user.Password, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "User Not Found", http.StatusNotFound)
				return
			}
			logger.ErrorContext(ctx, "failed to update user", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		logger.InfoContext(ctx, "User updated successfully",
			slog.Uint64("id", uint64(user.ID)),
			slog.String("name", user.Name),
			slog.String("email", user.Email),
		)

		// Respond with updated user
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			logger.ErrorContext(ctx, "failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
