package client

import "fmt"

const (
	ExitSuccess     = 0
	ExitInvalidArgs = 2
	ExitNotFound    = 3
	ExitAuthFailure = 4
	ExitAPIError    = 5
	ExitRateLimited = 7
	ExitDryRun      = 9
)

type AppError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	ExitCode int    `json:"-"`
}

func (e AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func InvalidArgs(message string) error {
	return AppError{Code: "INVALID_ARGUMENTS", Message: message, ExitCode: ExitInvalidArgs}
}

func NotFound(message string) error {
	return AppError{Code: "NOT_FOUND", Message: message, ExitCode: ExitNotFound}
}

func AuthFailure(message string) error {
	return AppError{Code: "AUTH_FAILURE", Message: message, ExitCode: ExitAuthFailure}
}

func APIError(message string) error {
	return AppError{Code: "API_ERROR", Message: message, ExitCode: ExitAPIError}
}

func RateLimited(message string) error {
	return AppError{Code: "RATE_LIMITED", Message: message, ExitCode: ExitRateLimited}
}

func ReadOnlyError() error {
	return AppError{Code: "READ_ONLY", Message: "cars-il-pp-cli is a read-only tool.", ExitCode: ExitInvalidArgs}
}
