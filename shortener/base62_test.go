package shortener

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestBase62Encode(t *testing.T) {
	tts := []struct {
		in  uint64
		out string
	}{
		{0, "0"},
		{42, "G"},
		{1234567890, "1ly7vk"},
		{math.MaxUint64, "lYGhA16ahyf"},
	}

	for _, tt := range tts {
		got := Base62Encode(tt.in)
		assert.Equal(t, tt.out, string(got))
	}
}
