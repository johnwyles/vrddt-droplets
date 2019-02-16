package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

func addRedditVideosAPI(router *mux.Router, cons redditConstructor, des redditDestructor, ret redditRetriever, logger logger.Logger) {
	rvc := &redditVideosController{
		Logger: logger,
		cons:   cons,
		des:    des,
		ret:    ret,
	}

	// TODO: Implement search / ALL
	router.HandleFunc("/reddit_videos/queue", rvc.enqueue).Methods(http.MethodPost)
	router.HandleFunc("/reddit_videos/queue", rvc.dequeue).Methods(http.MethodGet)
	// router.HandleFunc("/reddit_videos/query/{query}", rvc.search).Methods(http.MethodGet)
	router.HandleFunc("/reddit_videos/{id}", rvc.getByID).Methods(http.MethodGet)
	router.HandleFunc("/reddit_videos/{id}/vrddt_video", rvc.getVrddtVideo).Methods(http.MethodGet)
	router.HandleFunc("/reddit_videos/{url}", rvc.getByURL).Methods(http.MethodGet)
	// router.HandleFunc("/reddit_videos/", rvc.search).Methods(http.MethodGet)
}

type redditVideosController struct {
	logger.Logger
	cons redditConstructor
	des  redditDestructor
	ret  redditRetriever
}

func (rvc *redditVideosController) enqueue(wr http.ResponseWriter, req *http.Request) {
	redditVideo := &domain.RedditVideo{
		Meta: domain.Meta{
			ID:        bson.NewObjectId(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	if err := readRequest(req, &redditVideo); err != nil {
		rvc.Warnf("failed to read reddit video request: %s", err)
		respond(wr, http.StatusBadRequest, err)
		return
	}

	dbRedditVideo, err := rvc.ret.GetByURL(req.Context(), redditVideo.URL)
	if err != nil {
		switch errors.Type(err) {
		case errors.TypeUnknown:
			rvc.Warnf("error getting URL from db: %s", err)
			respondErr(wr, err)
			return
		case errors.TypeResourceNotFound:
		}
	}

	if dbRedditVideo != nil {
		rvc.Infof("reddit video already found in db with id '%s'", dbRedditVideo.ID)
		respond(wr, http.StatusCreated, dbRedditVideo)
		return
	}

	if err := rvc.cons.Push(req.Context(), redditVideo); err != nil {
		rvc.Warnf("failed to create reddit video: %s", err)
		respondErr(wr, err)
		return
	}

	rvc.Infof("reddit video queued with id '%s'", redditVideo.ID)
	respond(wr, http.StatusCreated, redditVideo)
}

// TODO: Implement if queue is empty to simply return as such
func (rvc *redditVideosController) dequeue(wr http.ResponseWriter, req *http.Request) {
	redditVideo, err := rvc.des.Pop(req.Context())
	if err != nil {
		rvc.Warnf("failed to create reddit video: %s", err)
		respondErr(wr, err)
		return
	}

	rvc.Infof("reddit video dequeued with id '%s'", redditVideo.ID)
	respond(wr, http.StatusCreated, redditVideo)
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
	if err != nil {
		respondErr(wr, err)
		return
	}

	respond(wr, http.StatusOK, redditVideo)
}

func (rvc *redditVideosController) getVrddtVideo(wr http.ResponseWriter, req *http.Request) {
	id := bson.ObjectIdHex(mux.Vars(req)["id"])
	redditVideo, err := rvc.ret.GetByID(req.Context(), id)
	if err != nil {
		respondErr(wr, err)
		return
	}

	vrddtVideo, err := rvc.ret.GetVrddtVideoByID(req.Context(), redditVideo.VrddtVideoID)
	if err != nil {
		respondErr(wr, err)
		return
	}

	respond(wr, http.StatusOK, vrddtVideo)
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
	GetVrddtVideoByID(ctx context.Context, id bson.ObjectId) (*domain.VrddtVideo, error)
	Search(ctx context.Context, limit int) ([]domain.RedditVideo, error)
}
