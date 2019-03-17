package store

import (
	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/interfaces/config"
	"github.com/johnwyles/vrddt-droplets/pkg/logger"
)

// TODO: Finish this when you're are bored

// mongoSession contains all the information about a Mongo session
type memoryStore struct {
	log          logger.Logger
	redditVideos []*domain.RedditVideo
	vrddtVideos  []*domain.VrddtVideo
}

// Memory initiates a new Memory struct
func Memory(cfg *config.StoreMemoryConfig, loggerHandle logger.Logger) (store Store, err error) {
	loggerHandle.Debugf("Memory(cfg): %#v", cfg)

	store = &memoryStore{
		log:          loggerHandle,
		redditVideos: []*domain.RedditVideo{},
		vrddtVideos:  []*domain.VrddtVideo{},
	}

	return
}

// Cleanup will end the session
func (m *memoryStore) Cleanup() (err error) {
	m = &memoryStore{
		log:          m.log,
		redditVideos: []*domain.RedditVideo{},
		vrddtVideos:  []*domain.VrddtVideo{},
	}

	return
}

// CreateRedditVideo will add a RedditVideo to the Reddit videos
func (m *memoryStore) CreateRedditVideo(redditVideo *domain.RedditVideo) (err error) {
	m.redditVideos = append(m.redditVideos, redditVideo)
	return
}

// CreateVrddtVideo will add a vrddt video to the vrddt videos
func (m *memoryStore) CreateVrddtVideo(vrddtVideo *domain.VrddtVideo) (err error) {
	m.vrddtVideos = append(m.vrddtVideos, vrddtVideo)
	return
}

// DeleteRedditVideo is an alias to the same function but plural becaause the
// number of videos deleted is determined by the selector
func (m *memoryStore) DeleteRedditVideo(selector Selector) (err error) {
	return m.DeleteRedditVideos(selector)
}

// DeleteRedditVideos deletes Reddit video from the collection
func (m *memoryStore) DeleteRedditVideos(selector Selector) (err error) {
	return
}

// DeleteVrddtVideo is an alias to the same function but plural becaause the
// number of videos deleted is determined by the selector
func (m *memoryStore) DeleteVrddtVideo(selector Selector) (err error) {
	return m.DeleteVrddtVideos(selector)
}

// DeleteVrddtVideo deletes vrddtVideos from the collection
func (m *memoryStore) DeleteVrddtVideos(selector Selector) (err error) {
	return
}

// GetRedditVideo will return a RedditVideo from the database if the passed in
// key / value pair are found
func (m *memoryStore) GetRedditVideo(selector Selector) (redditVideo *domain.RedditVideo, err error) {
	return
}

func (m *memoryStore) GetRedditVideos(selector Selector) (redditVideos []*domain.RedditVideo, err error) {
	// selector := map[string]interface{}{
	// 	"audio_url": redditVideo.AudioURL,
	// 	"video_url": redditVideo.VideoURL,
	// }

	// type Video struct {
	// 	AudioURL     string        `json:"audio_url" bson:"audio_url"`
	// 	FilePath     string        `json:"-" bson:"-"`
	// 	FileHandle   *os.File      `json:"-" bson:"-"`
	// 	ID           bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	// 	URL          string        `json:"url" bson:"url"`
	// 	Timestamp    time.Time     `json:"timestamp" bson:"timestamp"`
	// 	Title        string        `json:"title" bson:"title"`
	// 	VideoURL     string        `json:"video_url" bson:"video_url"`
	// 	VrddtVideoID bson.ObjectId `json:"vrddt_video_id" bson:"vrddt_video_id"`
	// }

	// for i, rv := range m.redditVideos {
	// 	var rvMatch *reddit.Video
	// 	for k, v := range selector {
	//
	// 	}
	// }

	return
}

// GetVrddtVideo will return a VrddtVideo from the database if the passed
// in selector is found
func (m *memoryStore) GetVrddtVideo(selector Selector) (vrddtVideo *domain.VrddtVideo, err error) {
	return
}

// GetVrddtVideos will return a VrddtVideo from the database if the passed
// in key / value pair are found
func (m *memoryStore) GetVrddtVideos(selector Selector) (vrddtVideos []*domain.VrddtVideo, err error) {
	return
}

// Init provides some initialization for the store
func (m *memoryStore) Init() (err error) {
	return
}
