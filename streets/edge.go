package streets

import (
	"sync"
)

// Edge is a struct for an edge in the graph
type Edge struct {
	ID int

	FromVertexID int
	ToVertexID   int

	Length   float64
	MaxSpeed float64

	Q *ThreadSafeDeque[*Vehicle]

	Graph *Graph
	sync.Mutex
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

func (e *Edge) getIndex(v *Vehicle) int {
	e.Lock()
	defer e.Unlock()
	for i := 0; i < e.Q.Len(); i++ {
		if e.Q.At(i) == v {
			return i
		}
	}
	return -1
}

func (e *Edge) GetPosition(sourceVehicle *Vehicle) int {
	return e.getIndex(sourceVehicle)
}

// FrontVehicle returns the vehicle in front of itself
func (e *Edge) FrontVehicle(sourceVehicle *Vehicle) *Vehicle {
	idx := e.getIndex(sourceVehicle)

	if idx == -1 {
		return nil
	}

	return e.Q.At(idx - 1)
}
