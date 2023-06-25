package streets

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
	"github.com/rs/zerolog/log"
)

// MaxEdges is the maximum number of edges that can be added to the graph.
const MaxEdges = 1_000_000_000

// Graph is a struct for a graph
type Graph struct {
	Vertices []Vertex
	Edges    []Edge
	Rdb      *rg.Graph
}

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

func (g *Graph) GetCorrespondingEdge(src, dest *Vertex) (*Edge, error) {
	neighbours, err := g.GetNeighbours(src)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to get neighbours of vertex with ID %d.", src.ID)
		return &Edge{}, err
	}

	for edge := range neighbours {
		if edge.FromVertexID == src.ID && edge.ToVertexID == dest.ID {
			return &edge, nil
		}
	}

	return &Edge{}, errors.New("no edge found")
}

// FindPath finds the shortest path between two vertices in the graph.
// Maybe this could be converted to A* in the future.
func (g *Graph) FindPath(src, dest *Vertex) (Path, error) {
	log.Info().Msgf("Finding path from vertex with ID %d to vertex with ID %d.", src.ID, dest.ID)

	// make sure both vertices exist in the graph
	_, err := g.GetVertexByID(src.ID)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to find vertex with ID %d.", src.ID)
		return Path{}, err
	}
	_, err = g.GetVertexByID(dest.ID)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to find vertex with ID %d.", dest.ID)
		return Path{}, err
	}

	// find the shortest path between the two vertices
	visited := make(map[int]bool) // int is the vertex ID
	queue := [][]Vertex{{*src}}

	visited[src.ID] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		node := path[len(path)-1]

		if node.ID == dest.ID {
			return Path{
				StartVertex: src,
				EndVertex:   dest,
				Vertices:    path,
			}, nil
		}

		neighbours, err := g.GetNeighbours(&node)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to get neighbours of vertex with ID %d.", node.ID)
			return Path{}, err
		}

		for _, neighbour := range neighbours {
			if !visited[neighbour.ID] {
				visited[neighbour.ID] = true
				newPath := make([]Vertex, len(path))
				copy(newPath, path)
				newPath = append(newPath, neighbour)
				queue = append(queue, newPath)
			}
		}
	}

	return Path{}, nil
}

// GetNeighbours returns a map of all the neighbours of a vertex.
func (g *Graph) GetNeighbours(src *Vertex) (map[Edge]Vertex, error) {
	neighbours := make(map[Edge]Vertex)

	for _, edge := range g.Edges {
		if edge.FromVertexID == src.ID {
			neighbour, err := g.GetVertexByID(edge.ToVertexID)
			if err != nil {
				log.Panic().Err(err).Msgf("Failed to get vertex with ID %d", edge.ToVertexID)
				return nil, err
			}
			neighbours[edge] = *neighbour
		}
	}

	return neighbours, nil
}

// AddVertex adds a new vertex to the graph if a vertex with the same ID doesn't exist already.
func (g *Graph) AddVertex(v Vertex) error {
	log.Info().Msgf("Adding vertex with ID %d to graph.", v.ID)

	for _, vertex := range g.Vertices {
		if vertex.ID == v.ID {
			log.Warn().Msgf("Vertex with ID %d already exists in the graph.", v.ID)
			return nil // errors.New("vertex already exists in the graph")
		}
	}

	g.Vertices = append(g.Vertices, v)
	log.Info().Msgf("Successfully added vertex with ID %d to graph.", v.ID)
	return nil
}

// AddEdge adds a new edge to the graph if an edge with the same ID doesn't exist already.
func (g *Graph) AddEdge(e Edge) error {
	log.Info().Msgf("Adding edge with ID %d to graph.", e.ID)

	for _, edge := range g.Edges {
		if edge.ID == e.ID {
			log.Warn().Msgf("Edge with ID %d already exists in the graph.", e.ID)
			return nil // errors.New("edge already exists in the graph")
		}
	}

	g.Edges = append(g.Edges, e)

	// Updating the vertices with the new edge
	for i, v := range g.Vertices {
		if v.ID == e.FromVertexID || v.ID == e.ToVertexID {
			g.Vertices[i].Edges = append(v.Edges, e)
		}
	}

	log.Info().Msgf("Successfully added edge with ID %d to graph.", e.ID)
	return nil
}

