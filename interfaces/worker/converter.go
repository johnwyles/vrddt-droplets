package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/interfaces/converter"
	"github.com/johnwyles/vrddt-droplets/interfaces/queue"
	"github.com/johnwyles/vrddt-droplets/interfaces/storage"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// converter holds all of the information about the worker for converting
// videos from a queue saving storage and any information to a persistence
// store
type workerConverter struct {
	converter converter.Converter
	queue     queue.Queue
	log       logger.Logger
	store     store.Store
	storage   storage.Storage
	workItem  interface{}
}

// Converter will take a converter, queue, storage system, and persistence
// store to provide an initial struct
func Converter(cfg *config.WorkerConverterConfig, loggerHandle logger.Logger, c converter.Converter, q queue.Queue, str store.Store, stg storage.Storage) (worker Worker, err error) {
	worker = &workerConverter{
		converter: c,
		queue:     q,
		log:       loggerHandle,
		storage:   stg,
		store:     str,
		workItem:  nil,
	}

	loggerHandle.Debugf("Worker > Converter(cfg): %#v", worker)

	return
}

// SaveWork will complete the work
func (w *workerConverter) CompleteWork(ctx context.Context) (err error) {
	w.workItem = nil
	return
}

// DoWork will perform the work
func (w *workerConverter) DoWork(ctx context.Context) (err error) {
	if w.workItem == nil {
		return fmt.Errorf("The was no element of work to perform converter work on")
	}

	switch w.workItem.(type) {
	case *domain.RedditVideo:
		w.log.Debugf("Performing work on reddit video: %#v", w.workItem)
		return w.doWorkRedditVideo(ctx)
	case *domain.VrddtVideo:
		w.log.Debugf("Performing work on vrddt video: %#v", w.workItem)
		return fmt.Errorf("There is no work that can be performed on a vrddt video: %#v", w.workItem)
	case *domain.YoutubeVideo:
		w.log.Debugf("Performing work on youtube video: %#v", w.workItem)
		return fmt.Errorf("Working on youtube videos has not been implemented yet: %#v", w.workItem)
	default:
		w.log.Debugf("Performing work on unknown type: %#v", w.workItem)
		return fmt.Errorf("There is no work that can be performed on an unknown type: %#v", w.workItem)
	}
}

// GetWork will return some work to perform
func (w *workerConverter) GetWork(ctx context.Context) (err error) {
	// Make sure we are marking ourselves as a consumer of the queue
	w.queue.MakeConsumer(ctx)

	// Get an element of work from the queue
	work, err := w.queue.Pop(ctx)
	if err != nil {
		return
	}

	// See if work is in JSON first
	if byteWork, ok := work.([]byte); ok {
		w.unmarshalJSON(byteWork)
	} else {
		// TODO: throw an error here or deal with it in DoWork()?
		w.log.Debugf("Work is not of type []byte (which we expect should be unmarshalled JSON): %#v", work)
		w.workItem = &work
	}

	w.log.Debugf("Popped work off of queue: %#v", w.workItem)

	return
}

func (w *workerConverter) Init(ctx context.Context) (err error) {
	return
}

// convertVideo will do the ffmpeg bits of converting the video
func (w *workerConverter) convertVideo(inputVideoFilePath string, inputAudioFilePath string) (temporaryOutputFile *os.File, err error) {
	// Setup our temporary output file
	temporaryDirectory, err := ioutil.TempDir(
		os.TempDir(),
		TemporaryDirectoryPrefix,
	)
	if err != nil {
		return
	}
	temporaryOutputFile, err = ioutil.TempFile(
		temporaryDirectory,
		TemporaryFilePrefix,
	)
	if err != nil {
		return
	}

	// Convert the downloaded files
	// TODO: Context
	ctx := context.TODO()
	if err = w.converter.Convert(ctx, inputVideoFilePath, inputAudioFilePath, temporaryOutputFile.Name()); err != nil {
		return
	}

	return
}

// unmarshalJSON will attempt to unmarshal a byte array into known structs
func (w *workerConverter) unmarshalJSON(byteWork []byte) (err error) {
	redditVideo := domain.RedditVideo{}
	if err := json.Unmarshal(byteWork, &redditVideo); err != nil {
		w.log.Debugf("Work is not a valid Reddit video: %s", err)
	} else {
		w.workItem = &redditVideo
		return nil
	}

	vrddtVideo := domain.VrddtVideo{}
	if err := json.Unmarshal(byteWork, &vrddtVideo); err != nil {
		w.log.Debugf("Work is not a valid Vrddt video: %s", err)
	} else {
		w.workItem = &vrddtVideo
		return nil
	}

	youtubeVideo := domain.YoutubeVideo{}
	if err := json.Unmarshal(byteWork, &youtubeVideo); err != nil {
		w.log.Debugf("Work is not a valid Youtube video: %s", err)
	} else {
		w.workItem = &youtubeVideo
		return nil
	}

	if err := json.Unmarshal(byteWork, &w.workItem); err != nil {
		w.workItem = nil
		w.log.Errorf("work is not JSON or cannot be unmarshalled correctly: %s", err)
		return err
	}

	return fmt.Errorf("Unrecognized work item: %#v", w.workItem)
}
