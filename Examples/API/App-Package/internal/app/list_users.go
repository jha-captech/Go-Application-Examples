package app

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// @Summary		List Users
// @Description	List all users
// @Tags			user
// @Produce		json
// @Success		200	{array}	User
// @Failure		500	{object}	string
// @Router			/user [GET]
func listUsers(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger.InfoContext(ctx, "Listing all users")

		var users []User
		err := db.SelectContext(
			ctx,
			&users,
			`
            SELECT id, name, email, password
            FROM users
            `,
		)

		if err != nil {
			logger.ErrorContext(ctx, "failed to query users", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(users); err != nil {
			logger.ErrorContext(ctx, "failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
