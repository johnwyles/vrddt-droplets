package storage

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// TODO: Figure out context

// gcs contains all the information about a GCS client
type gcs struct {
	bucket          *storage.BucketHandle
	bucketName      string
	client          *storage.Client
	context         context.Context
	credentialsJSON string
	log             logger.Logger
}

// GCS initiates a new GCS storage connection
func GCS(cfg *config.StorageGCSConfig, loggerHandle logger.Logger) (stg Storage, err error) {
	loggerHandle.Debugf("GCS(cfg): %#v", cfg)

	stg = &gcs{
		bucketName:      cfg.Bucket,
		credentialsJSON: cfg.CredentialsJSON,
		log:             loggerHandle,
	}

	return
}

// Attributes returns attributes about a file
func (g *gcs) Attributes(remotePath string) (attributes interface{}, err error) {
	file := g.bucket.Object(remotePath)
	attributes, err = file.Attrs(g.context)
	if err != nil {
		return
	}

	return
}

// Cleanup closes the GCS connection
func (g *gcs) Cleanup() (err error) {
	if g.client == nil {
		return fmt.Errorf("A client has not been set in order to be cleaned up")
	}

	return g.client.Close()
}

// Delete will remove a file
func (g *gcs) Delete(remotePath string) (err error) {
	if err = g.bucket.Object(remotePath).Delete(g.context); err != nil {
		return err
	}

	return
}

// GetLocation returns the URL to a file
func (g *gcs) GetLocation(remotePath string) (url string, err error) {
	attributes, err := g.bucket.Object(remotePath).Attrs(g.context)
	if err != nil {
		return
	}

	url = attributes.MediaLink

	return
}

// Init establishes the connection
func (g *gcs) Init() (err error) {
	g.context = context.Background()

	gcsClient, err := storage.NewClient(g.context, option.WithCredentialsFile(g.credentialsJSON))
	if err != nil {
		return
	}

	g.bucket = gcsClient.Bucket(g.bucketName)
	g.client = gcsClient

	return
}

// List returns all files at a given path
func (g *gcs) List(remotePath string) (files []interface{}, err error) {
	iter := g.bucket.Objects(g.context, nil)
	for {
		attributes, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		files = append(files, attributes.Name)
	}

	return
}

// Download will download a remote path to the provided local path
func (g *gcs) Download(remotePath string, localPath string) (err error) {
	fileReader, err := g.bucket.Object(remotePath).NewReader(g.context)
	if err != nil {
		return
	}
	defer fileReader.Close()

	data, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return
	}

	destinationFile, err := os.Create(localPath)
	if err != nil {
		return
	}
	defer destinationFile.Close()

	_, err = destinationFile.Write(data)
	if err != nil {
		return
	}

	return
}

// Upload will upload a local path to the provided remote path
func (g *gcs) Upload(localPath string, remotePath string) (err error) {
	gcsObject := g.bucket.Object(remotePath)
	gcsWriter := gcsObject.NewWriter(g.context)

	sourceFile, err := os.Open(localPath)
	if err != nil {
		return
	}
	defer sourceFile.Close()

	if _, err = io.Copy(gcsWriter, sourceFile); err != nil {
		return
	}

	if err = gcsWriter.Close(); err != nil {
		gcsObject.Delete(g.context)
		return
	}

	if err = gcsObject.ACL().Set(g.context, storage.AllUsers, storage.RoleReader); err != nil {
		gcsObject.Delete(g.context)
		return
	}

	return
}
