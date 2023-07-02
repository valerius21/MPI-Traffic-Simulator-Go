package streets

import (
	"fmt"
	"github.com/cornelk/hashmap"
	"math"

	"github.com/aidarkhanov/nanoid"
	"github.com/dominikbraun/graph"
	"github.com/rs/zerolog/log"
)

type Vehicle struct {
	ID                string
	Path              []int
	DistanceTravelled float64
	Speed             float64
	Graph             graph.Graph[int, GVertex]
	IsParked          bool
	PathLengths       []float64
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

func (v *Vehicle) deductCurrentPathVertexIndex() (index int, delta float64) {
	tmpDistance := v.DistanceTravelled
	for i, pathLength := range v.PathLengths {
		tmpDistance -= pathLength
		if tmpDistance < 0 {
			delta = math.Abs(tmpDistance)
			index = i
			return index, delta
		}
	}

	return 0, 0.0
}

func (v *Vehicle) getEdgeByIndex(index int) (edge graph.Edge[GVertex], err error) {
	if index == len(v.Path)-1 {
		return edge, fmt.Errorf("index is out of range")
	}

	edge, err = v.Graph.Edge(v.Path[index], v.Path[index+1])
	if err != nil {
		log.Error().Err(err).Msg("Failed to get edge.")
		return edge, err
	}

	return edge, nil
}

func (v *Vehicle) getHashMapByEdge(edge graph.Edge[GVertex]) (*hashmap.Map[int, *Vehicle], error) {
	data, exists := edge.Properties.Data.(EdgeData)
	if !exists {
		err := fmt.Errorf("edge data is not of type EdgeData")
		log.Error().Err(err).Msg("Failed to get data from edge.")
		return nil, err
	}
	return data.Map, nil
}

func (v *Vehicle) String() string {
	return fmt.Sprintf("Vehicle: %s, Speed: %f, Path: %v", v.ID, v.Speed, v.Path)
}

func NewVehicle(speed float64, path []int, graph graph.Graph[int, GVertex]) Vehicle {
	v := Vehicle{
		ID:                nanoid.New(),
		Path:              path,
		Speed:             speed,
		Graph:             graph,
		DistanceTravelled: 0.0,
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
