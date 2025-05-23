package app

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
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
func createUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

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

		logger.InfoContext(ctx, "Creating user",
			slog.String("name", user.Name),
			slog.String("email", user.Email),
		)

		// Insert user into db
		query := `
			INSERT INTO users (name, email, password)
			VALUES ($1, $2, $3)
			RETURNING id
		`
		err = db.GetContext(ctx, &user.ID, query, user.Name, user.Email, user.Password)
		if err != nil {
			logger.ErrorContext(ctx, "failed to insert user", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		logger.InfoContext(ctx, "User created successfully",
			slog.Uint64("id", uint64(user.ID)),
			slog.String("name", user.Name),
			slog.String("email", user.Email),
		)

		// Respond with created user
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(user); err != nil {
			logger.ErrorContext(ctx, "failed to encode response", slog.String("error", err.Error()))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
