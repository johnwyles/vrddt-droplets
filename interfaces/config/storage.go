package config

// StorageType is the type of storage
type StorageType int

const (
	// StorageConfigGCS is the type reserved for GCS storage
	StorageConfigGCS StorageType = iota

	// StorageConfigS3 is the type reserved for S3 storage
	StorageConfigS3 StorageType = iota

	// StorageConfigLocal is the type reserved for local storage
	StorageConfigLocal StorageType = iota
)

// StorageConfig holds all the different implementations for cloud storage
type StorageConfig struct {
	GCS   StorageGCSConfig
	Local StorageLocalConfig
	Type  StorageType
}

// String will return the string representation of the iota
func (s StorageType) String() string {
	return [...]string{"gcs", "s3", "local"}[s]
}
