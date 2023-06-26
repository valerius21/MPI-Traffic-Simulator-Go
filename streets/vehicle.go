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
	for i, pathLength := range v.PathLengths {
		isLastElement := i == len(v.PathLengths)-1
		secondVertex := v.Path[i+1]
		firstVertex := v.Path[i]
		edge, err := v.Graph.Edge(firstVertex, secondVertex)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get edge.")
			return
		}
		q := edge.Properties.Data.(EdgeData).Deque
		if !q.Exists(v) {
			q.PushFront(v)
			log.Info().Msgf("init q: %v", q.Len())
		}

		// end
		if isLastElement && pathLength <= v.Speed {
			v.PathLengths[i] = 0.
			v.IsParked = true
			lastVertex := v.Path[i]
			beforeLastVertex := v.Path[i-1]
			edge, err := v.Graph.Edge(beforeLastVertex, lastVertex)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get edge.")
				return
			}
			q := edge.Properties.Data.(EdgeData).Deque
			if q.Len() > 0 {
				q.PopBack()
				log.Info().Msgf("End q: %v", q.Len())
			} else {
				log.Error().Msg("Deque is empty.")
			}
			return
		} else if pathLength <= v.Speed && pathLength != 0. { // new edge
			// maybe avoid additional step?
			length := pathLength
			v.PathLengths[i] = 0.
			v.PathLengths[i+1] += length

			lastVertex := v.Path[i+1]
			beforeLastVertex := v.Path[i]
			edge, err := v.Graph.Edge(beforeLastVertex, lastVertex)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get edge.")
				return
			}
			q := edge.Properties.Data.(EdgeData).Deque
			q.PopBack()
			log.Info().Msgf("Vehicle %s is on edge %d -> %d", v.ID, beforeLastVertex, lastVertex)
			log.Info().Msgf("New Edge %v", q.Len())
			return
		} else if pathLength > v.Speed { // current edge
			v.PathLengths[i] -= v.Speed
			return
		}
	}
}

func (v *Vehicle) PrintInfo() {
	log.Info().
		Str("id", v.ID).
		Bool("isParked", v.IsParked).
		Float64("speed", v.Speed).
		Str("path lengths", fmt.Sprintf("%v", v.PathLengths)).
		Msg("Vehicle info")
}
