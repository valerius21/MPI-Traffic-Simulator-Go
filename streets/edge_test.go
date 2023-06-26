package streets_test

import (
	"testing"

	"github.com/gammazero/deque"
	"github.com/stretchr/testify/assert"
	"pchpc/streets"
)

func TestEdge_PushVehicle(t *testing.T) {
	edge := streets.Edge{
		ID:           0,
		FromVertexID: 0,
		ToVertexID:   0,
		Length:       100,
		MaxSpeed:     0,
		Q:            &deque.Deque[*streets.Vehicle]{},
		Graph:        nil,
	}

	pathLengths := &deque.Deque[float64]{}

	vh1 := streets.Vehicle{
		ID:         "vh1",
		Speed:      10,
		PathLength: pathLengths,
	}

	vh2 := streets.Vehicle{
		ID:         "vh2",
		Speed:      10,
		PathLength: pathLengths,
	}

	edge.PushVehicle(&vh1)
	assert.Equal(t, "vh1", edge.Q.At(0).ID, "Expected vh1 to be at the front of the queue")
	assert.Equal(t, 1, edge.Q.Len(), "Expected queue length to be 1")

	edge.PushVehicle(&vh2)
	assert.Equal(t, "vh2", edge.Q.At(1).ID, "Expected vh2 to be at the back of the queue")
	assert.Equal(t, 2, edge.Q.Len(), "Expected queue length to be 2")
}

func TestEdge_FrontVehicle(t *testing.T) {
	edge := streets.Edge{
		ID:           0,
		FromVertexID: 0,
		ToVertexID:   0,
		Length:       100,
		MaxSpeed:     0,
		Q:            &deque.Deque[*streets.Vehicle]{},
		Graph:        nil,
	}

	pathLengths := &deque.Deque[float64]{}

	vh1 := streets.Vehicle{
		ID:         "vh1",
		Speed:      10,
		PathLength: pathLengths,
	}

	vh2 := streets.Vehicle{
		ID:         "vh2",
		Speed:      10,
		PathLength: pathLengths,
	}

	edge.Q.PushBack(&vh1)
	edge.Q.PushBack(&vh2)
	vFront := edge.Q.Front()
	assert.Equal(t, "vh1", vFront.ID, "Expected vh1 to be at the front of the queue")
	vFront2 := edge.FrontVehicle(&vh2)
	assert.Equal(t, "vh1", vFront2.ID, "Expected vh1 to be at the front of the queue")
}
