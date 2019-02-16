package mongo

import (
	"context"
	"time"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const colRedditVideos = "reddit_videos"

// NewRedditVideoStore initializes a users store with the given db handle.
func NewRedditVideoStore(db *mgo.Database) *RedditVideoStore {
	return &RedditVideoStore{
		db: db,
	}
}

// RedditVideoStore provides functions for persisting RedditVideo entities in MongoDB.
type RedditVideoStore struct {
	db *mgo.Database
}

// Delete removes one redditVideo identified by the id.
func (rvs *RedditVideoStore) Delete(ctx context.Context, id bson.ObjectId) (*domain.RedditVideo, error) {
	col := rvs.db.C(colRedditVideos)

	ch := mgo.Change{
		Remove:    true,
		ReturnNew: true,
		Upsert:    false,
	}
	redditVideo := domain.RedditVideo{}
	_, err := col.FindId(id).Apply(ch, &redditVideo)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.ResourceNotFound("RedditVideo ID", id.Hex())
		}
		return nil, err
	}

	return &redditVideo, nil
}

// Exists checks if the redditVideo identified by the given username already
// exists. Will return false in case of any error.
func (rvs *RedditVideoStore) Exists(ctx context.Context, id bson.ObjectId) bool {
	col := rvs.db.C(colRedditVideos)

	count, err := col.Find(bson.M{"_id": id}).Count()
	if err != nil {
		return false
	}
	return count > 0
}

// TODO
// FindAll finds all users matching the tags.
func (rvs *RedditVideoStore) FindAll(ctx context.Context, limit int) ([]domain.RedditVideo, error) {
	col := rvs.db.C(colRedditVideos)

	matches := []domain.RedditVideo{}
	if err := col.Find(nil).Limit(limit).All(&matches); err != nil {
		return nil, errors.Wrapf(err, "failed to query for all reddit videos")
	}
	return matches, nil
}

// FindByID finds a reddit video by id. If not found, returns ResourceNotFound error.
func (rvs *RedditVideoStore) FindByID(ctx context.Context, id bson.ObjectId) (*domain.RedditVideo, error) {
	col := rvs.db.C(colRedditVideos)

	redditVideo := domain.RedditVideo{}
	if err := col.FindId(id).One(&redditVideo); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.ResourceNotFound("RedditVideo ID", id.Hex())
		}
		return nil, errors.Wrapf(err, "failed to find reddit video by id")
	}

	redditVideo.SetDefaults()
	return &redditVideo, nil
}

// FindByURL finds a reddit video by url. If not found, returns ResourceNotFound error.
func (rvs *RedditVideoStore) FindByURL(ctx context.Context, url string) (*domain.RedditVideo, error) {
	col := rvs.db.C(colRedditVideos)

	redditVideo := domain.RedditVideo{}
	if err := col.Find(bson.M{"url": url}).One(&redditVideo); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.ResourceNotFound("RedditVideo URL", url)
		}
		return nil, errors.Wrapf(err, "failed to find reddit video by url")
	}

	redditVideo.SetDefaults()
	return &redditVideo, nil
}

// Save validates and persists the redditVideo.
func (rvs *RedditVideoStore) Save(ctx context.Context, redditVideo *domain.RedditVideo) (*domain.RedditVideo, error) {
	redditVideo.SetDefaults()
	if err := redditVideo.Validate(); err != nil {
		return nil, err
	}
	redditVideo.CreatedAt = time.Now()
	redditVideo.UpdatedAt = time.Now()

	col := rvs.db.C(colRedditVideos)
	if err := col.Insert(redditVideo); err != nil {
		return nil, err
	}
	return redditVideo, nil
}
