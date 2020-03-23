package store

type Stats struct {
	Day, Week, Total uint64
}

type Store interface {
	Set(url string) (string, error)
	Get(hash string) (string, error)

	Visit(hash string) error
	Stats(hash string) (Stats, error)

	Close() error
}
