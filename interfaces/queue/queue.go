package queue

const (
	// Client is a connection type which generally will push to the queue
	Client ConnectionType = iota
	// Consumer is a connection type which generally will pop from the queue
	Consumer ConnectionType = iota
)

// ConnectionType is the representation of the type of connection (Consumer or Client)
type ConnectionType int

// Queue is the generic interface for a queue
type Queue interface {
	Cleanup() (err error)
	Init() (err error)
	MakeClient() (err error)
	MakeConsumer() (err error)
	Push(msg interface{}) (err error)
	Pop() (msg interface{}, err error)
}

func (c ConnectionType) String() string {
	return [...]string{"client", "consumer"}[c]
}
