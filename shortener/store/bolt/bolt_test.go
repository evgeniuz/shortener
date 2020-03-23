package bolt

import (
	"github.com/evgeniuz/shortener/shortener/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

const testDBPath = "/tmp/test-shortener.db"

func TestBoltStore_SetGet(t *testing.T) {
	s, teardown := prepare(t)
	defer teardown()

	exampleUrl := "https://example.com"
	differentUrl := "https://example.org"

	// check that non-existing hash returns empty
	url, err := s.Get("does-not-exist")
	assert.NoError(t, err)
	assert.Equal(t, "", url)

	// create three shortened urls
	hashExample, err := s.Set(exampleUrl)
	assert.NoError(t, err)
	assert.NotZero(t, len(hashExample))

	hashDifferent, err := s.Set(differentUrl)
	assert.NoError(t, err)
	assert.NotZero(t, len(hashDifferent))

	hashSame, err := s.Set(exampleUrl)
	assert.NoError(t, err)
	assert.NotZero(t, len(hashSame))

	// check that hashes are different
	assert.NotEqual(t, hashExample, hashDifferent)
	assert.NotEqual(t, hashExample, hashSame)

	// check that they can be reversed
	got, err := s.Get(hashExample)
	assert.NoError(t, err)
	assert.Equal(t, exampleUrl, got)

	got, err = s.Get(hashDifferent)
	assert.NoError(t, err)
	assert.Equal(t, differentUrl, got)

	got, err = s.Get(hashSame)
	assert.NoError(t, err)
	assert.Equal(t, exampleUrl, got)
}

func TestBoltStore_VisitStats(t *testing.T) {
	s, teardown := prepare(t)
	defer teardown()

	hash := "example"
	now := time.Now().UTC()

	// add visit more than 1 week ago (counts only to total)
	err := s.(*Store).visitTimestamp(hash, now.AddDate(-1, 0, 0))
	assert.NoError(t, err)

	stats, err := s.Stats(hash)
	assert.NoError(t, err)
	assert.Equal(t, store.Stats{0, 0, 1}, stats)

	// add visit more than 1 day ago (counts to week and total)
	err = s.(*Store).visitTimestamp(hash, now.AddDate(0, 0, -3))
	assert.NoError(t, err)

	stats, err = s.Stats(hash)
	assert.NoError(t, err)
	assert.Equal(t, store.Stats{0, 1, 2}, stats)

	// add few visits with same timestamp
	err = s.(*Store).visitTimestamp(hash, now)
	assert.NoError(t, err)

	err = s.(*Store).visitTimestamp(hash, now)
	assert.NoError(t, err)

	stats, err = s.Stats(hash)
	assert.NoError(t, err)
	assert.Equal(t, store.Stats{2, 3, 4}, stats)

	// add single visit
	err = s.Visit(hash)
	assert.NoError(t, err)

	stats, err = s.Stats(hash)
	assert.NoError(t, err)
	assert.Equal(t, store.Stats{3, 4, 5}, stats)
}

func prepare(t *testing.T) (store.Store, func()) {
	_ = os.Remove(testDBPath)

	s, err := NewStore(testDBPath, 7)
	assert.NoError(t, err)

	teardown := func() {
		require.NoError(t, s.Close())
		_ = os.Remove(testDBPath)
	}

	return s, teardown
}
