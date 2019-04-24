package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// redditVideosController holds all of the internal implementations of our
// usecases
type redditVideosController struct {
	cons redditConstructor
	des  redditDestructor
	log  logger.Logger
	ret  redditRetriever
}

// AddRedditVideosAPI will register the various routes and their methods
func AddRedditVideosAPI(loggerHandle logger.Logger, router *mux.Router, cons redditConstructor, des redditDestructor, ret redditRetriever) {
	rvc := &redditVideosController{
		log: loggerHandle,

		cons: cons,
		des:  des,
		ret:  ret,
	}

	// TODO: Implement search / ALL
	rvrouter := router.PathPrefix("/reddit_videos").Subrouter()

	// rvrouter.HandleFunc("/", rvrouter.create).Methods(http.MethodPost)

	// These will handle API calls to the internal queue
	// TODO: Needs auth
	// rvrouter.HandleFunc("/queue", rvc.enqueue).Methods(http.MethodPost)
	// rvrouter.HandleFunc("/queue", rvc.dequeue).Methods(http.MethodGet)

	// These will handle paths that match an ID for a Reddit video
	rvrouter.HandleFunc("/{id:[0-9a-fA-F]+}", rvc.getByID).Methods(http.MethodGet)
	rvrouter.HandleFunc("/{id:[0-9a-fA-F]+}/vrddt_video", rvc.getVrddtVideoByID).Methods(http.MethodGet)

	rvrouter.HandleFunc("/", rvc.getByRedditURL).Queries("url", "{url}").Methods(http.MethodGet)

	// rvrouter.HandleFunc("/search", rvc.search).Methods(http.MethodGet)
}

// // TODO: API for interacting with the queue
// func (rvc *redditVideosController) enqueue(wr http.ResponseWriter, req *http.Request) {
// 	redditVideo := domain.NewRedditVideo()
//
// 	if err := readRequest(req, &redditVideo); err != nil {
// 		rvc.Warnf("failed to read reddit video request: %s", err)
// 		respond(wr, http.StatusBadRequest, err)
// 		return
// 	}
//
// 	dbRedditVideo, err := rvc.ret.GetByURL(req.Context(), redditVideo.URL)
// 	if err != nil {
// 		switch errors.Type(err) {
// 		case errors.TypeUnknown:
// 			rvc.Warnf("error getting URL from db: %s", err)
// 			respondErr(wr, err)
// 			return
// 		case errors.TypeResourceNotFound:
// 		}
// 	}
//
// 	if dbRedditVideo != nil {
// 		err := fmt.Errorf("reddit video already exists in db with id '%s' for Reddit URL: %s", dbRedditVideo.ID, redditVideo.URL)
// 		rvc.Debugf("reddit video already exists: %s", err)
// 		respondErr(wr, err)
// 		return
// 	}
//
// 	if err := rvc.cons.Push(req.Context(), redditVideo); err != nil {
// 		rvc.Warnf("failed to create reddit video: %s", err)
// 		respondErr(wr, err)
// 		return
// 	}
//
// 	rvc.Infof("reddit video queued with id '%s'", redditVideo.ID)
// 	respond(wr, http.StatusCreated, redditVideo)
// }
//
// // TODO: Implement if queue is empty to simply return as such
// func (rvc *redditVideosController) dequeue(wr http.ResponseWriter, req *http.Request) {
// 	redditVideo, err := rvc.des.Pop(req.Context())
// 	if err != nil {
// 		rvc.Warnf("failed to create reddit video: %s", err)
// 		respondErr(wr, err)
// 		return
// 	}
//
// 	rvc.Infof("reddit video dequeued with id '%s'", redditVideo.ID)
// 	respond(wr, http.StatusCreated, redditVideo)
// }

// TODO: Delete vrddt video if no other reddit videos are associated
func (rvc *redditVideosController) delete(wr http.ResponseWriter, req *http.Request) {
	if id, ok := mux.Vars(req)["id"]; ok {
		bsonID := bson.ObjectIdHex(id)
		if err := rvc.des.Delete(req.Context(), bsonID); err != nil {
			respondErr(wr, err)
			return
		}

		respond(wr, http.StatusOK, id)
		return
	}

	respondErr(wr, errors.MissingField("id"))

	return
}

// getByID will get the Reddit video by ID
func (rvc *redditVideosController) getByID(wr http.ResponseWriter, req *http.Request) {
	if id, ok := mux.Vars(req)["id"]; ok {
		bsonID := bson.ObjectIdHex(id)
		redditVideo, err := rvc.ret.GetByID(req.Context(), bsonID)
		if err != nil {
			respondErr(wr, err)
			return
		}

		respond(wr, http.StatusOK, redditVideo)

		return
	}

	respondErr(wr, errors.MissingField("id"))

	return
}

