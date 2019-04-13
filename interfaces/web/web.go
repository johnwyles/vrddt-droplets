package web

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// Controller holds the information about the web controller
type Controller struct {
	Router *mux.Router
}

// New initializes a new webapp server.
func New(loggerHandle logger.Logger) (controller *Controller, err error) {
	controller = &Controller{
		Router: mux.NewRouter(),
	}

	tpl, err := initTemplate(loggerHandle, "", "web/templates")
	if err != nil {
		return
	}

	app := &app{
		render: func(wr http.ResponseWriter, tplName string, data interface{}) {
			if err := tpl.ExecuteTemplate(wr, tplName, data); err != nil {
				loggerHandle.Errorf("Failed to render template '%s': %+v", tplName, err)
			}
		},
	}

	// Static file serving
	fsServer := newSafeFileSystemServer(loggerHandle, "web/static")
	controller.Router.PathPrefix("/static").Handler(http.StripPrefix("/static", fsServer))
	controller.Router.Handle("/favicon.ico", fsServer)

	// Root route
	controller.Router.HandleFunc("/", app.indexHandler)

	return
}

func initTemplate(loggerHandle logger.Logger, name string, path string) (tpl *template.Template, err error) {
	apath, err := filepath.Abs(path)
	if err != nil {
		return
	}

	files, err := ioutil.ReadDir(apath)
	if err != nil {
		return
	}

	loggerHandle.Infof("Loading templates from '%s'...", path)
	tpl = template.New(name)
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		fp := filepath.Join(apath, f.Name())
		loggerHandle.Debugf("Parsing template file '%s'", f.Name())
		tpl.New(f.Name()).ParseFiles(fp)
	}

	return
}
