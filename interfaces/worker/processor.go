package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
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
type processor struct {
	converter converter.Converter
	queue     queue.Queue
	log       logger.Logger
	store     store.Store
	storage   storage.Storage
	work      interface{}
}

// Processor will take a converter, queue, storage system, and persistence
// store to provide an initial struct
func Processor(cfg *config.WorkerProcessorConfig, loggerHandle logger.Logger, c converter.Converter, q queue.Queue, str store.Store, stg storage.Storage) (worker Worker, err error) {
	worker = &processor{
		converter: c,
		queue:     q,
		log:       loggerHandle,
		storage:   stg,
		store:     str,
		work:      nil,
	}

	loggerHandle.Debugf("Worker > Processor(cfg): %#v", worker)

	return
}

// SaveWork will complete the work
func (p *processor) CompleteWork(ctx context.Context) (err error) {
	p.work = nil
	return
}

// DoWork will perform the work
func (p *processor) DoWork(ctx context.Context) (err error) {
	if p.work == nil {
		return errors.MissingField("work")
	}

	// TODO: Implement other video types
	switch p.work.(type) {
	case *domain.RedditVideo:
		p.log.Debugf("Performing work on reddit video: %#v", p.work)
		return p.doWorkRedditVideo(ctx)
	case *domain.VrddtVideo:
		p.log.Debugf("Performing work on vrddt video: %#v", p.work)
		return fmt.Errorf("There is no work that can be performed on a vrddt video: %#v", p.work)
	case *domain.YoutubeVideo:
		p.log.Debugf("Performing work on youtube video: %#v", p.work)
		return fmt.Errorf("Working on youtube videos has not been implemented yet: %#v", p.work)
	default:
		p.log.Debugf("Performing work on unknown type: %#v", p.work)
		return fmt.Errorf("There is no work that can be performed on an unknown type: %#v", p.work)
	}
}

// GetWork will return some work to perform
func (p *processor) GetWork(ctx context.Context) (err error) {
	// Make sure we are marking ourselves as a consumer of the queue
	p.queue.MakeConsumer(ctx)

	// Get an element of work from the queue
	work, err := p.queue.Pop(ctx)
	if err != nil {
		return
	}

	// See if work is in JSON first
	if byteWork, ok := work.([]byte); ok {
		p.unmarshalJSON(byteWork)
	} else {
		// TODO: throw an error here or deal with it in DoWork()?
		p.log.Debugf("Work is not of type []byte (which we expect should be unmarshalled JSON): %#v", work)
		p.work = &work
	}

	p.log.Debugf("Popped work off of queue: %#v", p.work)

	return
}

func (p *processor) Init(ctx context.Context) (err error) {
	return
}

// convertVideo will do the ffmpeg bits of converting the video
func (p *processor) convertVideo(inputVideoFilePath string, inputAudioFilePath string) (temporaryOutputFile *os.File, err error) {
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
	if err = p.converter.Convert(ctx, inputVideoFilePath, inputAudioFilePath, temporaryOutputFile.Name()); err != nil {
		return
	}

	return
}

// unmarshalJSON will attempt to unmarshal a byte array into known structs
// TODO: Convert this to a type-switch?
func (p *processor) unmarshalJSON(byteWork []byte) (err error) {
	redditVideo := domain.RedditVideo{}
	if err := json.Unmarshal(byteWork, &redditVideo); err != nil {
		p.log.Debugf("Work is not a valid Reddit video: %s", err)
	} else {
		p.work = &redditVideo
		return nil
	}

	vrddtVideo := domain.VrddtVideo{}
	if err := json.Unmarshal(byteWork, &vrddtVideo); err != nil {
		p.log.Debugf("Work is not a valid Vrddt video: %s", err)
	} else {
		p.work = &vrddtVideo
		return nil
	}

	youtubeVideo := domain.YoutubeVideo{}
	if err := json.Unmarshal(byteWork, &youtubeVideo); err != nil {
		p.log.Debugf("Work is not a valid Youtube video: %s", err)
	} else {
		p.work = &youtubeVideo
		return nil
	}

	if err := json.Unmarshal(byteWork, &p.work); err != nil {
		p.work = nil
		p.log.Errorf("work is not JSON or cannot be unmarshalled correctly: %s", err)
		return err
	}

	return errors.InvalidValue("work", fmt.Sprintf("Unrecognized work item: %#v", p.work))
}
