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

// @Summary		Read User
// @Description	Read User by ID
// @Tags		user
// @Accept		json
// @Produce		json
// @Param		id	path		string	true	"User ID"
// @Success		200	{object}	models.User
// @Failure		400	{object}	string
// @Failure		404	{object}	string
// @Failure		500	{object}	string
// @Router		/user/{id}	[GET]
func readUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		const funcName = "app.readUser"
		logger = logger.With(slog.String("func", funcName), getTraceIDAsAttr(ctx))

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
			_ = encodeResponseJSON(
				w,
				http.StatusBadRequest,
				problemDetail{
					// ignore the error here because it should never happen with a defined struct
					Title:  "Bad Request",
					Status: http.StatusBadRequest,
					Detail: "The provided ID is not a valid integer.",
				},
			)

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
				_ = encodeResponseJSON(
					w, http.StatusNotFound, problemDetail{
						Title:  "User Not Found",
						Status: http.StatusNotFound,
						Detail: fmt.Sprintf("User with ID %d not found", id),
					},
				)
				return
			default:
				logger.ErrorContext(
					ctx,
					"failed to read user",
					slog.String("error", err.Error()),
				)
				_ = encodeResponseJSON(w, http.StatusInternalServerError, newInternalServerError())
				return
			}
		}

		// respond with user as JSON
		_ = encodeResponseJSON(w, http.StatusOK, user)
	}
}
