package streets

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"

	"github.com/stretchr/testify/assert"
)

func TestAddVertex(t *testing.T) {
	g := Graph{}

	v := Vertex{ID: 1}
	err := g.AddVertex(v)
	assert.Nil(t, err, "Expected no error adding vertex")
	assert.Contains(t, g.Vertices, v, "Graph should contain added vertex")
}

func TestAddEdge(t *testing.T) {
	g := Graph{}
	v1 := Vertex{ID: 1}
	v2 := Vertex{ID: 2}
	g.AddVertex(v1)
	g.AddVertex(v2)

	e := Edge{ID: 1, FromVertexID: 1, ToVertexID: 2, Length: 1.0, MaxSpeed: 1.0}
	err := g.AddEdge(e)
	assert.Nil(t, err, "Expected no error adding edge")
	assert.Contains(t, g.Edges, e, "Graph should contain added edge")
}

func TestGetVertexByID(t *testing.T) {
	g := Graph{}

	v := Vertex{ID: 1}
	g.AddVertex(v)

	vPtr, err := g.GetVertexByID(1)
	assert.Nil(t, err, "Expected no error getting vertex by ID")
	assert.Equal(t, v, *vPtr, "Returned vertex should match added vertex")
}

func TestGetEdgeByID(t *testing.T) {
	g := Graph{}
	v1 := Vertex{ID: 1}
	v2 := Vertex{ID: 2}
	g.AddVertex(v1)
	g.AddVertex(v2)

	e := Edge{ID: 1, FromVertexID: 1, ToVertexID: 2, Length: 1.0, MaxSpeed: 1.0}
	g.AddEdge(e)

	ePtr, err := g.GetEdgeByID(1)
	assert.Nil(t, err, "Expected no error getting edge by ID")
	assert.Equal(t, e, *ePtr, "Returned edge should match added edge")
}

func TestFindPath(t *testing.T) {
	g := Graph{}
	v1 := Vertex{ID: 1}
	v2 := Vertex{ID: 2}
	g.AddVertex(v1)
	g.AddVertex(v2)

	e := Edge{ID: 1, FromVertexID: 1, ToVertexID: 2, Length: 1.0, MaxSpeed: 1.0}
	g.AddEdge(e)

	path, err := g.FindPath(&v1, &v2)
	assert.Nil(t, err, "Expected no error finding path")
	assert.Equal(t, v1, *path.StartVertex, "Start vertex should match")
	assert.Equal(t, v2, *path.EndVertex, "End vertex should match")
	assert.Equal(t, v1.ID, path.Vertices[0].ID, "Path vertices should match")
	assert.Equal(t, v2.ID, path.Vertices[1].ID, "Path vertices should match")
	assert.Equal(t, 2, len(path.Vertices), "Path should have 2 vertices")
}

func TestPathDirection(t *testing.T) {
	g, conn, err := New()

	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			log.Panic().Err(err).Msg("Failed to close Redis connection")
		}
	}(conn)

	if err != nil {
		panic(err)
	}

	// path, err := g.FindPath(&Vertex{ID: 28127535}, &Vertex{ID: 208640196})
	path, err := g.FindPath(&Vertex{ID: 60347877}, &Vertex{ID: 73066996})

	assert.NotNil(t, path, "Path should not be nil")
	assert.Nil(t, err, "Expected no error finding path")

	// verify that the path is in the correct direction
	// the vertex ID corresponds to the Edge.FromVertexID
	// the vertex ID from the next vertex in the path corresponds to the Edge.ToVertexID
	for idx, v := range path.Vertices {
		if v.ID == 60347877 || v.ID == 73066996 {
			continue
		}

		correspondingEdge, err := g.GetCorrespondingEdge(&v, &path.Vertices[idx+1])

		assert.Nil(t, err, "Expected no error getting corresponding edge")
		assert.NotNil(t, correspondingEdge, "Edge ID should be greater than or equal to 0")
	}
}
