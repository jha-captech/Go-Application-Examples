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

// @Summary		Update User
// @Description	Update user fields by ID
// @Tags			user
// @Accept			json
// @Produce		json
// @Param			id		path	string	true	"User ID"
// @Param			user	body	User	true	"User data"
// @Success		200		{object}	User
// @Failure		400		{object}	string
// @Failure		404		{object}	string
// @Failure		500		{object}	string
// @Router			/user/{id} [PUT]
func updateUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
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

		// Request validation
		user, problems, err := decodeValid[User](r)
		if err != nil && len(problems) == 0 {
			logger.ErrorContext(
				ctx,
				"failed to decode request",
				slog.String("error", err.Error()))

			encodeErr := encodeResponse(w, http.StatusBadRequest, ProblemDetail{
				Title:  "Bad Request",
				Status: 400,
				Detail: "Invalid request body.",
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
		if len(problems) > 0 {
			logger.ErrorContext(
				ctx,
				"Validation error",
				slog.String("Validation errors: ", fmt.Sprintf("%#v", problems)),
			)
			encodeErr := encodeResponse(w, http.StatusBadRequest, NewValidationBadRequest(problems))
			if encodeErr != nil {
				logger.ErrorContext(
					ctx,
					"failed to encode response",
					slog.String("error", encodeErr.Error()),
				)
			}
			return
		}

		logger.InfoContext(ctx, "Updating user",
			slog.Int("id", id),
			slog.String("name", user.Name),
			slog.String("email", user.Email),
		)

		// Update user in db
		query := `
            UPDATE users
            SET name = $1, email = $2, password = $3
            WHERE id = $4
            RETURNING id, name, email, password
        `
		err = db.GetContext(ctx, &user, query, user.Name, user.Email, user.Password, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
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
			}
			logger.ErrorContext(ctx, "failed to update user", slog.String("error", err.Error()))
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

		logger.InfoContext(ctx, "User updated successfully",
			slog.Uint64("id", uint64(user.ID)),
			slog.String("name", user.Name),
			slog.String("email", user.Email),
		)

		// Respond with updated user
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
