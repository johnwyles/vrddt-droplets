package rest

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// addVrddtVideosAPI will register the various routes and their methods
func addVrddtVideosAPI(router *mux.Router, cons vrddtConstructor, des vrddtDestructor, ret vrddtRetriever, lg logger.Logger) {
	vvc := &vrddtVideosController{
		Logger: lg,

		cons: cons,
		des:  des,
		ret:  ret,
	}

	// TODO: Implement search / ALL
	vvrouter := router.PathPrefix("/vrddt_videos").Subrouter()

	vvrouter.HandleFunc("/{id}", vvc.getByID).Methods(http.MethodGet)
	vvrouter.HandleFunc("/{md5}", vvc.getByMD5).Methods(http.MethodGet)

	// vvrouter.HandleFun("/", vvc.create).Methods(http.MethodPost)
	// vvrouter.HandleFunc("/", vvc.search).Methods(http.MethodGet)
	// vvrouter.HandleFunc("/{id}", vvc.delete).Methods(http.MethodDelete)
	// vvrouter.HandleFunc("/", vvc.create).Methods(http.MethodPost)
}

type vrddtVideosController struct {
	logger.Logger

	cons vrddtConstructor
	des  vrddtDestructor
	ret  vrddtRetriever
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
	Create(ctx context.Context, vrddtVideo *domain.VrddtVideo) (*domain.VrddtVideo, error)
}

type vrddtDestructor interface {
	Delete(ctx context.Context, id bson.ObjectId) (*domain.VrddtVideo, error)
}

type vrddtRetriever interface {
	GetByID(ctx context.Context, id bson.ObjectId) (*domain.VrddtVideo, error)
	GetByMD5(ctx context.Context, md5 string) (*domain.VrddtVideo, error)
	Search(ctx context.Context, limit int) ([]domain.VrddtVideo, error)
}
