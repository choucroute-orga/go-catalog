package api

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type EchoError struct {
	Code     uint      `json:"code"`
	Message  string    `json:"message"`
	Error    string    `json:"error"`
	IssuedAt time.Time `json:"issued_at"`
}

type ValidationErrors struct {
	Code     uint      `json:"code"`
	Message  string    `json:"message"`
	Error    string    `json:"error"`
	IssuedAt time.Time `json:"issued_at"`
	Errors   []string  `json:"errors"`
}

func NewInternalServerError(err error) error {
	jsonError := EchoError{
		Code:     http.StatusInternalServerError,
		Message:  "Internal Server Error",
		Error:    err.Error(),
		IssuedAt: time.Now(),
	}
	return echo.NewHTTPError(http.StatusInternalServerError, jsonError)
}

func NewConflictError(err error) error {
	jsonError := EchoError{
		Code:     http.StatusConflict,
		Message:  "Conflict Error",
		Error:    err.Error(),
		IssuedAt: time.Now(),
	}
	return echo.NewHTTPError(http.StatusConflict, jsonError)
}

func NewNotFoundError(err error) error {
	jsonError := EchoError{
		Code:     http.StatusNotFound,
		Message:  "Not Found Error",
		Error:    err.Error(),
		IssuedAt: time.Now(),
	}
	return echo.NewHTTPError(http.StatusNotFound, jsonError)
}

func NewUnauthorizedError(err error) error {
	jsonError := EchoError{
		Code:     http.StatusUnauthorized,
		Message:  "Unauthorized Error",
		Error:    err.Error(),
		IssuedAt: time.Now(),
	}
	return echo.NewHTTPError(http.StatusUnauthorized, jsonError)
}

func NewBadRequestError(err error) error {
	jsonError := EchoError{
		Code:     http.StatusBadRequest,
		Message:  "Bad Request Error",
		Error:    err.Error(),
		IssuedAt: time.Now(),
	}
	return echo.NewHTTPError(http.StatusBadRequest, jsonError)
}

// Show the log and return true if there was an error
func FailOnError(logger *logrus.Entry, err error, msg string) bool {
	if err != nil {
		logger.WithError(err).Error(msg)
		return true
	}
	return false
}

func WarnOnError(logger *logrus.Entry, err error, msg string) bool {
	if err != nil {
		logger.WithError(err).Warn(msg)
		return true
	}
	return false
}

func DebugOnError(logger *logrus.Entry, err error, msg string) bool {
	if err != nil {
		logger.WithError(err).Debug(msg)
		return true
	}
	return false
}
