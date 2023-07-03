package streets

import (
	"fmt"
	"math"

	"pchpc/utils"

	"github.com/aidarkhanov/nanoid"
	"github.com/dominikbraun/graph"
	"github.com/rs/zerolog/log"
)

// Vehicle is a vehicle
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
}

// getPathLengths calculates the length of each edge in the path
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

// getCurrentEdge returns the current edge the vehicle is on
func (v *Vehicle) getCurrentEdge() (*graph.Edge[GVertex], error) {
	idx, _ := v.deductCurrentPathVertexIndex()
	edge, err := v.getEdgeByIndex(idx)
	if err != nil {
		return nil, err
	}
	return edge, nil
}

// deductCurrentPathVertexIndex returns the index of the current edge in the path
func (v *Vehicle) deductCurrentPathVertexIndex() (index int, delta float64) {
	tmpDistance := v.DistanceTravelled

	for i, length := range v.PathLengths {
		if tmpDistance < length {
			return i, math.Abs(tmpDistance)
		}
		tmpDistance -= length
	}

	return 0, 0.0
}

// getEdgeByIndex returns the edge at the given index
func (v *Vehicle) getEdgeByIndex(index int) (oEdge *graph.Edge[GVertex], err error) {
	if index == len(v.Path)-1 {
		return oEdge, fmt.Errorf("index is out of range")
	}

	g := *v.Graph
	ed, err := g.Edge(v.Path[index], v.Path[index+1])
	if err != nil {
		log.Error().Err(err).Msg("Failed to get edge.")
		return &ed, err
	}

	return &ed, nil
}

// getHashMapByEdge returns the hashmap of the given edge
func (v *Vehicle) getHashMapByEdge(edge *graph.Edge[GVertex]) (*utils.HashMap[string, *Vehicle], error) {
	data, exists := edge.Properties.Data.(EdgeData)
	if !exists {
		err := fmt.Errorf("edge data is not of type EdgeData")
		log.Error().Err(err).Msg("Failed to get data from edge.")
		return nil, err
	}
	return data.Map, nil
}

// isInMap checks if the vehicle is in the given hashmap
func (v *Vehicle) isInMap(hashMap *utils.HashMap[string, *Vehicle]) bool {
	_, exists := hashMap.Get(v.ID)
	return exists
}

// AddVehicleToMap adds the vehicle to the given hashmap
func (v *Vehicle) AddVehicleToMap(hashMap *utils.HashMap[string, *Vehicle]) {
	if v.isInMap(hashMap) {
		return
	}
	hashMap.Set(v.ID, v)
	v.updateVehiclePosition(hashMap)
}

// RemoveVehicleFromMap removes the vehicle from the given hashmap
func (v *Vehicle) RemoveVehicleFromMap(hashMap *utils.HashMap[string, *Vehicle]) {
	if hashMap.Len() == 0 {
		v.updateVehiclePosition(hashMap)
		return
	}
	hashMap.Del(v.ID)
	v.updateVehiclePosition(hashMap)
}

// updateVehiclePosition updates the vehicle position
func (v *Vehicle) updateVehiclePosition(hashMap *utils.HashMap[string, *Vehicle]) {
	if v.PathLimit <= v.DistanceTravelled {
		v.IsParked = true
		edge, err := v.getCurrentEdge()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get current edge.")
			return
		}
		hashMap, err := v.getHashMapByEdge(edge)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get hashmap.")
			return
		}

		hashMap.Del(v.ID)
	}

	log.Debug().Msgf("Current vehicles on edge: %d, %s", hashMap.Len(), v.ID)
}

// String returns the string representation of the vehicle
func (v *Vehicle) String() string {
	return fmt.Sprintf("Vehicle: %s, Speed: %f, Distance Travelled: %v Sum: %.2f", v.ID, v.Speed,
		v.DistanceTravelled, v.PathLimit)
}

// NewVehicle creates a new vehicle
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

// Step moves the vehicle one step forward
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

// drive moves the vehicle forward
func (v *Vehicle) drive() {
	v.DistanceTravelled += v.Speed
}

// PrintInfo prints the vehicle info
func (v *Vehicle) PrintInfo() {
	log.Debug().
		Str("id", v.ID).
		Bool("isParked", v.IsParked).
		Float64("speed", v.Speed).
		Str("path lengths", fmt.Sprintf("%v", v.PathLengths)).
		Msg("Vehicle info")
}
