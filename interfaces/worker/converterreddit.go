package worker

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"

	mgo "gopkg.in/mgo.v2"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/store"
)

// TODO: Turn const into configuration variables

const (
	// OutputFileExtension is the filename extension for the file we output
	OutputFileExtension = ".mp4"

	// TemporaryDirectoryPrefix is the prefix for the directory which will
	// store the temporarily converted video before uploading
	TemporaryDirectoryPrefix = "vrddt-worker-downloader"

	// TemporaryFilePrefix is the prefix for the file name which will
	// store the temporarily converted video before uploading
	TemporaryFilePrefix = "vrddt-worker-downloader-video"
)

// TODO: Fix comments

// checkIfVrddtMD5Exists will look to see if the processed video from the
// unique Reddit URL that was given matches a vrddt video we have already
// stored and if so make the association
func (w *workerConverter) checkIfVrddtMD5Exists(ctx context.Context, outputMD5Sum []byte, redditVideo *domain.RedditVideo) (exists bool, err error) {
	// Check the hash of the file against what is in the DB and only
	// add it to the DB if it is unique otherwise associate it with the
	// existing vrddt video
	temporaryVrddtVideo, err := w.store.GetVrddtVideo(
		ctx,
		store.Selector{
			"md5": outputMD5Sum,
		},
	)
	switch err {
	case nil:
		redditVideo.VrddtVideoID = temporaryVrddtVideo.ID
		if err = w.store.CreateRedditVideo(ctx, redditVideo); err != nil {
			return
		}
		exists = true
		return
	case mgo.ErrNotFound:
		// We do not care if a vrddt video was not found
		err = nil
	default:
	}

	return
}

// checkIfRedditURLExists will look in the database to see if the Reddit URL
// already exists or not.  If it does exist it will return true otherwise false
func (w *workerConverter) checkIfRedditURLExists(ctx context.Context, redditVideo *domain.RedditVideo) (exists bool, err error) {
	// Let's also see if the Reddit URL has been seen before
	_, err = w.store.GetRedditVideo(
		ctx,
		store.Selector{
			"url": redditVideo.URL,
		},
	)
	switch err {
	case nil:
		exists = true
	case mgo.ErrNotFound:
		// We do not care if a Reddit video was not found
		err = nil
	default:
		return
	}

	return
}

