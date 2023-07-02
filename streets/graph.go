package streets

import (
	"github.com/cornelk/hashmap"
	"github.com/dominikbraun/graph"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
)

type EdgeData struct {
	MaxSpeed float64
	Length   float64
	Map      *hashmap.Map[int, *Vehicle]
}

func VerticesListToEdgesList[K comparable, T any](g graph.Graph[K, T], vertices []K) ([]graph.Edge[T], error) {
	edges := make([]graph.Edge[T], 0)

	for i, vertex := range vertices {
		if i == len(vertices)-1 {
			break
		}
		edge, err := g.Edge(vertex, vertices[i+1])
		if err != nil {
			log.Error().Err(err).Msg("Failed to get edge.")
			return nil, err
		}
		edges = append(edges, edge)
	}

	return edges, nil
}

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
		err := g.AddEdge(
			edge.FromVertexID,
			edge.ToVertexID,
			graph.EdgeData(EdgeData{
				MaxSpeed: edge.MaxSpeed,
				Length:   edge.Length,
				Map:      hashmap.New[int, *Vehicle](),
			}))
		if err != nil {
			log.Warn().Err(err).Msg("Edge already exists.")
			continue
		}
	}

	return g
}
