package streets_test

import (
	"testing"

	"github.com/gammazero/deque"
	"github.com/stretchr/testify/assert"
	"pchpc/streets"
)

func TestVehicle_GetCurrentEdge(t *testing.T) {
	vertex1 := streets.Vertex{1, nil, nil}
	vertex2 := streets.Vertex{2, nil, nil}
	edge := streets.Edge{
		ID:           0,
		FromVertexID: 1,
		ToVertexID:   2,
		Length:       100,
		MaxSpeed:     0,
		Q:            deque.Deque[*streets.Vehicle]{},
		Graph:        nil,
	}
	g := streets.Graph{
		Vertices: []streets.Vertex{vertex1, vertex2},
		Edges:    []streets.Edge{edge},
		Rdb:      nil,
	}
	vertex1.Graph = &g
	vertex2.Graph = &g
	edge.Graph = &g
	pathLengths := []float64{100}

	vh1 := streets.Vehicle{
		ID:         "vh1",
		Speed:      10,
		PathLength: pathLengths,
		Path: streets.Path{
			StartVertex: &vertex1,
			EndVertex:   &vertex2,
			Vertices:    []streets.Vertex{vertex1, vertex2},
		},
		Graph: &g,
	}
	edge.PushVehicle(&vh1)
	testID := vh1.GetCurrentEdge().ID

	assert.Equal(t, edge.ID, testID, "Expected vh1 to be on edge")
}
