package domain_test

import (
	"fmt"
	"testing"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
)

func TestRedditVideo_Validate(suite *testing.T) {
	suite.Parallel()

	validMeta := domain.Meta{
		ID: bson.NewObjectId(),
	}

	invalidURL := "foo.html"
	validURL := "https://www.reddit.com/r/MadeMeSmile/comments/apt8tb/need_more_people_like_him/"

	cases := []struct {
		redditVideo domain.RedditVideo
		expectErr   bool
		errType     string
	}{
		{
			redditVideo: domain.RedditVideo{},
			expectErr:   true,
			errType:     errors.TypeInvalidValue,
		},
		{
			redditVideo: domain.RedditVideo{
				Meta: validMeta,
				URL:  invalidURL,
			},
			expectErr: true,
			errType:   errors.TypeInvalidValue,
		},
		{
			redditVideo: domain.RedditVideo{
				Meta: validMeta,
				URL:  validURL,
			},
			expectErr: false,
		},
	}

	for id, cs := range cases {
		suite.Run(fmt.Sprintf("Case#%d", id), func(t *testing.T) {
			testValidation(t, cs.redditVideo, cs.expectErr, cs.errType)
		})
	}
}
