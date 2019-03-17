package storage

// Storage is the generic interface for a file store
type Storage interface {
	Attributes(remotePath string) (attrs interface{}, err error)
	Cleanup() (err error)
	Delete(remotePath string) (err error)
	Download(remotePath string, localPath string) (err error)
	Init() (err error)
	GetLocation(remotePath string) (url string, err error)
	List(remotePath string) (files []interface{}, err error)
	Upload(localPath string, remotePath string) (err error)
}
