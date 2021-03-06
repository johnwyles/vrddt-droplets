package web

import (
	"net/http"
	"os"

	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// safeFileSystem implements http.FileSystem. It is used to prevent directory
// listing of static assets.
type safeFileSystem struct {
	logger.Logger

	fs http.FileSystem
}

func newSafeFileSystemServer(loggerHandle logger.Logger, root string) http.Handler {
	sfs := &safeFileSystem{
		Logger: loggerHandle,

		fs: http.Dir(root),
	}
	return http.FileServer(sfs)
}

func (sfs safeFileSystem) Open(path string) (http.File, error) {
	f, err := sfs.fs.Open(path)
	if err != nil {
		sfs.Warnf("Failed to open file '%s': %v", path, err)
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		sfs.Warnf("Path '%s' is a directory, rejecting static path request", path)
		return nil, os.ErrNotExist
	}

	return f, nil
}
