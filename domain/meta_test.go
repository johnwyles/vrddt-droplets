package domain_test

import (
	"fmt"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
)

func TestMeta_Validate(suite *testing.T) {
	suite.Parallel()

	var invalidID bson.ObjectId

	invalidMeta := domain.Meta{
		ID: invalidID,
	}

	validMeta := domain.Meta{
		ID:        bson.NewObjectId(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newValidMeta := domain.NewMeta()

	cases := []struct {
		meta      domain.Meta
		expectErr bool
		errType   string
	}{
		{
			meta:      invalidMeta,
			expectErr: true,
			errType:   "InvalidValue",
		},
		{
			meta:      validMeta,
			expectErr: false,
		},
		{
			meta:      newValidMeta,
			expectErr: false,
		},
	}

	for id, cs := range cases {
		suite.Run(fmt.Sprintf("Case#%d", id), func(t *testing.T) {
			testValidation(t, cs.meta, cs.expectErr, cs.errType)
		})
	}

}

func testValidation(t *testing.T, validator validatable, expectErr bool, errType string) {
	err := validator.Validate()
	if err != nil {
		if !expectErr {
			t.Errorf("unexpected error: %s", err)
			return
		}

		if actualType := errors.Type(err); actualType != errType {
			t.Errorf("expecting error type '%s', got '%s'", errType, actualType)
		}
		return
	}

	if expectErr {
		t.Errorf("was expecting an error of type '%s', got nil", errType)
		return
	}
}

type validatable interface {
	Validate() error
}
