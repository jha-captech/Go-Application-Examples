package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel"

	"example.com/examples/api/layered/internal/middleware"
)

const name = "example.com/examples/api/layered/internal/handlers"

var tracer = otel.Tracer(name)

// UserRequest represents the request for creating a user.
type UserRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=50"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=30"`
}

// UserResponse represents the response for creating a user.
type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ProblemDetail represents the structure for problem details as per RFC 7807.
type ProblemDetail struct {
	Title   string `json:"title"`
	Status  int    `json:"status"`
	Detail  string `json:"detail"`
	TraceID string `json:"traceId,omitempty"`
}

// validationProblem represents a single validation error detail.
type validationProblem struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ProblemDetailValidation extends ProblemDetail to include validation errors.
type ProblemDetailValidation struct {
	ProblemDetail
	InvalidParams []validationProblem `json:"invalidParams"` // A list of invalid parameters with error details.
}

// decodeValid decodes a model from an http request and performs validation
// on it.
func decodeValid[T any](r *http.Request) (T, []validationProblem, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, []validationProblem{}, fmt.Errorf(
			"[in handlers.decodeValid] decode body failed: %w",
			err,
		)
	}
	problems, err := validate(&v)
	if err != nil {
		return v, []validationProblem{}, fmt.Errorf(
			"[in handlers.decodeValid] validate failed: %w",
			err,
		)
	}
	if len(problems) > 0 {
		validationProblems := make([]validationProblem, len(problems))
		for i, problem := range problems {
			validationProblems[i] = validationProblem{
				Field:   problem.Field(),
				Code:    problem.Tag(),
				Message: problem.Error(),
			}
		}

		return v, validationProblems, nil
	}

	return v, []validationProblem{}, nil
}

// validate validates the provided data using the validator package.
func validate[T any](data *T, options ...validator.Option) ([]validator.FieldError, error) {
	v := validator.New(options...)
	if err := v.Struct(data); err != nil {
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			return []validator.FieldError{}, fmt.Errorf(
				"[in handlers.validate] invalid validation error: %w",
				err,
			)
		}
		return err.(validator.ValidationErrors), nil
	}
	return []validator.FieldError{}, nil
}

// NewValidationBadRequest creates a ProblemDetailValidation instance for a 400 Bad Request validation error.
func NewValidationBadRequest(
	ctx context.Context,
	invalidParams []validationProblem,
) ProblemDetailValidation {
	return ProblemDetailValidation{
		ProblemDetail: ProblemDetail{
			Title:   "Bad Request",
			Status:  http.StatusBadRequest,
			Detail:  "The request contains invalid parameters.",
			TraceID: middleware.GetTraceID(ctx),
		},
		InvalidParams: invalidParams,
	}
}

// NewNotFound is a helper that creates a ProblemDetail instance for a 500 error.
func NewInternalServerError(ctx context.Context) ProblemDetail {
	return ProblemDetail{
		Title:   "Internal Server Error",
		Status:  http.StatusInternalServerError,
		Detail:  "An unexpected error occurred.",
		TraceID: middleware.GetTraceID(ctx),
	}
}
