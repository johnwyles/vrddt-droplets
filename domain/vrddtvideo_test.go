package domain_test

import (
	"fmt"
	"testing"

	"github.com/johnwyles/vrddt-droplets/domain"
)

func TestPost_Validate(suite *testing.T) {
	suite.Parallel()

	validMeta := domain.Meta{
		Name: "hello",
	}

	cases := []struct {
		post      domain.VrddtVideo
		expectErr bool
	}{
		{
			post:      domain.VrddtVideo{},
			expectErr: true,
		},
		{
			post: domain.VrddtVideo{
				Meta: validMeta,
			},
			expectErr: true,
		},
		{
			post: domain.VrddtVideo{
				Meta: validMeta,
				Body: "hello world post!",
			},
			expectErr: true,
		},
		{
			post: domain.VrddtVideo{
				Meta:  validMeta,
				Type:  "blah",
				Owner: "johnwyles",
				Body:  "hello world post!",
			},
			expectErr: true,
		},
		{
			post: domain.VrddtVideo{
				Meta: validMeta,
				Type: domain.ContentLibrary,
				Body: "hello world post!",
			},
			expectErr: true,
		},
		{
			post: domain.VrddtVideo{
				Meta:  validMeta,
				Type:  domain.ContentLibrary,
				Body:  "hello world post!",
				Owner: "johnwyles",
			},
			expectErr: false,
		},
	}

	for id, cs := range cases {
		suite.Run(fmt.Sprintf("#%d", id), func(t *testing.T) {
			err := cs.post.Validate()
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
