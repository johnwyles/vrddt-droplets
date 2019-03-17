package store

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// TODO: Add logging
// TODO: Turn QueryLimit into a configuration variable

var (
	// QueryLimit will limit the number of results returned in requests to this
	QueryLimit = 100
)

// mongoSession contains all the information about a Mongo session
type mongoSession struct {
	database                   string
	log                        logger.Logger
	redditVideosCollectionName string
	session                    *mgo.Session
	vrddtVideosCollectionName  string
	uri                        string
}

// Mongo initiates a new MongoDB connection
func Mongo(cfg *config.StoreMongoConfig, loggerHandle logger.Logger) (store Store, err error) {
	loggerHandle.Debugf("Mongo(cfg): %#v", cfg)

	store = &mongoSession{
		log:                        loggerHandle,
		redditVideosCollectionName: cfg.RedditVideosCollectionName,
		uri:                        cfg.URI,
		vrddtVideosCollectionName:  cfg.VrddtVideosCollectionName,
	}

	return
}

// Cleanup will end the session
func (m *mongoSession) Cleanup() (err error) {
	m.log.Debugf("Cleanup()")

	if m.session == nil {
		return fmt.Errorf("A connection is set in order to be cleaned up")
	}

	m.session.Close()

	return
}

// CreateRedditVideo will add a RedditVideo to the Reddit videos collection
func (m *mongoSession) CreateRedditVideo(redditVideo *domain.RedditVideo) (err error) {
	redditVideosCollection, err := m.redditVideosCollection()
	if err != nil {
		return
	}

	return redditVideosCollection.Insert(redditVideo)
}

// CreateVrddtVideo will add a vrddt video to the vrddt videos collection
func (m *mongoSession) CreateVrddtVideo(vrddtVideo *domain.VrddtVideo) (err error) {
	vrddtVideosCollection, err := m.vrddtVideosCollection()
	if err != nil {
		return
	}

	return vrddtVideosCollection.Insert(vrddtVideo)
}

// DeleteRedditVideos is an alias to the same function but plural becaause the
// number of videos deleted is determined by the selector
func (m *mongoSession) DeleteRedditVideo(selector Selector) (err error) {
	return m.DeleteRedditVideos(selector)
}

// DeleteRedditVideos deletes Reddit video from the collection
func (m *mongoSession) DeleteRedditVideos(selector Selector) (err error) {
	redditVideos, err := m.redditVideosCollection()
	if err != nil {
		return err
	}

	if err = redditVideos.Remove(selector); err != nil {
		return err
	}

	return
}

// DeleteVrddtVideo is an alias to the same function but plural becaause the
// number of videos deleted is determined by the selector
func (m *mongoSession) DeleteVrddtVideo(selector Selector) (err error) {
	return m.DeleteVrddtVideos(selector)
}

// DeleteVrddtVideo deletes vrddtVideos from the collection
func (m *mongoSession) DeleteVrddtVideos(selector Selector) (err error) {
	vrddtVideosCollection, err := m.vrddtVideosCollection()
	if err != nil {
		return err
	}

	if err = vrddtVideosCollection.Remove(selector); err != nil {
		return err
	}

	return
}

// GetRedditVideo will return a Reddit video from the database if the passed in
// key / value pair are found
func (m *mongoSession) GetRedditVideo(selector Selector) (redditVideo *domain.RedditVideo, err error) {
	redditVideosCollection, err := m.redditVideosCollection()
	if err != nil {
		return
	}

	redditVideo = &domain.RedditVideo{}
	err = redditVideosCollection.Find(selector).One(redditVideo)

	return
}

// GetRedditVideos will return a collection of Reddit videos from the database
// with a LIMIT of 100 items that match the selector
func (m *mongoSession) GetRedditVideos(selector Selector) (redditVideos []*domain.RedditVideo, err error) {
	redditVideosCollection, err := m.redditVideosCollection()
	if err != nil {
		return
	}

	redditVideos = []*domain.RedditVideo{}
	iter := redditVideosCollection.Find(selector).Limit(100).Iter()
	err = iter.All(&redditVideos)

	return
}

// GetVrddtVideo will return a VrddtVideo from the database if the passed
// in selector is found
func (m *mongoSession) GetVrddtVideo(selector Selector) (vrddtVideo *domain.VrddtVideo, err error) {
	vrddtVideosCollection, err := m.vrddtVideosCollection()
	if err != nil {
		return
	}

	vrddtVideo = &domain.VrddtVideo{}
	err = vrddtVideosCollection.Find(selector).One(vrddtVideo)

	return
}

// GetVrddtVideos will return a VrddtVideo from the database if the passed
// in key / value pair are found
func (m *mongoSession) GetVrddtVideos(selector Selector) (vrddtVideos []*domain.VrddtVideo, err error) {
	vrddtVideosCollection, err := m.vrddtVideosCollection()
	if err != nil {
		return
	}

	vrddtVideos = []*domain.VrddtVideo{}
	iter := vrddtVideosCollection.Find(selector).Limit(100).Iter()
	err = iter.All(&vrddtVideos)

	return
}

// Init starts the session to the database
func (m *mongoSession) Init() (err error) {
	dialInfo, err := mgo.ParseURL(m.uri)
	if err != nil {
		return
	}

	m.database = dialInfo.Database

	m.session, err = mgo.DialWithInfo(dialInfo)
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
	if err = redditVideosCollection.EnsureIndex(
		mgo.Index{
			Key:        []string{"reddit_url"},
			Unique:     true,
			DropDups:   true,
			Background: true,
			Sparse:     true,
		},
	); err != nil {
		return nil, err
	}

	return redditVideosCollection, nil
}

// vrddtVideosCollection returns the collection of vrddt videos previously processed
func (m *mongoSession) vrddtVideosCollection() (vrddtVideosCollection *mgo.Collection, err error) {
	vrddtVideosCollection = m.session.DB(m.database).C(m.vrddtVideosCollectionName)
	vrddtVideosCollection.EnsureIndex(
		mgo.Index{
			Key:        []string{"md5"},
			Unique:     true,
			DropDups:   true,
			Background: true,
			Sparse:     true,
		},
	)

	return vrddtVideosCollection, nil
}
