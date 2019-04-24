package rest

import (
	"net/http"

	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/render"
)

// NotFoundHandler will return JSON for the error where a resource is not found
func NotFoundHandler(wr http.ResponseWriter, req *http.Request) {
	render.JSON(wr, http.StatusNotFound, errors.ResourceNotFound("path", req.URL.Path))
}

// MethodNotAllowedHandler will return JSON for a method which is not allowed
func MethodNotAllowedHandler(wr http.ResponseWriter, req *http.Request) {
	render.JSON(wr, http.StatusMethodNotAllowed, errors.ResourceNotFound("path", req.URL.Path))
}
