package domain

import (
	"fmt"
	"strings"

	"github.com/johnwyles/vrddt-droplets/pkg/errors"
)

// Common content types.
const (
	ContentLibrary = "library"
	ContentLink    = "link"
	ContentVideo   = "video"
)

var validTypes = []string{ContentLibrary, ContentLink, ContentVideo}

// VrddtVideo represents an article, link, video etc.
type VrddtVideo struct {
	Meta `json:",inline" bson:",inline"`

	// Type should state the type of the content. (e.g., library,
	// video, link etc.)
	Type string `json:"type" bson:"type"`

	// Body should contain the actual content according to the Type
	// specified. (e.g. github.com/johnwyles/parens when Type=link)
	Body string `json:"body" bson:"body"`

	// Owner represents the name of the user who created the vrddtVideo.
	Owner string `json:"owner" bson:"owner"`
}

// Validate performs validation of the vrddtVideo.
func (vrddtVideo VrddtVideo) Validate() error {
	if err := vrddtVideo.Meta.Validate(); err != nil {
		return err
	}

	if len(strings.TrimSpace(vrddtVideo.Body)) == 0 {
		return errors.MissingField("Body")
	}

	if len(strings.TrimSpace(vrddtVideo.Owner)) == 0 {
		return errors.MissingField("Owner")
	}

	if !contains(vrddtVideo.Type, validTypes) {
		return errors.InvalidValue("Type", fmt.Sprintf("type must be one of: %s", strings.Join(validTypes, ",")))
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
