package error_code

import "errors"

var (
	BadRequest          = errors.New("bad request")
	ServiceUnavailable  = errors.New("service unavailable")
	TemporaryRedirect   = errors.New("temporary redirect")
	InternalServerError = errors.New("internal server error")
)
