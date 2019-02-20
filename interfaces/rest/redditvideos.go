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

// addRedditVideosAPI will register the various routes and their methods
func addRedditVideosAPI(router *mux.Router, cons redditConstructor, des redditDestructor, ret redditRetriever, logger logger.Logger) {
	rvc := &redditVideosController{
		Logger: logger,

		cons: cons,
		des:  des,
		ret:  ret,
	}

	// TODO: Implement search / ALL
	rvrouter := router.PathPrefix("/reddit_videos").Subrouter()

	// These will handle API calls to the internal queue
	// TODO: Needs auth
	// rvrouter.HandleFunc("/queue", rvc.enqueue).Methods(http.MethodPost)
	// rvrouter.HandleFunc("/queue", rvc.dequeue).Methods(http.MethodGet)

	// These will handle paths that match an ID for a Reddit video
	rvrouter.HandleFunc("/{id:[0-9a-fA-F]+}", rvc.getByID).Methods(http.MethodGet)
	rvrouter.HandleFunc("/{id:[0-9a-fA-F]+}/vrddt_video", rvc.getVrddtVideoByID).Methods(http.MethodGet)

	// If we pass the query parameter "url" in we will return a vrddt video URL
	// to content generated. The follow scenarios can occur:
	//   - Check the database to see if this reddit URL has been seen before
	//     and if it has return the vrddt video URL for the content we created
	//     previously
	//   - Database does not have the Reddit URL so it is unique and will
	//     enqueue the URL in the queue
	//     - A worker will perform all the conversion and update the database
	//     - This process will continuously check the database for Reddit URL to
	//       appear and return the generated vrddt video URL
	//   - An invalid URL or error was supplied so return an error response
	rvrouter.HandleFunc("/", rvc.getVrddtVideoByURL).Queries("url", "{url}").Methods(http.MethodGet)

	// rvrouter.HandleFunc("/search", rvc.search).Methods(http.MethodGet)
}

// redditVideosController holds all of the internal implementations of our
// usecases
type redditVideosController struct {
	logger.Logger

	cons redditConstructor
	des  redditDestructor
	ret  redditRetriever
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
		redditVideo, err := rvc.des.Delete(req.Context(), bsonID)
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

func (rvc *redditVideosController) getVrddtVideoByURL(wr http.ResponseWriter, req *http.Request) {
	if url, ok := mux.Vars(req)["url"]; ok {
		finalURL, err := domain.GetFinalURL(url)
		if err != nil {
			respondErr(wr, err)
			return
		}

		redditVideo, err := rvc.ret.GetByURL(req.Context(), url)
		if err != nil {
			switch errors.Type(err) {
			default:
			case errors.TypeUnknown:
				respondErr(wr, errors.InvalidValue("url", url))
				return
			case errors.TypeResourceNotFound:
			}
		}

		if redditVideo != nil {
			if !redditVideo.VrddtVideoID.Valid() {
				respondErr(wr, errors.InvalidValue("VrddtVideoID", "invalid bson id"))
				return
			}

			vrddtVideo, errVrddt := rvc.ret.GetVrddtVideoByID(context.TODO(), redditVideo.VrddtVideoID)
			if err != nil {
				switch errors.Type(errVrddt) {
				case errors.TypeResourceNotFound:
					rvc.Fatalf("Reddit Video found (ID: %s) but vrddt Video (ID: %s) was not", redditVideo.ID.Hex(), redditVideo.VrddtVideoID.Hex())
				default:
					rvc.Fatalf("something went wrong: %s", errVrddt)
				}
			}

			rvc.Infof("reddit video already in db with id '%s' and a vrddt video of: %#v", redditVideo.ID, vrddtVideo)
			respond(wr, http.StatusOK, vrddtVideo)
			return
		}

		redditVideo = domain.NewRedditVideo()
		redditVideo.URL = finalURL

		if err = rvc.cons.Push(context.TODO(), redditVideo); err != nil {
			rvc.Fatalf("failed to push reddit video to queue: %s", err)
		}

		rvc.Infof("unique reddit video URL queued with URL of: %s", redditVideo.URL)

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
				rvc.Fatalf("operation timed out at after '%d' seconds.", timeout)
			case <-tick:
				// If the Reddit URL is not found in the database yet keep checking
				temporaryRedditVideo, err := rvc.ret.GetByURL(context.TODO(), redditVideo.URL)
				if err != nil {
					switch errors.Type(err) {
					default:
					case errors.TypeUnknown:
						rvc.Fatalf("something went wrong: %s", err)
					case errors.TypeResourceNotFound:
						continue
					}
				}
				rvc.Debugf("reddit video now exists in db with a vrddt video ID: reddit video: %#v | vrddt video ID: %s", temporaryRedditVideo, temporaryRedditVideo.VrddtVideoID.Hex())

				vrddtVideo, errVrddt := rvc.ret.GetVrddtVideoByID(context.TODO(), temporaryRedditVideo.VrddtVideoID)
				if errVrddt != nil {
					switch errors.Type(errVrddt) {
					case errors.TypeResourceNotFound:
						rvc.Fatalf("Reddit video found (ID: %s) but associated vrddt video (ID: %s) was not", temporaryRedditVideo.ID.Hex(), temporaryRedditVideo.VrddtVideoID.Hex())
					default:
						rvc.Fatalf("something went wrong: %s", errVrddt)
					}
				}

				rvc.Infof("unique url created new vrddt video: %#v", vrddtVideo)
				respond(wr, http.StatusOK, vrddtVideo)
				return
			}
		}
	}

	respondErr(wr, errors.MissingField("url"))

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
	Search(ctx context.Context, q []string, limit int) ([]domain.RedditVideo, error)
}
