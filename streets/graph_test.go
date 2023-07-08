package streets

import (
	"testing"

	"github.com/cornelk/hashmap/assert"

	"github.com/dominikbraun/graph"
	"pchpc/utils"
)

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
		currentPosition:   0,
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
		currentPosition:   0,
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

	frontVehicle, err := GetFrontVehicleFromEdge(&e, &v2)
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
	frontVehicle, err = GetFrontVehicleFromEdge(&e, &v2)
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

	frontVehicle, err = GetFrontVehicleFromEdge(&e, &v2)
	if err != nil && frontVehicle != nil {
		t.Errorf("Error: %v", err)
	}
}
