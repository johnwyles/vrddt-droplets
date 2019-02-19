package domain

import (
	"net/url"
	"strings"

	"github.com/johnwyles/vrddt-droplets/pkg/errors"
)

// VrddtVideo represents a vrddt video.
type VrddtVideo struct {
	// Meta holds the generic information about the vrddt video.
	Meta `json:",inline" bson:",inline"`

	// MD5 is the md5 hash of the contents of the vrddt video.
	MD5 []byte `json:"md5,omitempty" bson:"md5,omitempty"`

	// URL represents a publicly accessibly path to the asset.
	URL string `json:"url,omitempty" bson:"url,omitempty"`
}

// NewVrddtVideo will return a new vrddt video.
func NewVrddtVideo() *VrddtVideo {
	return &VrddtVideo{
		Meta: NewMeta(),
	}
}

// Validate performs validation of the vrddt video.
func (vrddtVideo VrddtVideo) Validate() error {
	if err := vrddtVideo.Meta.Validate(); err != nil {
		return err
	}

	if len(strings.TrimSpace(string(vrddtVideo.MD5))) == 0 {
		return errors.MissingField("MD5")
	}

	_, err := url.ParseRequestURI(vrddtVideo.URL)
	if err != nil {
		return errors.MissingField("URL")
	}

	return nil
}

func contains(val string, vals []string) bool {
	for _, item := range vals {
		if val == item {
			return true
		}
	}
	return false
}
