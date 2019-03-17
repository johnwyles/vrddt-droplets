package domain

import (
	"net/url"

	"github.com/johnwyles/vrddt-droplets/pkg/errors"
)

// YoutubeVideo represents a vrddt video.
type YoutubeVideo struct {
	// Meta holds the generic information about the vrddt video.
	Meta `json:",inline" bson:",inline"`

	// MD5 is the md5 hash of the contents of the vrddt video.
	MD5 []byte `json:"md5,omitempty" bson:"md5,omitempty"`

	// URL represents a publicly accessibly path to the asset.
	URL string `json:"url,omitempty" bson:"url,omitempty"`
}

// NewYoutubeVideo will return a new YouTube video.
func NewYoutubeVideo() *YoutubeVideo {
	return &YoutubeVideo{
		Meta: NewMeta(),
	}
}

// Validate performs validation of the YouTube video.
func (youtubeVideo YoutubeVideo) Validate() error {
	if err := youtubeVideo.Meta.Validate(); err != nil {
		return err
	}
	_, err := url.ParseRequestURI(youtubeVideo.URL)
	if err != nil {
		return errors.MissingField("URL")
	}

	return nil
}
