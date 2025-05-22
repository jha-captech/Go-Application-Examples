package app

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
)

// @Summary		List Users
// @Description	List all users
// @Tags			user
// @Produce		json
// @Success		200	{array}	User
// @Failure		500	{object}	string
// @Router			/user [GET]
func listUsers(logger *slog.Logger, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger.DebugContext(ctx, "Listing all users")

		rows, err := db.QueryContext(ctx, `
            SELECT id, name, email, password
            FROM users
        `)
		if err != nil {
			logger.ErrorContext(ctx, "failed to query users", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []User
		for rows.Next() {
			var user User
			if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password); err != nil {
				logger.ErrorContext(ctx, "failed to scan user", slog.String("error", err.Error()))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			users = append(users, user)
		}

		if err := rows.Err(); err != nil {
			logger.ErrorContext(ctx, "row iteration error", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(users); err != nil {
			logger.ErrorContext(ctx, "failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}
