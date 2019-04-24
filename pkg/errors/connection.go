package errors

import (
	"net/http"
)

// Common resource related error codes.
const (
	TypeConnectionFailure = "ConnectionFailure"
	TypeConnectionTimeout = "ConnectionTimeout"
)

// ConnectionFailure returns an error that represents an attempt to make a
// connection to a  service which failes
func ConnectionFailure(service string, description string) error {
	return WithStack(&Error{
		Code:    http.StatusInternalServerError,
		Type:    TypeConnectionFailure,
		Message: "Connection failure",
		Context: map[string]interface{}{
			"service":     service,
			"description": description,
		},
	})
}

// ConnectionTimeout returns an error that represents an attempt to access a
// resource which exceeds some timeout.
func ConnectionTimeout(cType string, timeout int) error {
	return WithStack(&Error{
		Code:    http.StatusRequestTimeout,
		Type:    TypeConnectionTimeout,
		Message: "Connection has timed out",
		Context: map[string]interface{}{
			"operation": cType,
			"timeout":   timeout,
		},
	})
}
