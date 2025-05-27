package app

import (
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
			encodeResponse(w, logger, http.StatusInternalServerError, NewInternalServerError())
			return
		}

		encodeResponse(w, logger, http.StatusOK, users)
	}
}
