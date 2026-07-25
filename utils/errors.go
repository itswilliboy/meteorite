package utils

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func NewHTTPError(status int, message string) *HTTPError {
	return &HTTPError{Status: status, Message: message}
}

func NotFoundIfNoRows(err error, message string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return NewHTTPError(http.StatusNotFound, message)
	}
	return err
}
