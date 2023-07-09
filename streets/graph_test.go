package streets

import (
	"testing"

	"github.com/cornelk/hashmap/assert"

	"github.com/dominikbraun/graph"
)

var g graph.Graph[int, GVertex]

func setUpGraph(t *testing.T) {
	t.Helper()

	g = NewGraph("redis://default:valerius21@wutlatte.com:6379")
}

func TestGetTopRightBottomLeftVertices(t *testing.T) {
	setUpGraph(t)

	bot, top := GetTopRightBottomLeftVertices(&g)

	hasBiggerTop := false
	hasSmallerBot := false

	vertices, err := GetVertices(&g)
	if err != nil {
		t.Error(err)
	}

	for _, v := range vertices {
		if v.X > top.X && v.Y > top.Y {
			hasBiggerTop = true
		}
		if v.X < bot.X && v.Y < bot.Y {
			hasSmallerBot = true
		}
	}

	assert.True(t, !hasBiggerTop)
	assert.True(t, !hasSmallerBot)
}

func TestDivideGraph(t *testing.T) {
	setUpGraph(t)

	// Divide graph into 4 quadrants
	quadrants, err := DivideGraph(4, &g)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(quadrants), 4)

	vertices, err := GetVertices(&g)
	if err != nil {
		t.Error(err)
	}

	vertexPresent := make(map[int]bool)
	for _, v := range vertices {
		vertexPresent[v.ID] = false
	}

	for _, q := range quadrants {
		for _, v := range q.Vertices {
			vertexPresent[v.ID] = true
		}
	}

	hasFalse := false
	i := 0
	for id, v := range vertexPresent {
		if !v {
			hasFalse = true
			i++
			t.Logf("Vertex %d not present in quadrants, %d", i, id)
		}
	}

	t.Logf("Number of vertices not present in quadrants: %d of %d", i, len(vertices))
	assert.True(t, !hasFalse)
}
