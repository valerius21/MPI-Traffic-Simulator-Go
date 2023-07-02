package streets

import (
	"fmt"

	"github.com/aidarkhanov/nanoid"
	"github.com/dominikbraun/graph"
	"github.com/rs/zerolog/log"
)

type Vehicle struct {
	ID          string
	Path        []int
	Speed       float64
	Graph       graph.Graph[int, GVertex]
	IsParked    bool
	PathLengths []float64
}

func (v *Vehicle) getPathLengths() error {
	lengthsArray := make([]float64, 0)
	for i, vertex := range v.Path {
		if i == len(v.Path)-1 {
			break
		}
		edge, err := v.Graph.Edge(vertex, v.Path[i+1])
		if err != nil {
			log.Error().Err(err).Msg("Failed to get edge.")
			return err
		}

		length := edge.Properties.Data.(EdgeData).Length
		lengthsArray = append(lengthsArray, length)
	}
	v.PathLengths = lengthsArray
	return nil
}

func NewVehicle(speed float64, path []int, graph graph.Graph[int, GVertex]) Vehicle {
	v := Vehicle{
		ID:    nanoid.New(),
		Path:  path,
		Speed: speed,
		Graph: graph,
	}
	err := v.getPathLengths()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get path lengths.")
		return Vehicle{}
	}
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
}

func (v *Vehicle) PrintInfo() {
	log.Info().
		Str("id", v.ID).
		Bool("isParked", v.IsParked).
		Float64("speed", v.Speed).
		Str("path lengths", fmt.Sprintf("%v", v.PathLengths)).
		Msg("Vehicle info")
}
