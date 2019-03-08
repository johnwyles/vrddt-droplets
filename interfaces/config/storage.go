package config

type StorageType int

const (
	StorageConfigGCS   StorageType = iota
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
	return [...]string{"gcs", "local"}[s]
}