// getByRedditURL will get the vrddt video by a query parameter for
// the URL from Reddit
func (rvc *redditVideosController) getByRedditURL(wr http.ResponseWriter, req *http.Request) {
	if url, ok := mux.Vars(req)["url"]; ok {
		finalURL, err := domain.GetFinalURL(url)
		if err != nil {
			respondErr(wr, err)
			return
		}

		redditVideo, err := rvc.ret.GetByURL(req.Context(), url)
		if err != nil {
			switch errors.Type(err) {
			case errors.TypeUnknown:
				respondErr(wr, errors.InvalidValue("url", url))
				return
			case errors.TypeResourceNotFound:
			default:
			}
		}

		if redditVideo != nil {
			rvc.log.Infof("Reddit video already in the database with ID '%s': %#v", redditVideo.ID, redditVideo)
			respond(wr, http.StatusOK, redditVideo)
			return
		}

		redditVideo = domain.NewRedditVideo()
		redditVideo.URL = finalURL

		if err = rvc.cons.Push(context.TODO(), redditVideo); err != nil {
			rvc.log.Errorf("Failed to push Reddit video to queue: %s", err)
		}

		rvc.log.Infof("Unique Reddit video URL queued with URL of: %s", redditVideo.URL)

		var pollTime int
		pollTime = 500
		if pollTime > 5000 {
			pollTime = 5000
		}
		if pollTime < 10 {
			pollTime = 10
		}

		timeoutTime := 60
		if timeoutTime > 600 {
			pollTime = 600
		}
		if timeoutTime < 1 {
			timeoutTime = 1
		}

		// Wait a pre-determined amount of time for the worker to fetch, convert,
		// store in the database, and store in storage the video
		timeout := time.After(
			time.Duration(
				time.Duration(timeoutTime) * time.Second,
			),
		)
		tick := time.Tick(time.Duration(pollTime) * time.Millisecond)
		for {
			select {
			case <-timeout:
				respondErr(wr, errors.ConnectionTimeout("database", timeoutTime))
				rvc.log.Errorf("Operation timed out at after '%d' seconds.", timeoutTime)
				return
			case <-tick:
				// If the Reddit URL is not found in the database yet keep checking
				redditVideo, err = rvc.ret.GetByURL(context.TODO(), redditVideo.URL)
				if err != nil {
					switch errors.Type(err) {
					default:
					case errors.TypeUnknown:
						rvc.log.Errorf("Something went wrong: %s", err)
					case errors.TypeResourceNotFound:
						continue
					}
				}

				rvc.log.Infof("Unique URL created new video: %#v", redditVideo)

				respond(wr, http.StatusOK, redditVideo)

				return
			}
		}
	}

	respondErr(wr, errors.MissingField("url"))

	return
}

// getVrddtVideoByID will get a vrddt video by the Reddit video ID
func (rvc *redditVideosController) getVrddtVideoByID(wr http.ResponseWriter, req *http.Request) {
	if id, ok := mux.Vars(req)["id"]; ok {
		bsonID := bson.ObjectId(id)
		redditVideo, err := rvc.ret.GetByID(req.Context(), bsonID)
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
		return
	}

	respondErr(wr, errors.MissingField("id"))

	return
}

// TODO: This is incomplete
// func (rvc *redditVideosController) search(wr http.ResponseWriter, req *http.Request) {
// 	vals := req.URL.Query()
//
// 	if vals == emptyVals {
// 		respondErr(wr, errors.MissingField("no fields were given to search by"))
// 		return
// 	}
//
// 	var foo []string
// 	redditVideos, err := rvc.ret.Search(req.Context(), foo, 10)
// 	if err != nil {
// 		respondErr(wr, err)
// 		return
// 	}
//
// 	respond(wr, http.StatusOK, redditVideos)
// 	return
// }

type redditConstructor interface {
	Create(ctx context.Context, redditVideo *domain.RedditVideo) (err error)
	Push(ctx context.Context, redditVideo *domain.RedditVideo) (err error)
}

type redditDestructor interface {
	Delete(ctx context.Context, id bson.ObjectId) (err error)
	Pop(ctx context.Context) (redditVideo *domain.RedditVideo, err error)
}

// TODO: Search
type redditRetriever interface {
	GetByID(ctx context.Context, id bson.ObjectId) (redditVideo *domain.RedditVideo, err error)
	GetByURL(ctx context.Context, url string) (redditVideo *domain.RedditVideo, err error)
	GetVrddtVideoByID(ctx context.Context, id bson.ObjectId) (rredditVideov *domain.VrddtVideo, err error)
	Search(ctx context.Context, selector store.Selector, limit int) (redditVideos []*domain.RedditVideo, err error)
}
