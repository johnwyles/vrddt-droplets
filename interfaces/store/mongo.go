package store

import (
	"context"
	"fmt"
	"time"

	mgo "gopkg.in/mgo.v2"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// TODO: Add logging

// mongoSession contains all the information about a Mongo session
type mongoSession struct {
	database                   string
	log                        logger.Logger
	redditVideosCollectionName string
	session                    *mgo.Session
	timeout                    int
	vrddtVideosCollectionName  string
	uri                        string
}

// Mongo initiates a new MongoDB connection
func Mongo(cfg *config.StoreMongoConfig, loggerHandle logger.Logger) (store Store, err error) {
	loggerHandle.Debugf("Mongo(cfg): %#v", cfg)

	store = &mongoSession{
		log:                        loggerHandle,
		redditVideosCollectionName: cfg.RedditVideosCollectionName,
		timeout:                    cfg.Timeout,
		uri:                        cfg.URI,
		vrddtVideosCollectionName:  cfg.VrddtVideosCollectionName,
	}

	return
}

// Cleanup will end the session
func (m *mongoSession) Cleanup(ctx context.Context) (err error) {
	m.log.Debugf("Cleanup()")

	if m.session == nil {
		return errors.ConnectionFailure("mongo", "A connection is not set in order to be cleaned up")
	}

	m.session.Close()

	return
}

// CreateRedditVideo will add a RedditVideo to the Reddit videos collection
func (m *mongoSession) CreateRedditVideo(ctx context.Context, redditVideo *domain.RedditVideo) (err error) {
	redditVideosCollection, err := m.redditVideosCollection()
	if err != nil {
		return
	}

	return redditVideosCollection.Insert(redditVideo)
}

// CreateVrddtVideo will add a vrddt video to the vrddt videos collection
func (m *mongoSession) CreateVrddtVideo(ctx context.Context, vrddtVideo *domain.VrddtVideo) (err error) {
	vrddtVideosCollection, err := m.vrddtVideosCollection()
	if err != nil {
		return
	}

	return vrddtVideosCollection.Insert(vrddtVideo)
}

// DeleteRedditVideos is an alias to the same function but plural becaause the
// number of videos deleted is determined by the selector
func (m *mongoSession) DeleteRedditVideo(ctx context.Context, selector Selector) (err error) {
	return m.DeleteRedditVideos(ctx, selector)
}

// DeleteRedditVideos deletes Reddit video from the collection
func (m *mongoSession) DeleteRedditVideos(ctx context.Context, selector Selector) (err error) {
	redditVideos, err := m.redditVideosCollection()
	if err != nil {
		return err
	}

	return redditVideos.Remove(selector)
}

// DeleteVrddtVideo is an alias to the same function but plural becaause the
// number of videos deleted is determined by the selector
func (m *mongoSession) DeleteVrddtVideo(ctx context.Context, selector Selector) (err error) {
	return m.DeleteVrddtVideos(ctx, selector)
}

// DeleteVrddtVideo deletes vrddtVideos from the collection
func (m *mongoSession) DeleteVrddtVideos(ctx context.Context, selector Selector) (err error) {
	vrddtVideosCollection, err := m.vrddtVideosCollection()
	if err != nil {
		return
	}

	return vrddtVideosCollection.Remove(selector)
}

// GetRedditVideo will return a Reddit video from the database if the passed in
// key / value pair are found
func (m *mongoSession) GetRedditVideo(ctx context.Context, selector Selector) (redditVideo *domain.RedditVideo, err error) {
	redditVideosCollection, err := m.redditVideosCollection()
	if err != nil {
		return
	}

	redditVideo = &domain.RedditVideo{}

	err = redditVideosCollection.Find(selector).One(redditVideo)
	// Turn error into a ResourceNotFound error type
	if err == mgo.ErrNotFound {
		return nil, errors.ResourceNotFound("RedditVideo", fmt.Sprintf("%#v", selector))
	}

	return
}

// GetRedditVideos will return a collection of Reddit videos from the database
// with a LIMIT of 100 items that match the selector
func (m *mongoSession) GetRedditVideos(ctx context.Context, selector Selector, limit int) (redditVideos []*domain.RedditVideo, err error) {
	redditVideosCollection, err := m.redditVideosCollection()
	if err != nil {
		return
	}

	redditVideos = []*domain.RedditVideo{}
	iter := redditVideosCollection.Find(selector).Limit(limit).Iter()
	err = iter.All(&redditVideos)

	return
}

// GetVrddtVideo will return a VrddtVideo from the database if the passed
// in selector is found
func (m *mongoSession) GetVrddtVideo(ctx context.Context, selector Selector) (vrddtVideo *domain.VrddtVideo, err error) {
	vrddtVideosCollection, err := m.vrddtVideosCollection()
	if err != nil {
		return
	}

	vrddtVideo = &domain.VrddtVideo{}

	err = vrddtVideosCollection.Find(selector).One(vrddtVideo)
	// Turn error into a ResourceNotFound error type
	if err == mgo.ErrNotFound {
		return nil, errors.ResourceNotFound("VrddtVideo", fmt.Sprintf("%#v", selector))
	}

	return
}

// GetVrddtVideos will return a VrddtVideo from the database if the passed
// in key / value pair are found
func (m *mongoSession) GetVrddtVideos(ctx context.Context, selector Selector, limit int) (vrddtVideos []*domain.VrddtVideo, err error) {
	vrddtVideosCollection, err := m.vrddtVideosCollection()
	if err != nil {
		return
	}

	vrddtVideos = []*domain.VrddtVideo{}
	iter := vrddtVideosCollection.Find(selector).Limit(limit).Iter()
	err = iter.All(&vrddtVideos)

	return
}

// Init starts the session to the database
func (m *mongoSession) Init(ctx context.Context) (err error) {
	dialInfo, err := mgo.ParseURL(m.uri)
	if err != nil {
		return
	}

	dialInfo.Timeout = time.Duration(m.timeout) * time.Second

	m.database = dialInfo.Database

	m.session, err = mgo.DialWithInfo(dialInfo)
	// m.session, err = mgo.DialWithTimeout(m.uri, 5*time.Second)
	if err != nil {
		return
	}

	m.session.SetMode(mgo.Monotonic, true)
	m.session.SetSafe(&mgo.Safe{})

	return
}

// redditVideosCollection returns the collection of Reddit videos previously processed
func (m *mongoSession) redditVideosCollection() (redditVideosCollection *mgo.Collection, err error) {
	redditVideosCollection = m.session.DB(m.database).C(m.redditVideosCollectionName)
	err = redditVideosCollection.EnsureIndex(
		mgo.Index{
			Key:        []string{"reddit_url"},
			Unique:     true,
			DropDups:   true,
			Background: true,
			Sparse:     true,
		},
	)

	return
}

// vrddtVideosCollection returns the collection of vrddt videos previously processed
func (m *mongoSession) vrddtVideosCollection() (vrddtVideosCollection *mgo.Collection, err error) {
	vrddtVideosCollection = m.session.DB(m.database).C(m.vrddtVideosCollectionName)
	err = vrddtVideosCollection.EnsureIndex(
		mgo.Index{
			Key:        []string{"md5"},
			Unique:     true,
			DropDups:   true,
			Background: true,
			Sparse:     true,
		},
	)

	return
}
