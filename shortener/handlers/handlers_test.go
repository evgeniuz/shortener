package handlers

import (
	"encoding/json"
	"errors"
	"github.com/evgeniuz/shortener/shortener/store/bolt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

const testDBPath = "/tmp/test-shortener.db"

func TestShortener(t *testing.T) {
	ts, teardown := prepare(t)
	defer teardown()

	hash := ""

	t.Run("shorten link", func(t *testing.T) {
		res, err := http.Post(ts.URL, "", strings.NewReader("{\"url\": \"https://example.com\"}"))
		require.NoError(t, err)

		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode, )

		result := struct {
			Data struct {
				Hash string `json:"hash"`
			} `json:"data"`
		}{}

		err = json.NewDecoder(res.Body).Decode(&result)
		require.NoError(t, err)

		hash = result.Data.Hash

		assert.NotEqual(t, "", hash)
	})

	t.Run("access link", func(t *testing.T) {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		res, err := client.Get(ts.URL + "/" + hash)
		require.NoError(t, err)

		defer res.Body.Close()

		assert.Equal(t, http.StatusMovedPermanently, res.StatusCode)
		assert.Equal(t, "https://example.com", res.Header.Get("Location"))
	})

	t.Run("access non existing link", func(t *testing.T) {
		res, err := http.Get(ts.URL + "/not-found")
		require.NoError(t, err)

		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})

	t.Run("get stats", func(t *testing.T) {
		assert.Eventually(t, func() bool {
			res, err := http.Get(ts.URL + "/" + hash + "/stats")
			require.NoError(t, err)

			defer res.Body.Close()

			assert.Equal(t, res.StatusCode, http.StatusOK)

			result := struct {
				Data struct {
					Day   int `json:"day"`
					Week  int `json:"week"`
					Total int `json:"total"`
				} `json:"data"`
			}{}

			err = json.NewDecoder(res.Body).Decode(&result)
			require.NoError(t, err)

			return result.Data.Day == 1 && result.Data.Week == 1 && result.Data.Total == 1
		}, 5 * time.Second, 100 * time.Millisecond)
	})
}

func TestRenderSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		renderSuccess(res, "example")
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	require.NoError(t, err)

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, res.StatusCode, http.StatusOK)
	assert.JSONEq(t, "{\"success\": true, \"data\": \"example\"}", string(body))
}

func TestRenderError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		renderError(res, errors.New("error"))
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	require.NoError(t, err)

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, res.StatusCode, http.StatusInternalServerError)
	assert.JSONEq(t, "{\"success\": false, \"message\": \"error\"}", string(body))
}

func prepare(t *testing.T) (*httptest.Server, func()) {
	_ = os.Remove(testDBPath)

	s, err := bolt.NewStore(testDBPath, 7)
	assert.NoError(t, err)

	shortener, err := NewShortener(s)
	assert.NoError(t, err)

	ts := httptest.NewServer(shortener.router())

	teardown := func() {
		ts.Close()
		require.NoError(t, s.Close())
		_ = os.Remove(testDBPath)
	}

	return ts, teardown
}
