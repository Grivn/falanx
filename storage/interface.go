package storage

// Storage
type Storage interface {
	// StoreState stores key-value to non-volatile memory.
	StoreState(key string, value []byte) error
	// DelState deletes data with specified key from non-volatile memory.
	DelState(key string) error
	// ReadState retrieves data with specified key from non-volatile memory.
	ReadState(key string) ([]byte, error)
	// ReadState retrieves data with specified key prefix from non-volatile memory.
	ReadStateSet(key string) (map[string][]byte, error)
	// Destroy clears the non-volatile memory.
	Destroy(key string) error
}
