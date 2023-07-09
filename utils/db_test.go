package utils

import (
	"testing"

	"github.com/cornelk/hashmap/assert"
)

func setupDB(t *testing.T) {
	t.Helper()

	SetDBPath("../assets/db.sqlite")
}

func TestGetVertices(t *testing.T) {
	setupDB(t)
	vert := GetVertices()

	assert.Equal(t, len(vert), 30387)
}

func TestGetEdges(t *testing.T) {
	setupDB(t)
	edg := GetEdges()

	assert.Equal(t, len(edg), 831)
}
