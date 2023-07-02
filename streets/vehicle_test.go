package streets

import (
	"fmt"
	"testing"

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

	vh1 := NewVehicle(4.5, path, &g)
	vh2 := NewVehicle(3.5, path, &g)
	vh3 := NewVehicle(2.5, path, &g)

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
}
