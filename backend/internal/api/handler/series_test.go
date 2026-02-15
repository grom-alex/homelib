package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeriesHandler_New(t *testing.T) {
	h := NewSeriesHandler(nil)
	assert.NotNil(t, h)
}
