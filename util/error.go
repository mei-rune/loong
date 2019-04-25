package util

import (
	"errors"
	"net/http"
)

type HTTPError interface {
	error

	HTTPCode() int
}

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) HTTPCode() int {
	return e.Code
}

func ErrBadArgument(paramName string, value interface{}, err ...error) error {
	if len(err) == 0 {
		return &Error{Code: http.StatusBadRequest, Message: "param '" + paramName + "' is invalid"}
	}
	return &Error{Code: http.StatusBadRequest, Message: "param '" + paramName + "' is invalid - " + err[0].Error()}
}

type httpError struct {
	err      error
	httpCode int
}

func (e *httpError) Error() string {
	return e.err.Error()
}

func (e *httpError) HTTPCode() int {
	return e.httpCode
}

func WithHTTPCode(code int, err error) HTTPError {
	if err == nil {
		panic(errors.New("err is nil"))
	}
	return &httpError{err: err, httpCode: code}
}

func Wrap(err error, msg string) error {
	if he, ok := err.(*Error); ok {
		he.Message = msg + ": " + he.Message
		return he
	}
	if he, ok := err.(HTTPError); ok {
		return &Error{
			Code:    he.HTTPCode(),
			Message: msg + ": " + he.Error(),
		}
	}
	return errors.New(msg + ": " + err.Error())
}

func ToError(err error, defaultCode int) *Error {
	if he, ok := err.(*Error); ok {
		return he
	}

	if he, ok := err.(HTTPError); ok {
		return &Error{
			Code:    he.HTTPCode(),
			Message: he.Error(),
		}
	}

	return &Error{
		Code:    defaultCode,
		Message: err.Error(),
	}
}
