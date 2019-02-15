package domain

import (
	"net/url"
	// "strings"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/pkg/errors"
)

// TODO: Do we add AudioURL, Title, and VideoURL, and VrddtVideo to Validate()?

// RedditVideo represents information about registered reddit videos.
type RedditVideo struct {
	// AudioURL is the URL to the audio for the reddit video.
	AudioURL string `json:"audio_url,omitempty" bson:"audio_url,omitempty"`

	// Meta holds the generic information about the vrddt video.
	Meta `json:",inline,omitempty" bson:",inline,omitempty"`

	// URL should contain a valid URL for the reddit link for the reddit video.
	URL string `json:"url,omitempty" bson:"url,omitempty"`

	// Title is the title of the reddit link for the reddit video.
	Title string `json:"title,omitempty" bson:"title,omitempty"`

	// VideoURL is the URL to the video for the reddit video.
	VideoURL string `json:"video_url,omitempty" bson:"video_url,omitempty"`

	// VrddtVideo represents the ID of the vrddt video that we created with this video.
	VrddtVideoID bson.ObjectId `json:"vrddt_video_id,omitempty" bson:"vrddt_video_id,omitempty"`
}

// Validate performs basic validation of user information.
func (redditVideo RedditVideo) Validate() error {
	// _, err := url.ParseRequestURI(redditVideo.AudioURL)
	// if err != nil {
	// 	return errors.InvalidValue("AudioURL", err.Error())
	// }

	if err := redditVideo.Meta.Validate(); err != nil {
		return err
	}

	// if len(strings.TrimSpace(string(redditVideo.Title))) == 0 {
	// 	return errors.MissingField("Title")
	// }

	_, err := url.ParseRequestURI(redditVideo.URL)
	if err != nil {
		return errors.InvalidValue("URL", err.Error())
	}

	// _, err = url.ParseRequestURI(redditVideo.VideoURL)
	// if err != nil {
	// 	return errors.InvalidValue("VideoURL", err.Error())
	// }

	// if !redditVideo.VrddtVideo.Valid() {
	// 	return errors.InvalidValue("VrddtVideo", "Not a valid bson.ObjectId")
	// }

	return nil
}
