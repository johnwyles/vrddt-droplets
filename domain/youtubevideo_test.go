package domain_test

import (
	"fmt"
	"testing"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
)

func TestYoutubeVideo_Validate(suite *testing.T) {
	suite.Parallel()

	validMeta := domain.Meta{
		ID: bson.NewObjectId(),
	}

	invalidURL := "foo.html"
	validURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

	cases := []struct {
		youtubeVideo domain.YoutubeVideo
		expectErr    bool
	}{
		{
			youtubeVideo: domain.YoutubeVideo{},
			expectErr:    true,
		},
		{
			youtubeVideo: domain.YoutubeVideo{
				Meta: validMeta,
			},
			expectErr: true,
		},
		{
			youtubeVideo: domain.YoutubeVideo{
				Meta: validMeta,
				URL:  validURL,
			},
			expectErr: false,
		},
		{
			youtubeVideo: domain.YoutubeVideo{
				Meta: validMeta,
				URL:  invalidURL,
			},
			expectErr: true,
		},
	}

	for id, cs := range cases {
		suite.Run(fmt.Sprintf("#%d", id), func(t *testing.T) {
			err := cs.youtubeVideo.Validate()
			if err != nil {
				if !cs.expectErr {
					t.Errorf("was not expecting error, got '%s'", err)
				}
				return
			}

			if cs.expectErr {
				t.Errorf("was expecting error, got nil")
			}
		})
	}
}
