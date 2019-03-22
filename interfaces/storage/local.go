package storage

import (
	"context"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// TODO: Finish this

// gcs contains all the information about a GCS client
type local struct {
	log  logger.Logger
	path string
}

// Local initiates a new local storage connection
func Local(cfg *config.StorageLocalConfig, loggerHandle logger.Logger) (stg Storage, err error) {
	loggerHandle.Debugf("Local(cfg): %#v", cfg)

	stg = &local{
		log:  loggerHandle,
		path: cfg.Path,
	}

	return
}

// Attributes returns attributes about a file
func (l *local) Attributes(ctx context.Context, remotePath string) (attributes interface{}, err error) {
	return
}

// Cleanup closes the session
func (l *local) Cleanup(ctx context.Context) (err error) {
	return
}

// Delete will remove a file
func (l *local) Delete(ctx context.Context, remotePath string) (err error) {
	return
}

// GetLocation returns the URL to a file
func (l *local) GetLocation(ctx context.Context, remotePath string) (url string, err error) {
	return
}

// Init establishes the session
func (l *local) Init(ctx context.Context) (err error) {
	return
}

// List returns all files at a given path
func (l *local) List(ctx context.Context, remotePath string) (files []interface{}, err error) {
	return
}

// Download will download a remote path to the provided local path
func (l *local) Download(ctx context.Context, remotePath string, localPath string) (err error) {
	return
}

// Upload will upload a local path to the provided remote path
func (l *local) Upload(ctx context.Context, localPath string, remotePath string) (err error) {
	return
}
