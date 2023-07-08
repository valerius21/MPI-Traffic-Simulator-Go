package streets

import (
	"fmt"
	"testing"

	"github.com/rs/zerolog"

	"pchpc/utils"

	"github.com/cornelk/hashmap/assert"

	"github.com/dominikbraun/graph"
)

func TestVehicle_Step(t *testing.T) {
	g := NewGraph(utils.GetRedisURL())

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

func TestVehicle_AddVehicleToMap(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	// update speed test
	g := NewGraph(utils.GetRedisURL())
	// src := GVertex{ID: 2617388513}
	// dst := GVertex{ID: 2290171245}
	path := []int{2617388513, 2290171245}

	vh := NewVehicle(4.0, path, &g)
	vh2 := NewVehicle(6.0, path, &g)

	vh.Step()  // ensure vehicle is present
	vh.Step()  // ensure vehicle is present
	vh2.Step() // ensure vehicle is present

	edge, _ := vh.getCurrentEdge()

	l := edge.Properties.Data.(EdgeData).Map.Len()

	assert.Equal(t, 2, l)

	//for {
	//	vh.Step()  // ensure vehicle is present
	//	vh2.Step() // ensure vehicle is present
	//	if vh.IsParked && vh2.IsParked {
	//		break
	//	}
	//}

	assert.Equal(t, vh.Speed, vh2.Speed)

	// existingVehicle := NewVehicle(2.0, nil, &g)
}

func TestGetFrontVehicleFromEdge(t *testing.T) {
	hm := utils.NewMap[string, *Vehicle]()
	emptyHm := utils.NewMap[string, *Vehicle]()
	lonleyHm := utils.NewMap[string, *Vehicle]()

	v1 := Vehicle{
		ID:                "test_front",
		Path:              nil,
		DistanceTravelled: 0,
		Speed:             1.0,
		Graph:             nil,
		IsParked:          false,
		PathLengths:       nil,
		PathLimit:         0,
	}

	v2 := Vehicle{
		ID:                "test_back",
		Path:              nil,
		DistanceTravelled: 0,
		Speed:             1.0,
		Graph:             nil,
		IsParked:          false,
		PathLengths:       nil,
		PathLimit:         0,
	}
	hm.Set(v1.ID, &v1)
	hm.Set(v2.ID, &v2)
	lonleyHm.Set(v1.ID, &v1)

	e := graph.Edge[GVertex]{
		Source: GVertex{ID: 0},
		Target: GVertex{ID: 1},
		Properties: graph.EdgeProperties{
			Attributes: nil,
			Weight:     0,
			Data: EdgeData{
				MaxSpeed: 10,
				Length:   10,
				Map:      &hm,
			},
		},
	}

	frontVehicle, err := v2.GetFrontVehicleFromEdge(&e)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	assert.Equal(t, frontVehicle.ID, v1.ID)
	e = graph.Edge[GVertex]{
		Source: GVertex{ID: 0},
		Target: GVertex{ID: 1},
		Properties: graph.EdgeProperties{
			Attributes: nil,
			Weight:     0,
			Data: EdgeData{
				MaxSpeed: 10,
				Length:   10,
				Map:      &emptyHm,
			},
		},
	}
	frontVehicle, err = v2.GetFrontVehicleFromEdge(&e)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	if frontVehicle != nil {
		t.Errorf("Expected nil, got %v", frontVehicle)
	}

	e = graph.Edge[GVertex]{
		Source: GVertex{ID: 0},
		Target: GVertex{ID: 1},
		Properties: graph.EdgeProperties{
			Attributes: nil,
			Weight:     0,
			Data: EdgeData{
				MaxSpeed: 10,
				Length:   10,
				Map:      &lonleyHm,
			},
		},
	}

	frontVehicle, err = v2.GetFrontVehicleFromEdge(&e)
	if err != nil && frontVehicle != nil {
		t.Errorf("Error: %v", err)
	}
}
