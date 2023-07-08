package streets

import (
	"fmt"

	"github.com/dominikbraun/graph"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
	"pchpc/utils"
)

// EdgeData is the data stored in an edge
type EdgeData struct {
	MaxSpeed float64
	Length   float64
	Map      *utils.HashMap[string, *Vehicle]
}

// NewGraph creates a new graph
func NewGraph() graph.Graph[int, GVertex] {
	log.Info().Msg("Creating new graph.")
	vertexHash := func(vertex GVertex) int {
		return vertex.ID
	}
	g := graph.New(vertexHash, graph.Directed())

	info, conn, err := GetRedisInfo()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get RedisInfo.")
		return g
	}
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
		}
	}(conn)

	for _, vertex := range info.Vertices {
		err := g.AddVertex(vertex)
		if err != nil {
			log.Warn().Err(err).Msg("Vertex already exists.")
			continue
		}
	}

	for _, edge := range info.Edges {
		hMap := utils.NewMap[string, *Vehicle]()
		err := g.AddEdge(
			edge.FromVertexID,
			edge.ToVertexID,
			graph.EdgeData(EdgeData{
				MaxSpeed: edge.MaxSpeed,
				Length:   edge.Length,
				Map:      &hMap,
			}))
		if err != nil {
			log.Warn().Err(err).Msg("Edge already exists.")
			continue
		}
	}

	return g
}

// GetFrontVehicleFromEdge returns the vehicle in front of the given vehicle
func GetFrontVehicleFromEdge(edge *graph.Edge[GVertex], vehicle *Vehicle) (*Vehicle, error) {
	edgeData := edge.Properties.Data.(EdgeData)

	eMap := edgeData.Map

	if eMap.Len() == 0 {
		return nil, nil
	}

	lst := eMap.ToList()

	for idx, v := range lst {
		if v.ID == vehicle.ID {
			if idx == 0 {
				return nil, nil
			}
			return lst[idx-1], nil
		}
	}
	return nil, fmt.Errorf("there was an error retrieving the front vehicle")
}
