package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/examples/api/layered/internal/middleware"
	"go.opentelemetry.io/otel/codes"
)

// uerDeleter represents a type capable of deleting a user from storage
type userDeleter interface {
	DeleteUser(ctx context.Context, id uint64) error
}

// HandleDeleteUser handles the deletion of a user by ID.
//
// @Summary		Delete User
// @Description	Delete User by ID
// @Tags		user
// @Accept		json
// @Produce		json
// @Param		id	path	string	true	"User ID"
// @Success		204
// @Failure		400	{object}	string
// @Failure		404	{object}	string
// @Failure		500	{object}	string
// @Router		/user/{id}  [DELETE]
func HandleDeleteUser(logger *slog.Logger, userDeleter userDeleter) http.HandlerFunc {
	const name = "handlers.HandleDeleteUser"
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
			span.SetStatus(codes.Error, "ID conversion failed")
			span.RecordError(err)

			_ = encodeResponseJSON(
				w, http.StatusBadRequest, ProblemDetail{
					Title:   "Invalid ID",
					Status:  http.StatusBadRequest,
					Detail:  "The provided ID is not a valid integer.",
					TraceID: middleware.GetTraceID(ctx),
				},
			)

			return
		}

		// Delete the user
		err = userDeleter.DeleteUser(ctx, uint64(id))
		if err != nil {
			logger.ErrorContext(
				ctx,
				"failed to read user",
				slog.String("error", err.Error()),
			)
			span.SetStatus(codes.Error, "user deletion failed")
			span.RecordError(err)

			_ = encodeResponseJSON(w, http.StatusInternalServerError, NewInternalServerError(ctx))

			return
		}

		// Encode the response model as JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
	}
}
