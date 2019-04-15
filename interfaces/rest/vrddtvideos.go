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

type vrddtVideosController struct {
	logger.Logger

	cons vrddtConstructor
	des  vrddtDestructor
	ret  vrddtRetriever

	rcons redditConstructor
	rret  redditRetriever
}

// AddVrddtVideosAPI will register the various routes and their methods
func (c *Controller) AddVrddtVideosAPI(loggerHandle logger.Logger, cons vrddtConstructor, des vrddtDestructor, ret vrddtRetriever, rcons redditConstructor, rret redditRetriever) {
	vvc := &vrddtVideosController{
		Logger: loggerHandle,

		cons: cons,
		des:  des,
		ret:  ret,
		rret: rret,
	}

	// TODO: Implement search / ALL
	vvrouter := c.Router.PathPrefix("/vrddt_videos").Subrouter()

	vvrouter.HandleFunc("/{id}", vvc.getByID).Methods(http.MethodGet)
	vvrouter.HandleFunc("/{md5}", vvc.getByMD5).Methods(http.MethodGet)

	// vvrouter.HandleFun("/", vvc.create).Methods(http.MethodPost)
	// vvrouter.HandleFunc("/", vvc.search).Methods(http.MethodGet)
	// vvrouter.HandleFunc("/{id}", vvc.delete).Methods(http.MethodDelete)
	// vvrouter.HandleFunc("/", vvc.create).Methods(http.MethodPost)

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
	vvrouter.HandleFunc("/", vvc.getByRedditURL).Queries("url", "{url}").Methods(http.MethodGet)
}

// func (vvc *vrddtVideosController) create(wr http.ResponseWriter, req *http.Request) {
// 	vrddtVideo := &domain.VrddtVideo{}
// 	if err := readRequest(req, &vrddtVideo); err != nil {
// 		vvc.Warnf("failed to read vrddt video request: %s", err)
// 		respond(wr, http.StatusBadRequest, err)
// 		return
// 	}
//
// 	vrddtVideo, err := vvc.cons.Create(req.Context(), vrddtVideo)
// 	if err != nil {
// 		respondErr(wr, err)
// 		return
// 	}
//
// 	respond(wr, http.StatusCreated, vrddtVideo)
// }

// TODO: Delete only if other reddit videos aren't associated OR delete all
// associated reddit videos as well
// func (vvc *vrddtVideosController) delete(wr http.ResponseWriter, req *http.Request) {
// 	id := bson.ObjectIdHex(mux.Vars(req)["id"])
// 	vrddtVideo, err := vvc.des.Delete(req.Context(), id)
// 	if err != nil {
// 		respondErr(wr, err)
// 		return
// 	}
//
// 	respond(wr, http.StatusOK, vrddtVideo)
// }

func (vvc *vrddtVideosController) getByID(wr http.ResponseWriter, req *http.Request) {
	id := bson.ObjectIdHex(mux.Vars(req)["id"])
	vrddtVideo, err := vvc.ret.GetByID(req.Context(), id)
	if err != nil {
		respondErr(wr, err)
		return
	}

	respond(wr, http.StatusOK, vrddtVideo)
}

func (vvc *vrddtVideosController) getByMD5(wr http.ResponseWriter, req *http.Request) {
	md5 := mux.Vars(req)["md5"]
	vrddtVideo, err := vvc.ret.GetByMD5(req.Context(), md5)
	if err != nil {
		respondErr(wr, err)
		return
	}

	respond(wr, http.StatusOK, vrddtVideo)
}

