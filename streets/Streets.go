package streets

import (
	"container/heap"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
	"github.com/rs/zerolog/log"
)

type Vertex struct {
	ID int
	X  float32
	Y  float32

	Edges []Edge
	Graph *Graph
}

type Edge struct {
	ID int

	FromVertexID int
	ToVertexID   int

	Length   float32
	MaxSpeed float32

	Graph *Graph
}

type Graph struct {
	Vertices []Vertex
	Edges    []Edge
	Rdb      *rg.Graph
}

type Path struct {
	StartVertex *Vertex
	EndVertex   *Vertex
	Vertices    []Vertex
	Edges       []Edge
}

// RConnects is a struct for the RedisGraph database edge
type RConnects struct {
	Name  string
	OsmID string
	From  int
	To    int
}

// RVertex is a struct for the RedisGraph database vertex
type RVertex struct {
	Highway string
	OsmID   int
	X       float32
	Y       float32
}

type Item struct {
	value    *Vertex // The value of the item; arbitrary.
	priority float32 // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the lowest, not highest, priority so we use less than here.
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value *Vertex, priority float32) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

type VertexDistances struct {
	dist  float32
	prev  *Vertex
	index int
}

func (g *Graph) FindPath(src, dest *Vertex) (Path, error) {
	// Initialize distances and previous vertices
	dist := make(map[int]float32)
	prev := make(map[int]*Vertex)
	h := &PriorityQueue{}
	heap.Init(h)

	for _, v := range g.Vertices {
		if v.ID == src.ID {
			dist[v.ID] = 0
			heap.Push(h, &VertexDistances{dist: 0, index: v.ID})
		} else {
			dist[v.ID] = math.MaxFloat32
		}
		prev[v.ID] = nil
	}

	for h.Len() > 0 {
		item := heap.Pop(h).(*VertexDistances)
		u, err := g.GetVertexByID(item.index)
		if err != nil {
			fmt.Printf("There was an error %v", err)
			return Path{}, err
		}

		for _, e := range u.Edges {
			alt := dist[u.ID] + e.Length
			if alt < dist[e.ToVertexID] {
				dist[e.ToVertexID] = alt
				prev[e.ToVertexID] = u
				heap.Push(h, &VertexDistances{dist: alt, index: e.ToVertexID})
			}
		}
	}

	// Reconstruct the path
	u := dest
	path := Path{}
	for u != nil {
		path.Vertices = append([]Vertex{*u}, path.Vertices...)
		if prev[u.ID] != nil {
			for _, e := range prev[u.ID].Edges {
				if e.ToVertexID == u.ID {
					path.Edges = append([]Edge{e}, path.Edges...)
				}
			}
		}
		u = prev[u.ID]
	}

	return path, nil
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

func (g *Graph) GetVertexByID(id int) (*Vertex, error) {
	for _, v := range g.Vertices {
		if v.ID == id {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("vertex with ID %d not found", id)
}

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

	nGraph := rg.GraphNew("traffic_0", conn)
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

				var rv RConnects
				err = json.Unmarshal(jsonData, &rv)
				if err != nil {
					log.Error().Err(err).Msg("Failed to unmarshal JSON data into RConnects struct.")
					return g, conn, err
				}

				intID, err := strconv.Atoi(rv.OsmID)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to convert OSM ID to integer. Skipping edge.")
					continue
				}

				// Add to graph
				e := Edge{ID: intID, FromVertexID: rv.From, ToVertexID: rv.To, Graph: &g} // Additional fields need to be set accordingly
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

				var rv RVertex
				err = json.Unmarshal(jsonData, &rv)
				if err != nil {
					log.Error().Err(err).Msg("Failed to unmarshal JSON data into RVertex struct.")
					return g, conn, err
				}

				// Add to graph
				v := Vertex{ID: rv.OsmID, X: rv.X, Y: rv.Y, Graph: &g} // Additional fields need to be set accordingly
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
