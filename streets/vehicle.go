package streets

import (
	"fmt"
	"math"
	"sort"

	"pchpc/utils"

	"github.com/aidarkhanov/nanoid"
	"github.com/dominikbraun/graph"
	"github.com/rs/zerolog/log"
)

// Vehicle is a vehicle
type Vehicle struct {
	ID                string                     `json:"id,omitempty"`
	Path              []int                      `json:"path,omitempty"`
	DistanceTravelled float64                    `json:"distance_travelled,omitempty"`
	Speed             float64                    `json:"speed,omitempty"`
	g                 *graph.Graph[int, JVertex] `json:"g,omitempty"`
	IsParked          bool                       `json:"is_parked,omitempty"`
	PathLengths       []float64                  `json:"path_lengths,omitempty"`
	PathLimit         float64                    `json:"path_limit,omitempty"`
}

// getPathLengths calculates the length of each edge in the path
func (v *Vehicle) getPathLengths() error {
	lengthsArray := make([]float64, 0)
	sum := 0.0
	for i, vertex := range v.Path {
		if i == len(v.Path)-1 {
			break
		}
		edge, err := (*v.g).Edge(vertex, v.Path[i+1])
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
func (v *Vehicle) getCurrentEdge() (*graph.Edge[JVertex], error) {
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
func (v *Vehicle) getEdgeByIndex(index int) (oEdge *graph.Edge[JVertex], err error) {
	if index == len(v.Path)-1 {
		return oEdge, fmt.Errorf("index is out of range")
	}

	ed, err := (*v.g).Edge(v.Path[index], v.Path[index+1])
	if err != nil {
		log.Error().Err(err).Msg("Failed to get edge.")
		return &ed, err
	}

	return &ed, nil
}

// getHashMapByEdge returns the hashmap of the given edge
func (v *Vehicle) getHashMapByEdge(edge *graph.Edge[JVertex]) (*utils.HashMap[string, *Vehicle], error) {
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

// AddVehicleToEdge adds the vehicle to the given hashmap
func (v *Vehicle) AddVehicleToEdge(edge *graph.Edge[JVertex]) error {
	edgeData := edge.Properties.Data.(EdgeData)
	msEdgeSpeed := edgeData.MaxSpeed / 3.6
	hashMap := edgeData.Map
	if v.isInMap(hashMap) {
		return nil
	}

	frontVehicle, err := v.GetFrontVehicleFromEdge(edge)
	if err != nil {
		return err
	}

	if frontVehicle != nil && frontVehicle.Speed < v.Speed {
		v.Speed = frontVehicle.Speed
	} else if frontVehicle != nil && frontVehicle.Speed > v.Speed && msEdgeSpeed > v.Speed {
		minAcceleration := 0.1
		maxAcceleration := 0.5
		v.Speed += utils.RandomFloat64(minAcceleration, maxAcceleration)
	}

	hashMap.Set(v.ID, v)
	v.updateVehiclePosition(hashMap)
	return nil
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
func NewVehicle(speed float64, path []int, graph *graph.Graph[int, JVertex]) Vehicle {
	v := Vehicle{
		ID:                nanoid.New(),
		Path:              path,
		Speed:             speed,
		g:                 graph,
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
	log.Debug().Msgf("Current index: %d, delta: %f", idx, delta)
	log.Debug().Msgf("Current path: %v", v.Path)
	log.Debug().Msgf("Current vehicle: %v", v)
	edge, err := v.getEdgeByIndex(idx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get edge.")
		return
	}

	if v.Speed >= delta && idx != 0 {
		isInGraph := VertexInGraph(v.g, edge.Target)
		if !isInGraph {
			panic("not in graph")
		}
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
	if err != nil {
		log.Error().Err(err).Msg("Failed to get hashmap.")
		return
	}
	err = v.AddVehicleToEdge(edge)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add vehicle to map.")
		return
	}
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

// GetFrontVehicleFromEdge returns the vehicle in front of the given vehicle
func (v *Vehicle) GetFrontVehicleFromEdge(edge *graph.Edge[JVertex]) (*Vehicle, error) {
	edgeData := edge.Properties.Data.(EdgeData)

	eMap := edgeData.Map

	if eMap.Len() < 1 {
		return nil, nil
	}

	lst := eMap.ToList()

	sort.Slice(lst, func(i, j int) bool {
		return lst[i].DistanceTravelled > lst[j].DistanceTravelled
	})

	var frontIndex int

	for i, vh := range lst {
		if v.ID == vh.ID && i < 0 {
			frontIndex = i - 1
		}
	}

	return lst[frontIndex], nil
}
