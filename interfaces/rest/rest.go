package rest

import (
	"net/http"

	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/render"
)

func NotFoundHandler(wr http.ResponseWriter, req *http.Request) {
	render.JSON(wr, http.StatusNotFound, errors.ResourceNotFound("path", req.URL.Path))
}

func MethodNotAllowedHandler(wr http.ResponseWriter, req *http.Request) {
	render.JSON(wr, http.StatusMethodNotAllowed, errors.ResourceNotFound("path", req.URL.Path))
}
