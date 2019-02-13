package rest

import (
	"net/http"

	"github.com/johnwyles/vrddt-droplets/pkg/render"

	"github.com/gorilla/mux"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// New initializes the server with routes exposing the given usecases.
func New(logger logger.Logger, reg registration, ret retriever, postsRet postRetriever, postPub postPublication) http.Handler {
	// setup router with default handlers
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	// setup api endpoints
	addUsersAPI(router, reg, ret, logger)
	addPostsAPI(router, postPub, postsRet, logger)

	return router
}

func notFoundHandler(wr http.ResponseWriter, req *http.Request) {
	render.JSON(wr, http.StatusNotFound, errors.ResourceNotFound("path", req.URL.Path))
}

func methodNotAllowedHandler(wr http.ResponseWriter, req *http.Request) {
	render.JSON(wr, http.StatusMethodNotAllowed, errors.ResourceNotFound("path", req.URL.Path))
}
