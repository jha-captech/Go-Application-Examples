package app

import (
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// listUsers is an HTTP handler function that retrieves a list of all users from
// the database.
//
//	@Summary		List Users
//	@Description	List all users
//	@Tags			user
//	@Produce		json
//	@Success		200			{array}		user
//	@Failure		500			{object}	problemDetail
//	@Router			/api/user	[GET]
func listUsers(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	const funcName = "app.listUsers"
	logger = logger.With(slog.String("func", funcName))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger = logger.With(getTraceIDAsAttr(ctx))

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
			_ = encodeResponseJSON(w, http.StatusInternalServerError, problemDetail{
				Title:   "Internal Server Error",
				Status:  http.StatusInternalServerError,
				Detail:  "An unexpected error occurred.",
				TraceID: getTraceID(ctx),
			})

			return
		}

		// Convert []user to []userResponse to exclude password
		userResponses := make([]userResponse, 0, len(users))
		for _, u := range users {
			userResponses = append(userResponses, userResponse{
				ID:    u.ID,
				Name:  u.Name,
				Email: u.Email,
			})
		}

		_ = encodeResponseJSON(w, http.StatusOK, userResponses)
	}
}
