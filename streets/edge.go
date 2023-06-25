package streets

import "github.com/gammazero/deque"

// Edge is a struct for an edge in the graph
type Edge struct {
	ID int

	FromVertexID int
	ToVertexID   int

	Length   float64
	MaxSpeed float64

	Q deque.Deque[*Vehicle]

	Graph *Graph
}

// PushVehicle pushes a vehicle to the edge
func (e *Edge) PushVehicle(v *Vehicle) {
	e.Q.PushBack(v)
}

// FrontVehicle returns the vehicle in front of itself
func (e *Edge) FrontVehicle(vehicle *Vehicle) *Vehicle {
	idx := e.Q.Index(func(vv *Vehicle) bool {
		return vv.ID == vehicle.ID
	})

	if idx == -1 {
		return nil
	}

	return e.Q.At(idx - 1)
}
