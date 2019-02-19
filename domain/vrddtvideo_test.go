package domain_test

import (
	"fmt"
	"testing"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
)

func TestVrddtVideo_Validate(suite *testing.T) {
	suite.Parallel()

	invalidMD5 := []byte("")
	validMD5 := []byte("h5K3pevUsf64fkbEr1CPVQ==")

	validMeta := domain.Meta{
		ID: bson.NewObjectId(),
	}

	invalidURL := "foo.html"
	validURL := "https://www.googleapis.com/download/storage/v1/b/vrddt-media/o/5c630e846161b663394dd342.mp4?generation=1549995660850823&alt=media"

	cases := []struct {
		vrddtVideo domain.VrddtVideo
		expectErr  bool
	}{
		{
			vrddtVideo: domain.VrddtVideo{},
			expectErr:  true,
		},
		{
			vrddtVideo: domain.VrddtVideo{
				Meta: validMeta,
			},
			expectErr: true,
		},
		{
			vrddtVideo: domain.VrddtVideo{
				MD5:  validMD5,
				Meta: validMeta,
			},
			expectErr: true,
		},
		{
			vrddtVideo: domain.VrddtVideo{
				MD5:  validMD5,
				Meta: validMeta,
				URL:  validURL,
			},
			expectErr: false,
		},
		{
			vrddtVideo: domain.VrddtVideo{
				MD5:  invalidMD5,
				Meta: validMeta,
				URL:  validURL,
			},
			expectErr: true,
		},
		{
			vrddtVideo: domain.VrddtVideo{
				MD5:  invalidMD5,
				Meta: validMeta,
				URL:  invalidURL,
			},
			expectErr: true,
		},
	}

	for id, cs := range cases {
		suite.Run(fmt.Sprintf("#%d", id), func(t *testing.T) {
			err := cs.vrddtVideo.Validate()
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
