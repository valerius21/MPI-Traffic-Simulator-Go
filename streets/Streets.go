package streets

import (
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/redis/rueidis"
	rg "github.com/redislabs/redisgraph-go"
)

type Vertex struct {
	ID int
	X  float32
	Y  float32

	Edges []Edge
}

type Edge struct {
	ID           int
	FromVertexID int
	ToVertexID   int
	Length       float32
	MaxSpeed     float32
}

type Graph struct {
	Vertices []Vertex
	Edges    []Edge
	r        *rueidis.Client
}

type Path struct {
	Vertices []Vertex
	Edges    []Edge
}

func (g Graph) GetVertices() []Vertex {
	return g.Vertices
}

func (g Graph) FindPath(dest Vertex) Path {
	return nil

}

type RConnects struct {
	Name  string
	Osmid string
	From  int
	To    int
}

type RVertex struct {
	Highway string
	Osmid   int
	X       float32
	Y       float32
}

func New() {
	conn, _ := redis.DialURL("redis://default:valerius21@159.69.195.83:6379")
	defer conn.Close()

	graph := rg.GraphNew("traffic_0", conn)
	result, _ := graph.Query("MATCH v = (a:vertex)-[r:CONNECTS]->(b:vertex) RETURN v,r,a,b")
	for result.Next() {
		r := result.Record()
		for _, key := range r.Keys() {
			if key == "r" {
				value, _ := r.Get(key)
				rr, ok := value.(*rg.Edge)
				if !ok {
					fmt.Println("error")
					return
				}
				props := rr.Properties
				jsonData, err := json.Marshal(props)
				if err != nil {
					fmt.Println(err)
					return
				}
				//fmt.Printf("%+v\n", string(jsonData))
				var rv RConnects
				err = json.Unmarshal(jsonData, &rv)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("EDGE: %+v\n", rv)
			}
			if key == "a" || key == "b" {
				// print the key name
				// print the value
				value, _ := r.Get(key)
				rr, ok := value.(*rg.Node)
				if !ok {
					fmt.Println("error")
					return
				}
				props := rr.Properties

				jsonData, err := json.Marshal(props)
				if err != nil {
					fmt.Println(err)
					return
				}
				var rv RVertex
				err = json.Unmarshal(jsonData, &rv)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("VERTEX: %+v\n", rv)
			}
		}
		//break
	}
}
