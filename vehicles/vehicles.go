package vehicles

import "github.com/aidarkhanov/nanoid"

type Vehicle struct {
	// length  int
	ID            string
	X             float32
	Y             float32
	Speed         float32
	QueuePosition int
	SourceNodeID  int
	DestNodeID    int
}

func New(source int, dest int) Vehicle {
	veh := Vehicle{}
	// assign a random vehicle.id
	veh.ID = nanoid.New()
	veh.X = -1
	veh.Y = -1
	veh.Speed = -1
	veh.QueuePosition = -1
	veh.SourceNodeID = source
	veh.DestNodeID = dest

	return veh
}

func (v *Vehicle) IsLeading(frontVehicle Vehicle) bool {
	return v.QueuePosition == frontVehicle.QueuePosition-1
}
