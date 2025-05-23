package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
)

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
func readUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
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

			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		// Read the user
		logger.InfoContext(ctx, "Reading user", slog.Int("id", id))

		var user User
		err = db.GetContext(
			ctx,
			&user,
			`
			SELECT id,
				name,
				email,
				password
			FROM users
			WHERE id = $1::int
			`,
			id,
		)
		
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				http.Error(w, "User Not Found", http.StatusNotFound)
				return
			default:
				logger.ErrorContext(
					ctx,
					"failed to read user",
					slog.String("error", err.Error()),
				)

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		// Encode the user model as JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			logger.ErrorContext(
				ctx,
				"failed to encode response",
				slog.String("error", err.Error()))

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
