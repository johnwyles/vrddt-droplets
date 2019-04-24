package errors

import "net/http"

// Common resource related error codes.
const (
	TypeNotImplemented   = "NotImplemented"
	TypeResourceConflict = "ResourceConflict"
	TypeResourceNotFound = "ResourceNotFound"
	TypeResourceUnknown  = "ResourceUnknown"
)

// Conflict returns an error that represents a resource identifier conflict
func Conflict(rType string, rID string) error {
	return WithStack(&Error{
		Code:    http.StatusConflict,
		Type:    TypeResourceConflict,
		Message: "A resource with same name already exists",
		Context: map[string]interface{}{
			"type": rType,
			"id":   rID,
		},
	})
}

// NotImplemented returns an error that represents that the feature is not
// implemented yet
func NotImplemented(request string, reason string) error {
	return WithStack(&Error{
		Code:    http.StatusNotImplemented,
		Type:    TypeNotImplemented,
		Message: "The request has not been implemented yet",
		Context: map[string]interface{}{
			"request": request,
			"reason":  reason,
		},
	})
}

// ResourceLimit returns an error that represents a value which exceeds a
// given threshold
func ResourceLimit(rType string, limit interface{}) error {
	return WithStack(&Error{
		Code:    http.StatusInternalServerError,
		Type:    TypeResourceNotFound,
		Message: "Resource limit has been exceeded",
		Context: map[string]interface{}{
			"type":  rType,
			"limit": limit,
		},
	})
}

// ResourceNotFound returns an error that represents an attempt to access a
// non-existent resource
func ResourceNotFound(rType string, rID string) error {
	return WithStack(&Error{
		Code:    http.StatusNotFound,
		Type:    TypeResourceNotFound,
		Message: "Resource you are requesting does not exist",
		Context: map[string]interface{}{
			"type": rType,
			"id":   rID,
		},
	})
}

// ResourceUnknown returns an error that represents an attempt to access an
// unknown resource
func ResourceUnknown(rType string, resource string) error {
	return WithStack(&Error{
		Code:    http.StatusUnsupportedMediaType,
		Type:    TypeResourceUnknown,
		Message: "Resource you are accessing is unknown or not defined",
		Context: map[string]interface{}{
			"type":     rType,
			"resource": resource,
		},
	})
}
