package mongo

import (
	"context"
	"time"

	"github.com/johnwyles/vrddt-droplets/domain"
	"github.com/johnwyles/vrddt-droplets/pkg/errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const colRedditVideo = "users"

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

// Exists checks if the redditVideo identified by the given username already
// exists. Will return false in case of any error.
func (users *RedditVideoStore) Exists(ctx context.Context, name string) bool {
	col := users.db.C(colRedditVideo)

	count, err := col.Find(bson.M{"name": name}).Count()
	if err != nil {
		return false
	}
	return count > 0
}

// Save validates and persists the redditVideo.
func (users *RedditVideoStore) Save(ctx context.Context, redditVideo domain.RedditVideo) (*domain.RedditVideo, error) {
	redditVideo.SetDefaults()
	if err := redditVideo.Validate(); err != nil {
		return nil, err
	}
	redditVideo.CreatedAt = time.Now()
	redditVideo.UpdatedAt = time.Now()

	col := users.db.C(colRedditVideo)
	if err := col.Insert(redditVideo); err != nil {
		return nil, err
	}
	return &redditVideo, nil
}

// FindByName finds a redditVideo by name. If not found, returns ResourceNotFound error.
func (users *RedditVideoStore) FindByName(ctx context.Context, name string) (*domain.RedditVideo, error) {
	col := users.db.C(colRedditVideo)

	redditVideo := domain.RedditVideo{}
	if err := col.Find(bson.M{"name": name}).One(&redditVideo); err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.ResourceNotFound("RedditVideo", name)
		}
		return nil, errors.Wrapf(err, "failed to fetch redditVideo")
	}

	redditVideo.SetDefaults()
	return &redditVideo, nil
}

// FindAll finds all users matching the tags.
func (users *RedditVideoStore) FindAll(ctx context.Context, tags []string, limit int) ([]domain.RedditVideo, error) {
	col := users.db.C(colRedditVideo)

	filter := bson.M{}
	if len(tags) > 0 {
		filter["tags"] = bson.M{
			"$in": tags,
		}
	}

	matches := []domain.RedditVideo{}
	if err := col.Find(filter).Limit(limit).All(&matches); err != nil {
		return nil, errors.Wrapf(err, "failed to query for users")
	}
	return matches, nil
}

// Delete removes one redditVideo identified by the name.
func (users *RedditVideoStore) Delete(ctx context.Context, name string) (*domain.RedditVideo, error) {
	col := users.db.C(colRedditVideo)

	ch := mgo.Change{
		Remove:    true,
		ReturnNew: true,
		Upsert:    false,
	}
	redditVideo := domain.RedditVideo{}
	_, err := col.Find(bson.M{"name": name}).Apply(ch, &redditVideo)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, errors.ResourceNotFound("RedditVideo", name)
		}
		return nil, err
	}

	return &redditVideo, nil
}
