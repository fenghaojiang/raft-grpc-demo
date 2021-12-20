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

//TODO 
//var (
//	_messages atomic.Value
//	_code     = map[int]struct{}{}
//)
//
//func Register(cm map[int]string) {
//	_messages.Store(cm)
//}
