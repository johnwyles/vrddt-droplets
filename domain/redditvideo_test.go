package domain_test

import (
	"fmt"
	"testing"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
)

func TestUser_CheckSecret(t *testing.T) {
	password := "hello@world!"

	user := domain.RedditVideo{}
	user.Secret = password
	err := user.HashSecret()
	if err != nil {
		t.Errorf("was not expecting error, got '%s'", err)
	}

	if !user.CheckSecret(password) {
		t.Errorf("CheckSecret expected to return true, but got false")
	}
}

func TestUser_Validate(suite *testing.T) {
	suite.Parallel()

	cases := []struct {
		user      domain.RedditVideo
		expectErr bool
		errType   string
	}{
		{
			user:      domain.RedditVideo{},
			expectErr: true,
			errType:   errors.TypeMissingField,
		},
		{
			user: domain.RedditVideo{
				Meta: domain.Meta{
					Name: "johnwyles",
				},
				Email: "blah.com",
			},
			expectErr: true,
			errType:   errors.TypeInvalidValue,
		},
		{
			user: domain.RedditVideo{
				Meta: domain.Meta{
					Name: "johnwyles",
				},
				Email: "johnwyles <no-mail@nomail.com>",
			},
			expectErr: false,
		},
	}

	for id, cs := range cases {
		suite.Run(fmt.Sprintf("Case#%d", id), func(t *testing.T) {
			testValidation(t, cs.user, cs.expectErr, cs.errType)
		})
	}
}
