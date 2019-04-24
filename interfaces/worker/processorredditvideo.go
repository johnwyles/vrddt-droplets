package worker

import (
	"context"
	"crypto/md5"
	"fmt"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"io"
	"os"

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

// checkIfRedditURLExists will look in the database to see if the Reddit URL
// already exists or not.  If it does exist it will return true otherwise false
func (p *processor) checkIfRedditURLExists(ctx context.Context, redditVideo *domain.RedditVideo) (exists bool, err error) {
	// Let's also see if the Reddit URL has been seen before
	_, err = p.store.GetRedditVideo(
		ctx,
		store.Selector{
			"url": redditVideo.URL,
		},
	)
	if err != nil {
		switch errors.Type(err) {
		case errors.TypeResourceNotFound:
			// We do not care if a Reddit video was not found
			err = nil
		default:
			return
		}
	} else {
		exists = true
	}

	return
}

// checkIfVrddtMD5Exists will look to see if the processed video from the
// unique Reddit URL that was given matches a vrddt video we have already
// stored and if so make the association
func (p *processor) checkIfVrddtMD5Exists(ctx context.Context, outputMD5Sum []byte, redditVideo *domain.RedditVideo) (exists bool, err error) {
	// Check the hash of the file against what is in the DB and only
	// add it to the DB if it is unique otherwise associate it with the
	// existing vrddt video
	temporaryVrddtVideo, err := p.store.GetVrddtVideo(
		ctx,
		store.Selector{
			"md5": outputMD5Sum,
		},
	)
	if err != nil {
		switch errors.Type(err) {
		case errors.TypeResourceNotFound:
			// We do not care if a vrddt video was not found
			err = nil
		default:
			return
		}
	} else {
		redditVideo.VrddtVideoID = temporaryVrddtVideo.ID
		if err = p.store.CreateRedditVideo(ctx, redditVideo); err != nil {
			return
		}
		exists = true
	}

	return
}

// doWorkReditVideo will perform all of the steps for a video conversion for a
// Reddit video, store a reference of it in the store, and upload the result
// to storage
func (p *processor) doWorkRedditVideo(ctx context.Context) (err error) {
	if _, ok := p.work.(*domain.RedditVideo); !ok {
		return errors.InvalidValue("work", fmt.Sprintf("Work item is not a valid Reddit video: %#v", p.work))
	}

	redditVideo := domain.NewRedditVideo()
	if p.work.(*domain.RedditVideo).URL != "" {
		redditVideo.URL = p.work.(*domain.RedditVideo).URL
	} else {
		return errors.MissingField("url")
	}

	// We shouldn't need this if all the entries to the queue are done
	if err = redditVideo.SetFinalURL(); err != nil {
		return
	}

	urlExists, err := p.checkIfRedditURLExists(ctx, redditVideo)
	if err != nil {
		return
	} else if urlExists {
		p.log.Infof("Reddit URL already exists in the database: %s", redditVideo.URL)
		return
	}
	p.log.Debugf("Reddit URL is unique and does not exist in the database: %s", redditVideo.URL)

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
	temporaryRedditVideo, err := p.store.GetRedditVideo(
		ctx,
		store.Selector{
			"audio_url": redditVideo.AudioURL,
			"video_url": redditVideo.VideoURL,
		},
	)
	if err != nil {
		switch errors.Type(err) {
		case errors.TypeResourceNotFound:
			// This simply means a duplicate was not found in the database (i.e.
			// we have a unique Reddit URL)
			p.log.Debugf("Reddit audio and/or video URLs is unique: %s", redditVideo.URL)
		default:
			// Something unexpected happened
			return
		}
	} else {
		redditVideo.VrddtVideoID = temporaryRedditVideo.VrddtVideoID
		if createErr := p.store.CreateRedditVideo(ctx, redditVideo); createErr != nil {
			return createErr
		}
	}

	if err = redditVideo.Download(); err != nil {
		return
	}

	p.log.Debugf("Downloaded Reddit video: %#v", redditVideo)

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

	p.log.Infof("Converting media for Reddit URL: %s", redditVideo.URL)

	temporaryOutputFileHandle, err := p.convertVideo(redditVideo.FilePath, redditVideo.RedditAudio.FilePath)

	defer temporaryOutputFileHandle.Close()
	defer os.Remove(temporaryOutputFileHandle.Name())

	// Get an MD5 hash of the converted file
	outputMD5 := md5.New()
	if _, err = io.Copy(outputMD5, temporaryOutputFileHandle); err != nil {
		return
	}
	outputMD5Sum := outputMD5.Sum(nil)

	md5Exists, err := p.checkIfVrddtMD5Exists(ctx, outputMD5Sum, redditVideo)
	if err != nil {
		return
	} else if md5Exists {
		p.log.Debugf("Vrddt MD5 already exists in the database")
		return
	}
	p.log.Debugf("MD5 for the resulting vrddt video does not exist in the database")

	// The vrddt video is unique so setup a new one and assign the hash
	vrddtVideo := domain.NewVrddtVideo()
	vrddtVideo.MD5 = outputMD5Sum

	p.log.Debugf("Uploading media to storage for Reddit URL: %s", redditVideo.URL)

	destinationFilename := vrddtVideo.ID.Hex() + OutputFileExtension
	if err = p.storage.Upload(ctx, temporaryOutputFileHandle.Name(), destinationFilename); err != nil {
		return
	}
	vrddtVideo.URL, err = p.storage.GetLocation(ctx, destinationFilename)
	if err != nil {
		p.storage.Delete(ctx, destinationFilename)
		return
	}

	p.log.Debugf("Vrddt media uploaded to storage as URL: %s", vrddtVideo.URL)

	// Save the vrddt video information to the database
	err = p.store.CreateVrddtVideo(ctx, vrddtVideo)
	if err != nil {
		p.storage.Delete(ctx, destinationFilename)
		return
	}

	// If we got this far then the Reddit URL is unique and either:
	// A) A vrddt video was found that already has processed the Reddit
	// audio and video before so create the association
	// B) This link, video,audio, and the generated vrddt video are all
	// unique so store all of these values
	redditVideo.VrddtVideoID = vrddtVideo.ID
	if err = p.store.CreateRedditVideo(ctx, redditVideo); err != nil {
		p.store.DeleteVrddtVideo(
			ctx,
			store.Selector{
				"_id": vrddtVideo.ID,
			},
		)
		p.storage.Delete(ctx, destinationFilename)
		return
	}

	p.log.Infof("Completed storing media [VrddtVideo URL: %s] for Reddit URL: %s",
		vrddtVideo.URL,
		redditVideo.URL,
	)

	return
}
