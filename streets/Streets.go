package streets

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	rg "github.com/redislabs/redisgraph-go"
)

type Vertex struct {
	ID int
	X  float32
	Y  float32

	Edges []Edge
	Graph *Graph
}

type Edge struct {
	ID           int
	FromVertexID int
	ToVertexID   int
	Length       float32
	MaxSpeed     float32
	Graph        *Graph
}

type Graph struct {
	Vertices []Vertex
	Edges    []Edge
	Rdb      *rg.Graph
}

type Path struct {
	Vertices []Vertex
	Edges    []Edge
}

func (g Graph) FindPath(src, dest *Vertex) Path {
	query := fmt.Sprintf(`MATCH (startNode), (endNode)
    WHERE ID(startNode) = %d AND ID(endNode) = %d
    RETURN shortestPath((startNode)-[:CONNECTS*]->(endNode)) AS path
    ORDER BY length(path) ASC`, src.ID, dest.ID)

	result, _ := g.Rdb.Query(query)
	for result.Next() {
		r := result.Record()
		for _, key := range r.Keys() {
			value, _ := r.Get(key)
			fmt.Println(key, value)
		}
	}
	return Path{}
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

// New returns a new Graph, by querying the RedisGraph database.
func New() (Graph, redis.Conn, error) {
	g := Graph{}
	conn, _ := redis.DialURL("redis://default:valerius21@159.69.195.83:6379")

	nGraph := rg.GraphNew("traffic_0", conn)
	g.Rdb = &nGraph
	graph := g.Rdb
	result, _ := graph.Query("MATCH v = (a:vertex)-[r:CONNECTS]->(b:vertex) RETURN v,r,a,b")
	for result.Next() {
		r := result.Record()
		for _, key := range r.Keys() {
			if key == "r" {
				value, _ := r.Get(key)
				rr, ok := value.(*rg.Edge)
				if !ok {
					fmt.Println("error")
					return Graph{}, conn, errors.New("error")
				}
				props := rr.Properties
				jsonData, err := json.Marshal(props)
				if err != nil {
					fmt.Println(err)
					return Graph{}, conn, err
				}
				var rv RConnects
				err = json.Unmarshal(jsonData, &rv)
				if err != nil {
					fmt.Println(err)
					return Graph{}, conn, err
				}
			}
			if key == "a" || key == "b" {
				value, _ := r.Get(key)
				rr, ok := value.(*rg.Node)
				if !ok {
					fmt.Println("error")
					return Graph{}, conn, errors.New("error")
				}
				props := rr.Properties

				jsonData, err := json.Marshal(props)
				if err != nil {
					fmt.Println(err)
					return Graph{}, conn, err
				}
				var rv RVertex
				err = json.Unmarshal(jsonData, &rv)
				if err != nil {
					fmt.Println(err)
					return Graph{}, conn, err
				}
			}
		}
	}
	return g, conn, nil
}
