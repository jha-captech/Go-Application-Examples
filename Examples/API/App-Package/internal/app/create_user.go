package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// @Summary		Create User
// @Description	Create a new user
// @Tags		user
// @Accept		json
// @Produce		json
// @Param		user	body		User	true	"User data"
// @Success		201		{object}	User
// @Failure		400		{object}	string
// @Failure		500		{object}	string
// @Router		/user [POST]
func createUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		const funcName = "app.createUser"
		logger = logger.With(
			slog.String("func", funcName),
			getTraceIDAsAttr(ctx),
		)

		// request validation
		user, problems, err := decodeValid[user](r)
		if err != nil && len(problems) == 0 {
			logger.ErrorContext(
				ctx,
				"failed to decode request",
				slog.String("error", err.Error()),
			)

			// ignore the error here because it should never happen with a defined struct
			_ = encodeResponseJSON(
				w,
				http.StatusBadRequest,
				problemDetail{
					Title:  "Bad Request",
					Status: 400,
					Detail: "Invalid request body.",
				},
			)

			return
		}

		if len(problems) > 0 {
			logger.ErrorContext(
				ctx,
				"Validation error",
				slog.String("Validation errors: ", fmt.Sprintf("%#v", problems)),
			)
			_ = encodeResponseJSON(w, http.StatusBadRequest, newValidationBadRequest(problems))

			return
		}

		logger.InfoContext(
			ctx, "Creating user",
			slog.String("name", user.Name),
			slog.String("email", user.Email),
		)

		// insert user into db
		err = db.GetContext(
			ctx,
			&user.ID,
			`
			INSERT INTO users (name, email, password)
			VALUES ($1, $2, $3)
			RETURNING id
			`,
			user.Name,
			user.Email,
			user.Password,
		)
		if err != nil {
			logger.ErrorContext(ctx, "failed to insert user", slog.String("error", err.Error()))
			_ = encodeResponseJSON(w, http.StatusInternalServerError, newInternalServerError())

			return
		}

		logger.InfoContext(
			ctx, "User created successfully",
			slog.Uint64("id", uint64(user.ID)),
			slog.String("name", user.Name),
			slog.String("email", user.Email),
		)

		// respond with created user
		_ = encodeResponseJSON(w, http.StatusCreated, user)
	}
}
