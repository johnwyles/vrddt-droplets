package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
	"github.com/johnwyles/vrddt-droplets/pkg/render"
)

// Controller holds the information about the REST controller
type Controller struct {
	Router *mux.Router
}

// New initializes the server with routes exposing the given usecases.
func New(loggerHandle logger.Logger) *Controller {
	controller := &Controller{
		Router: mux.NewRouter(),
	}

	// Setup router with default handlers
	controller.Router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	controller.Router.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	return controller
}

func notFoundHandler(wr http.ResponseWriter, req *http.Request) {
	render.JSON(wr, http.StatusNotFound, errors.ResourceNotFound("path", req.URL.Path))
}

func methodNotAllowedHandler(wr http.ResponseWriter, req *http.Request) {
	render.JSON(wr, http.StatusMethodNotAllowed, errors.ResourceNotFound("path", req.URL.Path))
}
