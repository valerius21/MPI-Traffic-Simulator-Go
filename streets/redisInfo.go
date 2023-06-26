package streets

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
)

// rConnects is a struct for the RedisGraph database edge
type rConnects struct {
	Name     string  `json:"name,omitempty"`
	OsmID    string  `json:"osmid"`
	From     int     `json:"u"`
	To       int     `json:"v"`
	MaxSpeed string  `json:"maxspeed"`
	Length   float64 `json:"length"`
}

// rVertex is a struct for the RedisGraph database vertex
type rVertex struct {
	Highway string
	OsmID   int
	X       float32
	Y       float32
}

type GEdge struct {
	ID           int
	FromVertexID int
	ToVertexID   int
	Length       float64
	MaxSpeed     float64
}
type GVertex struct {
	ID int
}

type RedisInfo struct {
	Edges    []GEdge
	Vertices []GVertex
}

// GetRedisInfo returns a new Graph, by querying the RedisGraph database.
func GetRedisInfo() (RedisInfo, redis.Conn, error) {
	log.Info().Msg("Initializing new graph and connecting to RedisGraph database.")
	conn, err := redis.DialURL("redis://default:valerius21@159.69.195.83:6379")
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to RedisGraph database.")
		return RedisInfo{}, conn, err
	}

	nGraph := rg.GraphNew("traffic_1", conn)
	rdb := nGraph
	result, err := rdb.Query("MATCH v = (a:vertex)-[r:CONNECTS]->(b:vertex) RETURN v,r,a,b")
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute query on RedisGraph database.")
		return RedisInfo{}, conn, err
	}

	edges := make([]GEdge, 0)
	vertices := make([]GVertex, 0)

	for result.Next() {
		r := result.Record()
		for _, key := range r.Keys() {
			// Process edges
			if key == "r" {
				value, exists := r.Get(key)
				if !exists {
					log.Error().Msg("Failed to get edge record from result set.")
					return RedisInfo{}, conn, fmt.Errorf("failed to get edge record from result set: %v", err)
				}

				rr, ok := value.(*rg.Edge)
				if !ok {
					log.Error().Msg("Failed to assert type of edge record.")
					return RedisInfo{}, conn, errors.New("type assertion error")
				}

				props := rr.Properties
				jsonData, err := json.Marshal(props)
				if err != nil {
					log.Error().Err(err).Msg("Failed to marshal properties of edge record.")
					return RedisInfo{}, conn, err
				}

				var rv rConnects
				err = json.Unmarshal(jsonData, &rv)
				if err != nil {
					log.Error().Err(err).Msg("Failed to unmarshal JSON data into rConnects struct.")
					return RedisInfo{}, conn, err
				}

				// generate a random numeric ID for the edge, because openstreetmap
				// could provide multiple edges with the same ID
				intID := rand.Intn(1_000_000_000)

				// convert speed from string to float64
				speed, err := strconv.ParseFloat(rv.MaxSpeed, 64)

				// Add to graph
				e := GEdge{
					ID:           intID,
					FromVertexID: rv.From,
					ToVertexID:   rv.To,
					Length:       rv.Length,
					MaxSpeed:     speed,
				} // Additional fields need to be set accordingly

				edges = append(edges, e)
			}

			// Process vertices
			if key == "a" || key == "b" {
				value, exists := r.Get(key)
				if !exists {
					log.Error().Msg("Failed to get vertex record from result set.")
					return RedisInfo{}, conn, fmt.Errorf("failed to get vertex record from result set: %v", err)
				}

				rr, ok := value.(*rg.Node)
				if !ok {
					log.Error().Msg("Failed to assert type of vertex record.")
					return RedisInfo{}, conn, errors.New("type assertion error")
				}

				props := rr.Properties
				jsonData, err := json.Marshal(props)
				if err != nil {
					log.Error().Err(err).Msg("Failed to marshal properties of vertex record.")
					return RedisInfo{}, conn, err
				}

				var rv rVertex
				err = json.Unmarshal(jsonData, &rv)
				if err != nil {
					log.Error().Err(err).Msg("Failed to unmarshal JSON data into rVertex struct.")
					return RedisInfo{}, conn, err
				}

				// Add to graph
				v := GVertex{
					ID: rv.OsmID,
					// X: rv.X,
					// Y: rv.Y,
				} // Additional fields need to be set accordingly
				vertices = append(vertices, v)
			}
		}
	}

	log.Info().Msg("Successfully created a new graph from RedisGraph database.")
	return RedisInfo{
		Edges:    edges,
		Vertices: vertices,
	}, conn, nil
}
