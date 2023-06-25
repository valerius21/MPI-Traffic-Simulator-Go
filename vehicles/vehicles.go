package vehicles

// Author: Valerius Mattfeld

import (
	"pchpc/streets"
)

// Vehicle represents a vehicle in the simulation
type Vehicle struct {
	ID            string
	Speed         float64 // m/s
	QueuePosition int     // TODO: implement
	// SourceNode    *streets.Vertex
	// DestNode      *streets.Vertex
	Path streets.Path
	// Length?
}

// New creates a new vehicle
func New(path streets.Path, speed float64) Vehicle {
	//veh := Vehicle{
	//	ID:            nanoid.New(),
	//	Speed:         speed,
	//	QueuePosition: -1,
	//	Path: path
	//}
	//
	//return veh
	return Vehicle{}
}

func (v *Vehicle) step() {
}

//func (v *Vehicle) IsLeading(frontVehicle Vehicle) bool {
//	return v.QueuePosition == frontVehicle.QueuePosition-1
//}
