package app

import (
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// @Summary		Create User
// @Description	Create a new user
// @Tags			user
// @Accept			json
// @Produce		json
// @Param			user	body		User	true	"User data"
// @Success		201		{object}	User
// @Failure		400		{object}	string
// @Failure		500		{object}	string
// @Router			/user [POST]
func createUser(logger *slog.Logger, db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Read request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.ErrorContext(ctx, "failed to read request body", slog.String("error", err.Error()))
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

		// Insert user into db
		logger.DebugContext(ctx, "Creating user", "email", user.Email)

		query := `
            INSERT INTO users (name, email, password)
            VALUES ($1, $2, $3)
            RETURNING id
        `
		err = db.QueryRowContext(ctx, query, user.Name, user.Email, user.Password).Scan(&user.ID)
		if err != nil {
			logger.ErrorContext(ctx, "failed to insert user", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Respond with created user
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			logger.ErrorContext(ctx, "failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}
