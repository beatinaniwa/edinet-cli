package api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// ErrorCode represents a normalized error classification.
type ErrorCode string

const (
	ErrAuth       ErrorCode = "AUTH_REQUIRED"
	ErrAuthFailed ErrorCode = "AUTH_FAILED"
	ErrBadRequest ErrorCode = "BAD_REQUEST"
	ErrNotFound   ErrorCode = "NOT_FOUND"
	ErrServer     ErrorCode = "SERVER_ERROR"
	ErrNetwork    ErrorCode = "NETWORK_ERROR"
	ErrTimeout    ErrorCode = "TIMEOUT"
	ErrValidation ErrorCode = "VALIDATION_ERROR"
	ErrInternal   ErrorCode = "INTERNAL_ERROR"
)

// EDINETError represents a unified error from the EDINET API.
// It handles both the normal error format (metadata.status) and
// the 401-specific format (StatusCode).
type EDINETError struct {
	Code    ErrorCode `json:"code"`
	Status  int       `json:"status_code"`
	Message string    `json:"message"`
	Raw     string    `json:"raw_response,omitempty"`
}

func (e *EDINETError) Error() string {
	return fmt.Sprintf("%s: %s (status %d)", e.Code, e.Message, e.Status)
}

// ExitCode returns the CLI exit code for this error.
func (e *EDINETError) ExitCode() int {
	switch e.Code {
	case ErrAuth, ErrAuthFailed:
		return ExitAuth
	case ErrValidation:
		return ExitValidation
	case ErrBadRequest, ErrNotFound, ErrServer:
		return ExitAPI
	default:
		return ExitGeneral
	}
}

// Exit codes for the CLI.
const (
	ExitOK         = 0
	ExitGeneral    = 1
	ExitValidation = 2
	ExitAuth       = 3
	ExitAPI        = 4
)

// ParseErrorResponse parses an EDINET error response body.
// It handles both the normal format (metadata.status) and the 401 format (StatusCode).
// Returns nil if the response indicates success (status "200").
func ParseErrorResponse(body []byte) *EDINETError {
	if len(body) == 0 {
		return &EDINETError{Code: ErrInternal, Message: "empty response body"}
	}

	// Try normal error format first: {"metadata": {"status": "400", "message": "..."}}
	var normal ErrorResponse
	if err := json.Unmarshal(body, &normal); err == nil && normal.Metadata.Status != "" {
		if normal.Metadata.Status == "200" {
			return nil
		}
		status, _ := strconv.Atoi(normal.Metadata.Status)
		return &EDINETError{
			Code:    statusToErrorCode(status),
			Status:  status,
			Message: normal.Metadata.Message,
		}
	}

	// Try 401 auth error format: {"StatusCode": 401, "message": "..."}
	var auth AuthErrorResponse
	if err := json.Unmarshal(body, &auth); err == nil && auth.StatusCode != 0 {
		return &EDINETError{
			Code:    ErrAuthFailed,
			Status:  auth.StatusCode,
			Message: auth.Message,
		}
	}

	// Malformed or unrecognized response
	return &EDINETError{
		Code:    ErrInternal,
		Message: "unrecognized response format",
		Raw:     string(body),
	}
}

func statusToErrorCode(status int) ErrorCode {
	switch status {
	case 400:
		return ErrBadRequest
	case 401:
		return ErrAuthFailed
	case 404:
		return ErrNotFound
	case 500:
		return ErrServer
	default:
		return ErrInternal
	}
}
