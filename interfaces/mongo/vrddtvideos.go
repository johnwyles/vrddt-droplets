package mongo

import (
	"context"
	"time"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const colVrddtVideos = "vrddtVideos"

// NewVrddtVideo initializes the VrddtVideos store with given mongo db handle.
func NewVrddtVideoStore(db *mgo.Database) *VrddtVideoStore {
	return &VrddtVideoStore{
		db: db,
	}
}

// VrddtVideoStore manages persistence and retrieval of vrddtVideos.
type VrddtVideoStore struct {
	db *mgo.Database
}

// Exists checks if a vrddtVideo exists by name.
func (vrddtVideos *VrddtVideoStore) Exists(ctx context.Context, name string) bool {
	col := vrddtVideos.db.C(colVrddtVideos)

	count, err := col.Find(bson.M{"name": name}).Count()
	if err != nil {
		return false
	}
	return count > 0
}

// Get finds a vrddtVideo by name.
func (vrddtVideos *VrddtVideoStore) Get(ctx context.Context, name string) (*domain.VrddtVideo, error) {
	col := vrddtVideos.db.C(colVrddtVideos)

	vrddtVideo := domain.VrddtVideo{}
	if err := col.Find(bson.M{"name": name}).One(&vrddtVideo); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.ResourceNotFound("VrddtVideo", name)
		}
		return nil, errors.Wrapf(err, "failed to fetch vrddtVideo")
	}

	vrddtVideo.SetDefaults()
	return &vrddtVideo, nil
}

// Save validates and persists the vrddtVideo.
func (vrddtVideos *VrddtVideoStore) Save(ctx context.Context, vrddtVideo domain.VrddtVideo) (*domain.VrddtVideo, error) {
	vrddtVideo.SetDefaults()
	if err := vrddtVideo.Validate(); err != nil {
		return nil, err
	}
	vrddtVideo.CreatedAt = time.Now()
	vrddtVideo.UpdatedAt = time.Now()

	col := vrddtVideos.db.C(colVrddtVideos)
	if err := col.Insert(vrddtVideo); err != nil {
		return nil, err
	}
	return &vrddtVideo, nil
}

// Delete removes one vrddtVideo identified by the name.
func (vrddtVideos *VrddtVideoStore) Delete(ctx context.Context, name string) (*domain.VrddtVideo, error) {
	col := vrddtVideos.db.C(colVrddtVideos)

	ch := mgo.Change{
		Remove:    true,
		ReturnNew: true,
		Upsert:    false,
	}
	vrddtVideo := domain.VrddtVideo{}
	_, err := col.Find(bson.M{"name": name}).Apply(ch, &vrddtVideo)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.ResourceNotFound("VrddtVideo", name)
		}
		return nil, err
	}

	return &vrddtVideo, nil
}
