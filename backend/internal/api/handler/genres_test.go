package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenresHandler_New(t *testing.T) {
	h := NewGenresHandler(nil)
	assert.NotNil(t, h)
}
