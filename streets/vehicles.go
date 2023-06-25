package streets

// Author: Valerius Mattfeld

import (
	"github.com/aidarkhanov/nanoid"
	"github.com/rs/zerolog/log"
)

// Vehicle represents a vehicle in the simulation
type Vehicle struct {
	ID            string
	Speed         float64 // m/s
	QueuePosition int     // TODO: implement
	Path          Path
	Graph         *Graph
	PathLength    []float64
	IsParked      bool
	// Length?
}

// NewVehicle creates a new vehicle
func NewVehicle(path Path, speed float64, graph Graph) Vehicle {
	v := Vehicle{
		ID:       nanoid.New(),
		Speed:    speed,
		Path:     path,
		Graph:    &graph,
		IsParked: false,
	}
	v.PathLength = v.GetPathLengths()
	return v
}

func (v *Vehicle) Step() {
	// vehicle is at destination
	if v.IsParked {
		return
	}
	v.drive()
}

func (v *Vehicle) drive() {
	for i := 0; i < len(v.PathLength); i++ {
		if i == len(v.PathLength)-1 {
			if v.PathLength[i] < v.Speed {
				log.Info().Msgf("Vehicle %v has reached its destination", v.ID)
				v.PathLength[i] = 0
				v.IsParked = true
				return
			} else {
				v.PathLength[i] -= v.Speed
				break
			}
		}
		if v.PathLength[i] != 0 {
			if v.PathLength[i] < v.Speed {
				v.PathLength[i+1] += v.PathLength[i] - v.Speed
				v.PathLength[i] = 0
			} else {
				v.PathLength[i] -= v.Speed
				break
			}
		}
	}
}

func (v *Vehicle) PrintInfo() {
	log.Info().Msgf("Vehicle %v: Speed=%v m/s, PathLength=%v m", v.ID, v.Speed, v.PathLength)
}

func (v *Vehicle) GetPathLengths() []float64 {
	var lengths []float64
	for i, vertex := range v.Path.Vertices {
		if i == len(v.Path.Vertices)-1 {
			continue
		}
		edge, err := v.Graph.GetCorrespondingEdge(&vertex, &v.Path.Vertices[i+1])
		if err != nil {
			log.Panic().Err(err).Msg("Failed to get corresponding edge")
		}
		lengths = append(lengths, edge.Length)
	}
	return lengths
}

//func (v *Vehicle) IsLeading(frontVehicle Vehicle) bool {
//	return v.QueuePosition == frontVehicle.QueuePosition-1
//}