// getByRedditURL will get the vrddt video by a query parameter for
// the URL from Reddit
func (vvc *vrddtVideosController) getByRedditURL(wr http.ResponseWriter, req *http.Request) {
	if url, ok := mux.Vars(req)["url"]; ok {
		finalURL, err := domain.GetFinalURL(url)
		if err != nil {
			respondErr(wr, err)
			return
		}

		redditVideo, err := vvc.rret.GetByURL(req.Context(), url)
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
			if !redditVideo.VrddtVideoID.Valid() {
				respondErr(wr, errors.InvalidValue("VrddtVideoID", "Invalid BSON ID"))
				return
			}

			vrddtVideo, errVrddt := vvc.ret.GetByID(context.TODO(), redditVideo.VrddtVideoID)
			if err != nil {
				switch errors.Type(errVrddt) {
				case errors.TypeResourceNotFound:
					vvc.Errorf("Reddit Video found (ID: %s) but vrddt Video (ID: %s) was not", redditVideo.ID.Hex(), redditVideo.VrddtVideoID.Hex())
				default:
					vvc.Errorf("Something went wrong: %s", errVrddt)
				}
			}

			vvc.Infof("Reddit video already in the database with ID '%s' and a vrddt video of: %#v", redditVideo.ID, vrddtVideo)
			respond(wr, http.StatusOK, vrddtVideo)
			return
		}

		redditVideo = domain.NewRedditVideo()
		redditVideo.URL = finalURL

		if err = vvc.rcons.Push(context.TODO(), redditVideo); err != nil {
			vvc.Errorf("Failed to push Reddit video to queue: %s", err)
		}

		vvc.Infof("Unique Reddit video URL queued with URL of: %s", redditVideo.URL)

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
				respondErr(wr, errors.OperationTimeout("Fetching Reddit URL from database", timeoutTime))
				vvc.Errorf("Operation timed out at after '%d' seconds.", timeoutTime)
				return
			case <-tick:
				// If the Reddit URL is not found in the database yet keep checking
				temporaryRedditVideo, err := vvc.rret.GetByURL(context.TODO(), redditVideo.URL)
				if err != nil {
					switch errors.Type(err) {
					default:
					case errors.TypeUnknown:
						vvc.Errorf("Something went wrong: %s", err)
					case errors.TypeResourceNotFound:
						continue
					}
				}
				vvc.Debugf("Reddit video now exists in db with a vrddt video ID: reddit video: %#v | vrddt video ID: %s", temporaryRedditVideo, temporaryRedditVideo.VrddtVideoID.Hex())

				vrddtVideo, errVrddt := vvc.ret.GetByID(context.TODO(), temporaryRedditVideo.VrddtVideoID)
				if errVrddt != nil {
					switch errors.Type(errVrddt) {
					case errors.TypeResourceNotFound:
						vvc.Errorf("Reddit video found (ID: %s) but associated vrddt video (ID: %s) was not", temporaryRedditVideo.ID.Hex(), temporaryRedditVideo.VrddtVideoID.Hex())
					default:
						vvc.Errorf("Something went wrong: %s", errVrddt)
					}
				}

				vvc.Infof("Unique URL created new vrddt video: %#v", vrddtVideo)

				respond(wr, http.StatusOK, vrddtVideo)

				return
			}
		}
	}

	respondErr(wr, errors.MissingField("url"))

	return
}

// TODO: Implement
// func (vvc *vrddtVideosController) search(wr http.ResponseWriter, req *http.Request) {
// 	// vals := req.URL.Query()["t"]
// 	vrddtVideos, err := vvc.ret.Search(req.Context(), 10)
// 	if err != nil {
// 		respondErr(wr, err)
// 		return
// 	}
//
// 	respond(wr, http.StatusOK, vrddtVideos)
// }

type vrddtConstructor interface {
	Create(ctx context.Context, vrddtVideo *domain.VrddtVideo) (err error)
}

type vrddtDestructor interface {
	Delete(ctx context.Context, id bson.ObjectId) (err error)
}

type vrddtRetriever interface {
	GetByID(ctx context.Context, id bson.ObjectId) (vrddtVideo *domain.VrddtVideo, err error)
	GetByMD5(ctx context.Context, md5 string) (vrddtVideo *domain.VrddtVideo, err error)
	Search(ctx context.Context, selector store.Selector, limit int) (vrddtVideos []*domain.VrddtVideo, err error)
}
