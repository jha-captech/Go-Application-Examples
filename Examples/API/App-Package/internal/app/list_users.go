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
		
		const funcName = "app.listUsers"
		logger = logger.With(
			slog.String("func", funcName),
			getTraceIDAsAtter(ctx),
		)

		logger.InfoContext(ctx, "Listing all users")

		// query db to get all users
		var users []user
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
			_ = encodeResponse(w, http.StatusInternalServerError, newInternalServerError()) // ignore the error here because it should never happen with a defined struct
			return
		}

		_ = encodeResponse(w, http.StatusOK, users)
	}
}
