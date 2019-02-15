package mongo

import (
	"context"
	"time"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const colVrddtVideos = "vrddt_videos"

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
func (vrddtVideos *VrddtVideoStore) Exists(ctx context.Context, id bson.ObjectId) bool {
	col := vrddtVideos.db.C(colVrddtVideos)

	count, err := col.FindId(id).Count()
	if err != nil {
		return false
	}
	return count > 0
}

// TODO
// FindAll finds all users matching the tags.
func (vvs *VrddtVideoStore) FindAll(ctx context.Context, limit int) ([]domain.VrddtVideo, error) {
	col := vvs.db.C(colVrddtVideos)

	matches := []domain.VrddtVideo{}
	if err := col.Find(nil).Limit(limit).All(&matches); err != nil {
		return nil, errors.Wrapf(err, "failed to query for vrddt videos")
	}
	return matches, nil
}

// FindByID finds a vrddt video by id. If not found, returns ResourceNotFound error.
func (rvs *VrddtVideoStore) FindByID(ctx context.Context, id bson.ObjectId) (*domain.VrddtVideo, error) {
	col := rvs.db.C(colVrddtVideos)

	vrddtVideo := domain.VrddtVideo{}
	if err := col.FindId(id).One(&vrddtVideo); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.ResourceNotFound("VrddtVideo", id.Hex())
		}
		return nil, errors.Wrapf(err, "failed to fetch vrddt video")
	}

	vrddtVideo.SetDefaults()
	return &vrddtVideo, nil
}

// FindByMD5 finds a vrddt video by url. If not found, returns ResourceNotFound error.
func (vvs *VrddtVideoStore) FindByMD5(ctx context.Context, md5 string) (*domain.VrddtVideo, error) {
	col := vvs.db.C(colVrddtVideos)

	vrddtVideo := domain.VrddtVideo{}
	if err := col.FindId(bson.M{"md5": md5}).One(&vrddtVideo); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.ResourceNotFound("VrddtVideo MD5", md5)
		}
		return nil, errors.Wrapf(err, "failed to fetch vrddt video")
	}

	vrddtVideo.SetDefaults()
	return &vrddtVideo, nil
}

// Save validates and persists the vrddtVideo.
func (vrddtVideos *VrddtVideoStore) Save(ctx context.Context, vrddtVideo *domain.VrddtVideo) (*domain.VrddtVideo, error) {
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
	return vrddtVideo, nil
}

// Delete removes one vrddtVideo identified by the name.
func (vrddtVideos *VrddtVideoStore) Delete(ctx context.Context, id bson.ObjectId) (*domain.VrddtVideo, error) {
	col := vrddtVideos.db.C(colVrddtVideos)

	ch := mgo.Change{
		Remove:    true,
		ReturnNew: true,
		Upsert:    false,
	}
	vrddtVideo := domain.VrddtVideo{}
	_, err := col.FindId(id).Apply(ch, &vrddtVideo)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.ResourceNotFound("VrddtVideo ID", id.Hex())
		}
		return nil, err
	}

	return &vrddtVideo, nil
}
