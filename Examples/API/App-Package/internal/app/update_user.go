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

// readUser is an HTTP handler function that reads a user by ID from the database.
//
// @Summary		Update User
// @Description	Update user fields by ID
// @Tags		user
// @Accept		json
// @Produce		json
// @Param		id		path	string	true	"User ID"
// @Param		user	body	User	true	"User data"
// @Success		200		{object}	User
// @Failure		400		{object}	string
// @Failure		404		{object}	string
// @Failure		500		{object}	string
// @Router		/user/{id} [PUT]
func updateUser(logger *slog.Logger, db *sqlx.DB) http.HandlerFunc {
	const funcName = "app.updateUser"
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
				Title:   "Invalid ID",
				Status:  http.StatusBadRequest,
				Detail:  "The provided ID is not a valid integer.",
				TraceID: getTraceID(ctx),
			})

			return
		}

		// request validation
		req, problems, err := decodeValid[userRequest](r)
		if err != nil && len(problems) == 0 {
			logger.ErrorContext(
				ctx,
				"failed to decode request",
				slog.String("error", err.Error()),
			)
			_ = encodeResponseJSON(
				w, http.StatusBadRequest, problemDetail{
					Title:   "Bad Request",
					Status:  400,
					Detail:  "Invalid request body.",
					TraceID: getTraceID(ctx),
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
			ctx, "Updating user",
			slog.Int("id", id),
			slog.String("name", req.Name),
			slog.String("email", req.Email),
		)

		// update user in db
		var updatedUser user
		err = db.GetContext(
			ctx,
			&updatedUser,
			`
            UPDATE users
            SET name     = $1,
                email    = $2,
                password = $3
            WHERE id = $4
            RETURNING id, name, email, password
            `,
			req.Name,
			req.Email,
			req.Password,
			id,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_ = encodeResponseJSON(w, http.StatusNotFound, problemDetail{
					Title:   "User Not Found",
					Status:  http.StatusNotFound,
					Detail:  fmt.Sprintf("User with ID %d not found", id),
					TraceID: getTraceID(ctx),
				})

				return
			}

			logger.ErrorContext(ctx, "failed to update user", slog.String("error", err.Error()))
			_ = encodeResponseJSON(w, http.StatusInternalServerError, problemDetail{
				Title:   "Internal Server Error",
				Status:  500,
				Detail:  "An unexpected error occurred.",
				TraceID: getTraceID(ctx),
			})
			
			return
		}

		logger.InfoContext(
			ctx, "User updated successfully",
			slog.Uint64("id", uint64(updatedUser.ID)),
			slog.String("name", updatedUser.Name),
			slog.String("email", updatedUser.Email),
		)

		// respond with updated user (without password)
		resp := userResponse{
			ID:    updatedUser.ID,
			Name:  updatedUser.Name,
			Email: updatedUser.Email,
		}
		_ = encodeResponseJSON(w, http.StatusOK, resp)
	}
}
