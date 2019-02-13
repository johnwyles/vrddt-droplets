package domain

import (
	"net/mail"

	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// RedditVideo represents information about registered reddit videos.
type RedditVideo struct {
	Meta `json:",inline,omitempty" bson:",inline"`

	// Email should contain a valid email of the user.
	Email string `json:"email,omitempty" bson:"email"`

	// Secret represents the user secret.
	Secret string `json:"secret,omitempty" bson:"secret"`
}

// Validate performs basic validation of user information.
func (redditVideo RedditVideo) Validate() error {
	if err := redditVideo.Meta.Validate(); err != nil {
		return err
	}

	_, err := mail.ParseAddress(redditVideo.Email)
	if err != nil {
		return errors.InvalidValue("Email", err.Error())
	}

	return nil
}

// HashSecret creates bcrypt hash of the password.
func (redditVideo *RedditVideo) HashSecret() error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(redditVideo.Secret), 4)
	if err != nil {
		return err
	}
	redditVideo.Secret = string(bytes)
	return nil
}

// CheckSecret compares the cleartext password with the hash.
func (redditVideo RedditVideo) CheckSecret(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(redditVideo.Secret), []byte(password))
	return err == nil
}
