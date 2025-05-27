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
// @Tags			user
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"User ID"
// @Success		200	{object}	models.User
// @Failure		400	{object}	string
// @Failure		404	{object}	string
// @Failure		500	{object}	string
// @Router			/user/{id}  [GET]
func readUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Read id from path parameters
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to parse id from url",
				slog.String("id", idStr),
				slog.String("error", err.Error()),
			)
			encodeErr := encodeResponse(w, http.StatusBadRequest, ProblemDetail{
				Title:  "Invalid ID",
				Status: http.StatusBadRequest,
				Detail: "The provided ID is not a valid integer.",
			})
			if encodeErr != nil {
				logger.ErrorContext(
					ctx,
					"failed to encode response",
					slog.String("error", encodeErr.Error()),
				)
			}
			return
		}

		// Read the user
		logger.InfoContext(ctx, "Reading user", slog.Int("id", id))

		var user User
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
				encodeErr := encodeResponse(w, http.StatusNotFound, ProblemDetail{
					Title:  "User Not Found",
					Status: http.StatusNotFound,
					Detail: fmt.Sprintf("User with ID %d not found", id),
				})
				if encodeErr != nil {
					logger.ErrorContext(
						ctx,
						"failed to encode response",
						slog.String("error", encodeErr.Error()),
					)
				}
				return
			default:
				logger.ErrorContext(
					ctx,
					"failed to read user",
					slog.String("error", err.Error()),
				)
				encodeErr := encodeResponse(w, http.StatusInternalServerError, NewInternalServerError())
				if encodeErr != nil {
					logger.ErrorContext(
						ctx,
						"failed to encode response",
						slog.String("error", encodeErr.Error()),
					)
				}
				return
			}
		}

		// Respond with user as JSON
		encodeErr := encodeResponse(w, http.StatusOK, user)
		if encodeErr != nil {
			logger.ErrorContext(
				ctx,
				"failed to encode response",
				slog.String("error", encodeErr.Error()),
			)
		}
	}
}
