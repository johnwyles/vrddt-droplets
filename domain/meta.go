package domain

import (
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/johnwyles/vrddt-droplets/pkg/errors"
)

// Meta represents metadata about different entities.
type Meta struct {
	// CreatedAt represents the time at which this object was created.
	CreatedAt time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`

	// ID represents a unique identifier for the object.
	ID bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`

	// UpdatedAt represents the time at which this object was last
	// modified.
	UpdatedAt time.Time `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

// SetDefaults sets sensible defaults on meta.
func (meta *Meta) SetDefaults() {
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = time.Now()
		meta.UpdatedAt = time.Now()
	}
}

// Validate performs basic validation of the metadata.
func (meta Meta) Validate() error {
	switch {
	case !meta.ID.Valid():
		return errors.InvalidValue("ID", "Not a valid bson.ObjectId")
	}
	return nil
}

func empty(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}
