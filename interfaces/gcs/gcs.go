package gcs

import ()

// GCS initiates a new GCS storage connection
func GCS(cfg *config.StorageGCSConfig) (stg Storage, err error) {
	l := log.With().Str("component", "storage").Str("type", "gcs").Logger()
	l.Debug().Msgf("GCS(cfg): %#v", cfg)

	stg = &gcs{
		bucketName:      cfg.Bucket,
		credentialsJSON: cfg.CredentialsJSON,
		log:             &l,
	}

	return
}

func doNothing() {}
