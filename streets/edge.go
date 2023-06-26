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

func (e *Edge) PopVehicle() {
	if e.Q.Len() > 0 {
		e.Q.PopFront()
	}
}

func (e *Edge) GetPosition(sourceVehicle *Vehicle) int {
	idx := e.Q.Index(func(vv *Vehicle) bool {
		return vv.ID == sourceVehicle.ID
	})
	return idx
}

// FrontVehicle returns the vehicle in front of itself
func (e *Edge) FrontVehicle(sourceVehicle *Vehicle) *Vehicle {
	idx := e.Q.Index(func(vv *Vehicle) bool {
		return vv.ID == sourceVehicle.ID
	})

	if idx == -1 {
		return nil
	}

	return e.Q.At(idx - 1)
}
