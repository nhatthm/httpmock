package request

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeHeaders(t *testing.T) {
	t.Parallel()

	headers := Header{
		"Authorization": "Bearer token",
	}

	defaultHeaders := Header{
		"Authorization": "Bearer foobar",
		"Content-Type":  "application/json",
	}

	actual := mergeHeaders(headers, defaultHeaders)
	expected := Header{
		"Authorization": "Bearer token",
		"Content-Type":  "application/json",
	}

	assert.Equal(t, expected, actual)
}
