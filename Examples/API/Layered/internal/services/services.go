package services

import (
	"go.opentelemetry.io/otel"
)

const name = "example.com/examples/api/layered/internal/services"

var tracer = otel.Tracer(name)
