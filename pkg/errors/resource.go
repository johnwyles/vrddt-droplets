package errors

import "net/http"

// Common resource related error codes.
const (
	TypeResourceNotFound = "ResourceNotFound"
	TypeResourceConflict = "ResourceConflict"
)

// ResourceLimit returns an error that represents a value which exceeds a
// given threshold
func ResourceLimit(rType string, limit interface{}) error {
	return WithStack(&Error{
		Code:    http.StatusInternalServerError,
		Type:    TypeResourceNotFound,
		Message: "Resource limit has been exceeded",
		Context: map[string]interface{}{
			"resource_type": rType,
			"limit":         limit,
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
			"resource_type": rType,
			"resource_id":   rID,
		},
	})
}

// Conflict returns an error that represents a resource identifier conflict
func Conflict(rType string, rID string) error {
	return WithStack(&Error{
		Code:    http.StatusConflict,
		Type:    TypeResourceConflict,
		Message: "A resource with same name already exists",
		Context: map[string]interface{}{
			"resource_type": rType,
			"resource_id":   rID,
		},
	})
}
