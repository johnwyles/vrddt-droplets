package config

// StoreMongoConfig stores the configuration for the Mongo persistence store
type StoreMongoConfig struct {
	RedditVideosCollectionName string
	URI                        string
	VrddtVideosCollectionName  string
}
