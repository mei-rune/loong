package util

import (
	"errors"
	"net/http"
)

func ErrBadArgument(paramName string, value interface{}, err error) error {
	return &HTTPError{httpCode: http.StatusBadRequest, err: errors.New("param '" + paramName + "' is invalid - " + err.Error())}
}

type HTTPError struct {
	err      error
	httpCode int
}

func (e *HTTPError) Error() string {
	return e.err.Error()
}

func (e *HTTPError) HTTPCode() int {
	return e.httpCode
}

func WithHTTPCode(code int, err error) *HTTPError {
	return &HTTPError{err: err, httpCode: code}
}

func Wrap(err error, msg string) error {
	if he, ok := err.(*HTTPError); ok {
		he.err = errors.New(msg + ": " + he.err.Error())
		return he
	}
	return errors.New(msg + ": " + err.Error())
}
