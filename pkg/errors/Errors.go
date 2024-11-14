package errors

import (
	"errors"
)

const OpenApiSpecNotFoundError = "openapi-spec not found for the k8s version %s"

var ErrOpenApiSpecNotFound = errors.New(OpenApiSpecNotFoundError)
