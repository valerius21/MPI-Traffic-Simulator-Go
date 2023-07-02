package streets

import (
	"fmt"
	"math"
	"sort"

	"github.com/cornelk/hashmap"

	"github.com/aidarkhanov/nanoid"
	"github.com/dominikbraun/graph"
	"github.com/rs/zerolog/log"
)

type Vehicle struct {
	ID                string
	Path              []int
	DistanceTravelled float64
	Speed             float64
	Graph             *graph.Graph[int, GVertex]
	IsParked          bool
	PathLengths       []float64
	PathLimit         float64
	currentPosition   int
	currentEdge       *graph.Edge[GVertex]
}

func (v *Vehicle) getPathLengths() error {
	g := *v.Graph
	lengthsArray := make([]float64, 0)
	sum := 0.0
	for i, vertex := range v.Path {
		if i == len(v.Path)-1 {
			break
		}
		edge, err := g.Edge(vertex, v.Path[i+1])
		if err != nil {
			log.Error().Err(err).Msg("Failed to get edge.")
			return err
		}

		length := edge.Properties.Data.(EdgeData).Length
		lengthsArray = append(lengthsArray, length)
		sum += length
	}
	v.PathLengths = lengthsArray
	v.PathLimit = sum
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

	g := *v.Graph
	edge, err = g.Edge(v.Path[index], v.Path[index+1])
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

func (v *Vehicle) AddVehicleToMap(hashMap *hashmap.Map[int, *Vehicle]) {
	_, loaded := hashMap.GetOrInsert(hashMap.Len(), v)
	if loaded {
		return
	}
	v.updateVehiclePosition(hashMap)
}

func (v *Vehicle) RemoveVehicleFromMap(hashMap *hashmap.Map[int, *Vehicle]) {
	hashMap.Del(v.currentPosition)
	v.updateVehiclePosition(hashMap)
}

func (v *Vehicle) updateVehiclePosition(hashMap *hashmap.Map[int, *Vehicle]) {
	positions := make(map[float64]*Vehicle)
	// sort vehicles by distance travelled
	for i := 0; i < hashMap.Len(); i++ {
		vehicle, exists := hashMap.Get(i)
		if !exists {
			log.Error().Msg("Failed to get vehicle.")
			return
		}
		v := vehicle
		_, delta := v.deductCurrentPathVertexIndex()

		// if delta is already in map, add a small delta
		for positions[delta] != nil {
			delta += 0.00000001
		}

		positions[delta] = v
		hashMap.Del(i)
	}

	keys := make([]float64, 0, len(positions))
	for k := range positions {
		keys = append(keys, k)
	}

	sort.Float64s(keys)

	for i := len(keys) - 1; i >= 0; i-- {
		vehicle := positions[keys[i]]
		hashMap.Set(i, vehicle)
		if vehicle.ID == v.ID {
			v.currentPosition = i
		}
	}
}

func (v *Vehicle) String() string {
	return fmt.Sprintf("Vehicle: %s, Speed: %f, Path: %v", v.ID, v.Speed, v.Path)
}

func NewVehicle(speed float64, path []int, graph *graph.Graph[int, GVertex]) Vehicle {
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
	idx, delta := v.deductCurrentPathVertexIndex()
	edge, err := v.getEdgeByIndex(idx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get edge.")
		return
	}

	if v.Speed >= delta && idx != 0 {
		oldEdge, err := v.getEdgeByIndex(idx - 1)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get edge.")
			return
		}

		oldHashMap, err := v.getHashMapByEdge(oldEdge)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get hashmap.")
			return
		}
		v.RemoveVehicleFromMap(oldHashMap)
	}

	hashMap, err := v.getHashMapByEdge(edge)
	v.AddVehicleToMap(hashMap)
	// vehicle is at destination
	if v.IsParked {
		return
	}
	v.drive()
	v.updateVehiclePosition(hashMap)
}

func (v *Vehicle) drive() {
	v.DistanceTravelled += v.Speed
}

func (v *Vehicle) PrintInfo() {
	log.Info().
		Str("id", v.ID).
		Bool("isParked", v.IsParked).
		Float64("speed", v.Speed).
		Str("path lengths", fmt.Sprintf("%v", v.PathLengths)).
		Msg("Vehicle info")
}
