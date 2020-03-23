package bolt

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/evgeniuz/shortener/shortener"
	"github.com/evgeniuz/shortener/shortener/store"
	"math"
	"math/rand"
	"time"
)

type Store struct {
	db      *bolt.DB
	maxHash int
}

var (
	urlsBucket  = []byte("urls")
	statsBucket = []byte("stats")
)

const MAX_TRIES = 3

func NewStore(path string, maxHashLength uint) (store.Store, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(urlsBucket)
		if err != nil {
			return fmt.Errorf("failed to create urls bucket: %w", err)
		}

		_, err = tx.CreateBucketIfNotExists(statsBucket)
		if err != nil {
			return fmt.Errorf("failed to create stats bucket: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init database: %w", err)
	}

	maxHash := int(math.Pow(62, float64(maxHashLength)) - 1)

	rand.Seed(time.Now().UnixNano())

	return &Store{db, maxHash}, nil
}

func (s *Store) randomHash() []byte {
	return shortener.Base62Encode(uint64(rand.Intn(s.maxHash)))
}

func (s *Store) Set(url string) (string, error) {
	var hash []byte

	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(urlsBucket)

		for i := 0; i < MAX_TRIES; i++ {
			h := s.randomHash()

			url := b.Get(h)
			if url != nil {
				continue
			}

			hash = h
			break
		}

		if hash == nil {
			return errors.New("failed to create hash")
		}

		return b.Put(hash, []byte(url))
	})
	if err != nil {
		return "", fmt.Errorf("failed to store url: %w", err)
	}

	return string(hash), nil
}

func (s *Store) Stats(hash string) (store.Stats, error) {
	stats := store.Stats{0, 0, 0}

	now := time.Now().UTC()

	day := []byte(now.AddDate(0, 0, -1).Format(time.RFC3339))
	week := []byte(now.AddDate(0, 0, -7).Format(time.RFC3339))

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(statsBucket).Bucket([]byte(hash))
		if b == nil {
			return nil
		}

		c := b.Cursor()
		k, v := c.Last()

		for ; k != nil && bytes.Compare(k, day) >= 0; k, v = c.Prev() {
			stats.Day += binary.BigEndian.Uint64(v)
		}

		stats.Week = stats.Day
		for ; k != nil && bytes.Compare(k, week) >= 0; k, v = c.Prev() {
			stats.Week += binary.BigEndian.Uint64(v)
		}

		stats.Total = stats.Week
		for ; k != nil; k, v = c.Prev() {
			stats.Total += binary.BigEndian.Uint64(v)
		}

		return nil
	})
	if err != nil {
		return store.Stats{}, fmt.Errorf("failed to gather stats: %w", err)
	}

	return stats, nil
}

func (s *Store) visitTimestamp(hash string, timestamp time.Time) error {
	key := []byte(timestamp.Format(time.RFC3339))
	visits := uint64(1)

	err := s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.Bucket(statsBucket).CreateBucketIfNotExists([]byte(hash))
		if err != nil {
			return fmt.Errorf("failed to create stats bucket: %w", err)
		}

		current := b.Get(key)
		if current != nil {
			visits += binary.BigEndian.Uint64(current)
		}

		updated := make([]byte, 8)
		binary.BigEndian.PutUint64(updated, visits)

		err = b.Put(key, updated)
		if err != nil {
			return fmt.Errorf("failed to update stats: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to register visit: %w", err)
	}

	return nil
}

func (s *Store) Visit(hash string) error {
	return s.visitTimestamp(hash, time.Now().UTC())
}

func (s *Store) Get(hash string) (string, error) {
	var url string

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(urlsBucket)

		url = string(b.Get([]byte(hash)))

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	return url, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