// doWorkReditVideo will perform all of the steps for a video conversion for a
// Reddit video, store a reference of it in the store, and upload the result
// to storage
func (w *workerConverter) doWorkRedditVideo(ctx context.Context) (err error) {
	if _, ok := w.workItem.(*domain.RedditVideo); !ok {
		return fmt.Errorf("work item is not a reddit video: %#v", w.workItem)
	}

	redditVideo := domain.NewRedditVideo()
	if w.workItem.(*domain.RedditVideo).URL != "" {
		redditVideo.URL = w.workItem.(*domain.RedditVideo).URL
	} else {
		return fmt.Errorf("work item was a reddit video but did not have the URL field set")
	}

	// We shouldn't need this if all the entries to the queue are done
	if err = redditVideo.SetFinalURL(); err != nil {
		return
	}

	urlExists, err := w.checkIfRedditURLExists(ctx, redditVideo)
	if err != nil {
		return
	} else if urlExists {
		w.log.Infof("Reddit URL already exists in the database: %s", redditVideo.URL)
		return
	}
	w.log.Debugf("Reddit URL is unique and does not exist in the database: %s", redditVideo.URL)

	// Set the AudioURL, VideoURL, and Title
	if err = redditVideo.SetMetadata(); err != nil {
		return
	}

	// I am not sure that Reddit does this but it could save them some
	// trouble (and wouldn't be needed here if so). However, if someone
	// uploads the same video to multiple different subreddits and
	// Reddit notices the content is the same and points all references
	// back to the same URL this will catch those instances and save us
	// some work
	temporaryRedditVideo, err := w.store.GetRedditVideo(
		ctx,
		store.Selector{
			"audio_url": redditVideo.AudioURL,
			"video_url": redditVideo.VideoURL,
		},
	)
	switch err {
	case nil:
		redditVideo.VrddtVideoID = temporaryRedditVideo.VrddtVideoID
		if createErr := w.store.CreateRedditVideo(ctx, redditVideo); createErr != nil {
			return createErr
		}
	case mgo.ErrNotFound:
		// This simply means a duplicate was not found in the database (i.e.
		// we have a unique Reddit URL)
		w.log.Debugf("Reddit audio and/or video URLs is unique: %s", redditVideo.URL)
	default:
		// Something unexpected happened
		return
	}

	if err = redditVideo.Download(); err != nil {
		return
	}

	w.log.Debugf("Downloaded Reddit video: %#v", redditVideo)

	// We don't care if the Audio file fails to download as there are
	// plenty of videos on Reddit that do not have audio
	if redditVideo.RedditAudio != nil && redditVideo.RedditAudio.FileHandle != nil && redditVideo.RedditAudio.FilePath != "" {
		defer redditVideo.RedditAudio.FileHandle.Close()
		defer os.Remove(redditVideo.RedditAudio.FilePath)
	} else {
		redditVideo.RedditAudio = &domain.RedditAudio{
			FileHandle: nil,
			FilePath:   "",
		}
	}

	defer redditVideo.FileHandle.Close()
	defer os.Remove(redditVideo.FilePath)

	w.log.Infof("Converting media for Reddit URL: %s", redditVideo.URL)

	temporaryOutputFileHandle, err := w.convertVideo(redditVideo.FilePath, redditVideo.RedditAudio.FilePath)

	defer temporaryOutputFileHandle.Close()
	defer os.Remove(temporaryOutputFileHandle.Name())

	// Get an MD5 hash of the converted file
	outputMD5 := md5.New()
	if _, err = io.Copy(outputMD5, temporaryOutputFileHandle); err != nil {
		return
	}
	outputMD5Sum := outputMD5.Sum(nil)

	md5Exists, err := w.checkIfVrddtMD5Exists(ctx, outputMD5Sum, redditVideo)
	if err != nil {
		return
	} else if md5Exists {
		w.log.Debugf("Vrddt MD5 already exists in the database")
		return
	}
	w.log.Debugf("MD5 for the resulting vrddt video does not exist in the database")

	// The vrddt video is unique so setup a new one and assign the hash
	vrddtVideo := domain.NewVrddtVideo()
	vrddtVideo.MD5 = outputMD5Sum

	w.log.Debugf("Uploading media to storage for Reddit URL: %s", redditVideo.URL)

	destinationFilename := vrddtVideo.ID.Hex() + OutputFileExtension
	if err = w.storage.Upload(ctx, temporaryOutputFileHandle.Name(), destinationFilename); err != nil {
		return
	}
	vrddtVideo.URL, err = w.storage.GetLocation(ctx, destinationFilename)
	if err != nil {
		w.storage.Delete(ctx, destinationFilename)
		return
	}

	w.log.Debugf("Vrddt media uploaded to storage as URL: %s", vrddtVideo.URL)

	// Save the vrddt video information to the database
	err = w.store.CreateVrddtVideo(ctx, vrddtVideo)
	if err != nil {
		w.storage.Delete(ctx, destinationFilename)
		return
	}

	// If we got this far then the Reddit URL is unique and either:
	// A) A vrddt video was found that already has processed the Reddit
	// audio and video before so create the association
	// B) This link, video,audio, and the generated vrddt video are all
	// unique so store all of these values
	redditVideo.VrddtVideoID = vrddtVideo.ID
	if err = w.store.CreateRedditVideo(ctx, redditVideo); err != nil {
		w.store.DeleteVrddtVideo(
			ctx,
			store.Selector{
				"_id": vrddtVideo.ID,
			},
		)
		w.storage.Delete(ctx, destinationFilename)
		return
	}

	w.log.Infof("Completed storing media [VrddtVideo URL: %s] for Reddit URL: %s",
		vrddtVideo.URL,
		redditVideo.URL,
	)

	return
}
