package app

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
)

// readUser is an HTTP handler function that retrieves a user by ID from the database.
//
//	@Summary		Read User
//	@Description	Read User by ID
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string	true	"User ID"
//	@Success		200				{object}	userResponse
//	@Failure		400				{object}	problemDetail
//	@Failure		404				{object}	problemDetail
//	@Failure		500				{object}	problemDetail
//	@Router			/api/user/{id}	[GET]
func readUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	const funcName = "app.readUser"
	logger = logger.With(slog.String("func", funcName))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		logger = logger.With(getTraceIDAsAttr(ctx))

		// read id from path parameters
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to parse id from url",
				slog.String("id", idStr),
				slog.String("error", err.Error()),
			)

			_ = encodeResponseJSON(w, http.StatusBadRequest, problemDetail{
				Title:   "Bad Request",
				Status:  http.StatusBadRequest,
				Detail:  "The provided ID is not a valid integer.",
				TraceID: getTraceID(ctx),
			})

			return
		}

		// read the user
		logger.InfoContext(ctx, "Reading user", slog.Int("id", id))

		var user user
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
				_ = encodeResponseJSON(w, http.StatusNotFound, problemDetail{
					Title:   "User Not Found",
					Status:  http.StatusNotFound,
					Detail:  fmt.Sprintf("User with ID %d not found", id),
					TraceID: getTraceID(ctx),
				})

				return

			default:
				logger.ErrorContext(
					ctx,
					"failed to read user",
					slog.String("error", err.Error()),
				)

				_ = encodeResponseJSON(w, http.StatusInternalServerError, problemDetail{
					Title:   "Internal Server Error",
					Status:  http.StatusInternalServerError,
					Detail:  "An unexpected error occurred.",
					TraceID: getTraceID(ctx),
				})

				return
			}
		}

		// respond with userResponse (no password)
		resp := userResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		}
		_ = encodeResponseJSON(w, http.StatusOK, resp)
	}
}
