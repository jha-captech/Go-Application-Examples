package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// createUser is an HTTP handler function that creates a new user in the database.
//
//	@Summary		Create User
//	@Description	Create a new user
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			user	body		user	true	"User data"
//	@Success		201		{object}	userResponse
//	@Failure		400		{object}	problemDetailValidation
//	@Failure		500		{object}	problemDetail
//	@Router			/api/user [POST]
func createUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	const funcName = "app.createUser"
	logger = logger.With(slog.String("func", funcName))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger = logger.With(getTraceIDAsAttr(ctx))

		// request validation
		req, problems, err := decodeValid[userRequest](r)
		if err != nil && len(problems) == 0 {
			logger.ErrorContext(
				ctx,
				"failed to decode request",
				slog.String("error", err.Error()),
			)
			_ = encodeResponseJSON(w, http.StatusBadRequest, problemDetail{
				Title:   "Bad Request",
				Status:  400,
				Detail:  "Invalid request body.",
				TraceID: getTraceID(ctx),
			})
			return
		}

		if len(problems) > 0 {
			logger.ErrorContext(
				ctx,
				"Validation error",
				slog.String("Validation errors: ", fmt.Sprintf("%#v", problems)),
			)
			_ = encodeResponseJSON(w, http.StatusBadRequest, problemDetailValidation{
				problemDetail: problemDetail{
					Title:   "Bad Request",
					Status:  400,
					Detail:  "The request contains invalid parameters.",
					TraceID: getTraceID(ctx),
				},
				InvalidParams: problems,
			})
			return
		}

		logger.InfoContext(
			ctx, "Creating user",
			slog.String("name", req.Name),
			slog.String("email", req.Email),
		)

		// insert user into db
		var id uint
		err = db.GetContext(
			ctx,
			&id,
			`
			INSERT INTO users (name, email, password)
			VALUES ($1, $2, $3)
			RETURNING id
			`,
			req.Name,
			req.Email,
			req.Password,
		)
		if err != nil {
			logger.ErrorContext(ctx, "failed to insert user", slog.String("error", err.Error()))
			_ = encodeResponseJSON(w, http.StatusInternalServerError, problemDetail{
				Title:   "Internal Server Error",
				Status:  500,
				Detail:  "An unexpected error occurred.",
				TraceID: getTraceID(ctx),
			})
			return
		}

		logger.InfoContext(
			ctx, "User created successfully",
			slog.Uint64("id", uint64(id)),
			slog.String("name", req.Name),
			slog.String("email", req.Email),
		)

		// respond with created user (without password)
		resp := userResponse{
			ID:    id,
			Name:  req.Name,
			Email: req.Email,
		}
		_ = encodeResponseJSON(w, http.StatusCreated, resp)
	}
}
