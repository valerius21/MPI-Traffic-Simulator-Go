package streets

import (
	"fmt"
	"testing"

	"github.com/cornelk/hashmap/assert"

	"github.com/dominikbraun/graph"
)

func TestVehicle_Step(t *testing.T) {
	g := NewGraph()

	_, err := g.Edges()
	if err != nil {
		panic(err)
	}

	path, err := graph.ShortestPath(g, 2617388513, 1247500404)
	if err != nil {
		panic(err)
	}

	vh1 := NewVehicle(4.0, path, &g)
	vh2 := NewVehicle(3.0, path, &g)
	vh3 := NewVehicle(2.0, path, &g)

	vh1.currentPosition = 0
	vh1.currentPosition = 1
	vh1.currentPosition = 2

	for {
		vh1.Step()
		vh2.Step()
		vh3.Step()
		fmt.Println(vh1.String())
		fmt.Println(vh2.String())
		fmt.Println(vh3.String())
		if vh1.IsParked && vh2.IsParked && vh3.IsParked {
			break
		}
	}
	// assert
	edge, err := vh1.getCurrentEdge()
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 0, edge.Properties.Data.(EdgeData).Map.Len())

	// assert
	edge, err = vh2.getCurrentEdge()
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 0, edge.Properties.Data.(EdgeData).Map.Len())
	// assert
	edge, err = vh3.getCurrentEdge()
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 0, edge.Properties.Data.(EdgeData).Map.Len())
}
