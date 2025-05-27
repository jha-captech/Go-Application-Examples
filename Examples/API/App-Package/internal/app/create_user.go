package app

import (
	"fmt"
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

		logger.InfoContext(ctx, "User created successfully",
			slog.Uint64("id", uint64(user.ID)),
			slog.String("name", user.Name),
			slog.String("email", user.Email),
		)

		// Respond with created user
		encodeErr := encodeResponse(w, http.StatusCreated, user)
		if encodeErr != nil {
			logger.ErrorContext(
				ctx,
				"failed to encode response",
				slog.String("error", encodeErr.Error()),
			)
		}
	}
}
