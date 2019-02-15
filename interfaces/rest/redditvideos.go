package rest

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

func addRedditVideosAPI(router *mux.Router, cons redditConstructor, des redditDestructor, ret redditRetriever, logger logger.Logger) {
	rvc := &redditVideosController{
		Logger: logger,
		cons:   cons,
		des:    des,
		ret:    ret,
	}

	router.HandleFunc("/reddit_videos/", rvc.create).Methods(http.MethodPost)
	router.HandleFunc("/reddit_videos/{id}", rvc.getByID).Methods(http.MethodGet)
	router.HandleFunc("/reddit_videos/{url}", rvc.getByURL).Methods(http.MethodGet)
	router.HandleFunc("/reddit_videos/", rvc.search).Methods(http.MethodGet)
}

type redditVideosController struct {
	logger.Logger
	cons redditConstructor
	des  redditDestructor
	ret  redditRetriever
}

func (rvc *redditVideosController) create(wr http.ResponseWriter, req *http.Request) {
	redditVideo := domain.RedditVideo{}
	if err := readRequest(req, &redditVideo); err != nil {
		rvc.Warnf("failed to read reddit video request: %s", err)
		respond(wr, http.StatusBadRequest, err)
		return
	}

	registered, err := rvc.cons.Create(req.Context(), &redditVideo)
	if err != nil {
		rvc.Warnf("failed to create reddit video: %s", err)
		respondErr(wr, err)
		return
	}

	rvc.Infof("reddit video created with id '%s'", registered.ID)
	respond(wr, http.StatusCreated, registered)
}

// TODO: Delete vrddt video if no other reddit videos are associated
func (rvc *redditVideosController) delete(wr http.ResponseWriter, req *http.Request) {
	id := bson.ObjectIdHex(mux.Vars(req)["id"])
	redditVideo, err := rvc.des.Delete(req.Context(), id)
	if err != nil {
		respondErr(wr, err)
		return
	}

	respond(wr, http.StatusOK, redditVideo)
}

func (rvc *redditVideosController) getByID(wr http.ResponseWriter, req *http.Request) {
	id := bson.ObjectIdHex(mux.Vars(req)["id"])
	redditVideo, err := rvc.ret.GetByID(req.Context(), id)
	switch err.(type) {
	case nil:
		respondErr(wr, err)
		return
	}

	respond(wr, http.StatusOK, redditVideo)
}

func (rvc *redditVideosController) getByURL(wr http.ResponseWriter, req *http.Request) {
	url := mux.Vars(req)["url"]
	redditVideo, err := rvc.ret.GetByURL(req.Context(), url)
	if err != nil {
		respondErr(wr, err)
		return
	}

	respond(wr, http.StatusOK, redditVideo)
}

// TODO
func (rvc *redditVideosController) search(wr http.ResponseWriter, req *http.Request) {
	// vals := req.URL.Query()["t"]
	redditVideos, err := rvc.ret.Search(req.Context(), 10)
	if err != nil {
		respondErr(wr, err)
		return
	}

	respond(wr, http.StatusOK, redditVideos)
}

type redditConstructor interface {
	Create(ctx context.Context, redditVideo *domain.RedditVideo) (*domain.RedditVideo, error)
	Push(ctx context.Context, redditVideo *domain.RedditVideo) error
}

type redditDestructor interface {
	Delete(ctx context.Context, id bson.ObjectId) (*domain.RedditVideo, error)
	Pop(ctx context.Context) (*domain.RedditVideo, error)
}

// TODO: Search
type redditRetriever interface {
	GetByID(ctx context.Context, id bson.ObjectId) (*domain.RedditVideo, error)
	GetByURL(ctx context.Context, url string) (*domain.RedditVideo, error)
	Search(ctx context.Context, limit int) ([]domain.RedditVideo, error)
}
