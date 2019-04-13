package errors

import (
	"net/http"
)

// Common resource related error codes.
const (
	TypeOperationTimeout = "OperationTimeout"
)

// OperationTimeout returns an error that represents an attempt to access a
// resource which exceeds some timeout.
func OperationTimeout(operation string, timeout int) error {
	return WithStack(&Error{
		Code:    http.StatusRequestTimeout,
		Type:    TypeOperationTimeout,
		Message: "Operation has timed out",
		Context: map[string]interface{}{
			"operation": operation,
			"timeout":   timeout,
		},
	})
}
