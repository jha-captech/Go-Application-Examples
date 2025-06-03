package app

import (
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// listUsers is an HTTP handler function that retrieves a list of all users from
// the database.
//
// @Summary		List Users
// @Description	List all users
// @Tags		user
// @Produce		json
// @Success		200	{array}	User
// @Failure		500	{object}	string
// @Router		/user	[GET]
func listUsers(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		const funcName = "app.listUsers"
		logger = logger.With(slog.String("func", funcName), getTraceIDAsAttr(ctx))

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
			_ = encodeResponseJSON(
				w,
				http.StatusInternalServerError,
				newInternalServerError(),
			) // ignore the error here because it should never happen with a defined struct

			return
		}

		_ = encodeResponseJSON(w, http.StatusOK, users)
	}
}