// GetVertexByID returns a pointer to a vertex with the given ID.
func (g *Graph) GetVertexByID(id int) (*Vertex, error) {
	for _, v := range g.Vertices {
		if v.ID == id {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("vertex with ID %d not found", id)
}

// GetEdgeByID returns a pointer to an edge with the given ID.
func (g *Graph) GetEdgeByID(id int) (*Edge, error) {
	for _, e := range g.Edges {
		if e.ID == id {
			return &e, nil
		}
	}
	return nil, fmt.Errorf("edge with ID %d not found", id)
}

// New returns a new Graph, by querying the RedisGraph database.
func New() (Graph, redis.Conn, error) {
	log.Info().Msg("Initializing new graph and connecting to RedisGraph database.")
	g := Graph{}
	conn, err := redis.DialURL("redis://default:valerius21@159.69.195.83:6379")
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to RedisGraph database.")
		return g, conn, err
	}

	nGraph := rg.GraphNew("traffic_1", conn)
	g.Rdb = &nGraph
	graph := g.Rdb
	result, err := graph.Query("MATCH v = (a:vertex)-[r:CONNECTS]->(b:vertex) RETURN v,r,a,b")
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute query on RedisGraph database.")
		return g, conn, err
	}

	for result.Next() {
		r := result.Record()
		for _, key := range r.Keys() {
			// Process edges
			if key == "r" {
				value, exists := r.Get(key)
				if !exists {
					log.Error().Msg("Failed to get edge record from result set.")
					return g, conn, fmt.Errorf("failed to get edge record from result set: %v", err)
				}

				rr, ok := value.(*rg.Edge)
				if !ok {
					log.Error().Msg("Failed to assert type of edge record.")
					return g, conn, errors.New("type assertion error")
				}

				props := rr.Properties
				jsonData, err := json.Marshal(props)
				if err != nil {
					log.Error().Err(err).Msg("Failed to marshal properties of edge record.")
					return g, conn, err
				}

				var rv rConnects
				err = json.Unmarshal(jsonData, &rv)
				if err != nil {
					log.Error().Err(err).Msg("Failed to unmarshal JSON data into rConnects struct.")
					return g, conn, err
				}

				// generate a random numeric ID for the edge, because openstreetmap
				// could provide multiple edges with the same ID
				intID := rand.Intn(MaxEdges)

				for {
					exists, _ := g.GetEdgeByID(intID)
					if exists == nil {
						break
					} else {
						intID = rand.Intn(1_000_000_000)
					}
				}

				// convert speed from string to float64
				speed, err := strconv.ParseFloat(rv.MaxSpeed, 64)

				// Add to graph
				e := Edge{
					ID:           intID,
					FromVertexID: rv.From,
					ToVertexID:   rv.To,
					Length:       rv.Length,
					MaxSpeed:     speed,
					Graph:        &g,
				} // Additional fields need to be set accordingly
				err = g.AddEdge(e)
				if err != nil {
					log.Error().Err(err).Msg("Failed to add edge to graph.")
					return Graph{}, nil, err
				}
			}

			// Process vertices
			if key == "a" || key == "b" {
				value, exists := r.Get(key)
				if !exists {
					log.Error().Msg("Failed to get vertex record from result set.")
					return g, conn, fmt.Errorf("failed to get vertex record from result set: %v", err)
				}

				rr, ok := value.(*rg.Node)
				if !ok {
					log.Error().Msg("Failed to assert type of vertex record.")
					return g, conn, errors.New("type assertion error")
				}

				props := rr.Properties
				jsonData, err := json.Marshal(props)
				if err != nil {
					log.Error().Err(err).Msg("Failed to marshal properties of vertex record.")
					return g, conn, err
				}

				var rv rVertex
				err = json.Unmarshal(jsonData, &rv)
				if err != nil {
					log.Error().Err(err).Msg("Failed to unmarshal JSON data into rVertex struct.")
					return g, conn, err
				}

				// Add to graph
				v := Vertex{
					ID: rv.OsmID,
					// X: rv.X,
					// Y: rv.Y,
					Graph: &g,
				} // Additional fields need to be set accordingly
				err = g.AddVertex(v)
				if err != nil {
					log.Error().Err(err).Msg("Failed to add vertex to graph.")
					return Graph{}, nil, err
				}
			}
		}
	}

	log.Info().Msg("Successfully created a new graph from RedisGraph database.")
	return g, conn, nil
}
