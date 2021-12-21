package ecode

import (
	"errors"
)

var (
	BadRequest            = errors.New("bad request")
	ServiceUnavailable    = errors.New("service unavailable")
	TemporaryRedirect     = errors.New("temporary redirect")
	InternalServerError   = errors.New("internal server error")
	NoTypeIDError         = errors.New("no type id ")
	ErrNoAvailableService = errors.New("no service available")
)
