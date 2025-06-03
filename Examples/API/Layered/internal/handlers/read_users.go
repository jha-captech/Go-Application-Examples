package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"go.opentelemetry.io/otel/codes"

	"example.com/examples/api/layered/internal/models"
)

// userReader represents a type capable of reading a user from storage and
// returning it or an error.
type userReader interface {
	ReadUser(ctx context.Context, id uint64) (models.User, error)
}

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
func HandleReadUser(logger *slog.Logger, userReader userReader) http.HandlerFunc {
	const name = "handlers.HandleReadUser"
	logger = logger.With(slog.String("func", name))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), name)
		defer span.End()

		// Read id from path parameters
		idStr := r.PathValue("id")

		// Convert the ID from string to int
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to parse id from url",
				slog.String("id", idStr),
				slog.String("error", err.Error()),
			)
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			_ = encodeResponseJSON(w, http.StatusBadRequest, ProblemDetail{
				Title:  "Invalid ID",
				Status: http.StatusBadRequest,
				Detail: "The provided ID is not a valid integer.",
			})

			return
		}

		// Read the user
		user, err := userReader.ReadUser(ctx, uint64(id))
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to read user",
				slog.String("error", err.Error()),
			)
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			_ = encodeResponseJSON(w, http.StatusInternalServerError, NewInternalServerError())

			return
		}

		// Encode the response model as JSON
		_ = encodeResponseJSON(w, http.StatusOK, UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		})
	}
}
