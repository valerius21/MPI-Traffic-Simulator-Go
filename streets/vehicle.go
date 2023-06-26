package streets

import (
	"github.com/aidarkhanov/nanoid"
	"github.com/dominikbraun/graph"
)

type Vehicle struct {
	ID               string
	Path             []int
	Speed            float64
	Graph            graph.Graph[int, GVertex]
	IsParked         bool
	DistanceTraveled float64
}

func NewVehicle(speed float64, path []int, graph graph.Graph[int, GVertex]) Vehicle {
	return Vehicle{
		ID:               nanoid.New(),
		Path:             path,
		Speed:            speed,
		Graph:            graph,
		DistanceTraveled: 0.0,
	}
}

func (v *Vehicle) Step() {
	// vehicle is at destination
	if v.IsParked {
		return
	}
	v.drive()
}

func (v *Vehicle) drive() {
	// vehicle is at destination
	if v.IsParked {
		return
	}
	// vehicle is at the end of the path
	if len(v.Path) == 1 {
		v.IsParked = true
		return
	}

	//// vehicle is at the end of the current edge
	//if v.DistanceTraveled >= v.Graph.EdgeData(v.Path[0], v.Path[1]).(EdgeData).Length {
	//	v.DistanceTraveled = 0.0
	//	v.Path = v.Path[1:]
	//	return
	//}

	// vehicle is in the middle of the current edge
	v.DistanceTraveled += v.Speed
}

func (v *Vehicle) PrintInfo() {
}
