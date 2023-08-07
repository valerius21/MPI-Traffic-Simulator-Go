package streets

import (
	"os"
	"strconv"

	"github.com/dominikbraun/graph"
	"github.com/rs/zerolog/log"
	"pchpc/utils"
)

// Point is a point in 2D space
type Point struct {
	X, Y float64
}

// Rect is a rectangle in 2D space, holding the top right and bottom left points
// and the vertices of the rectangle
type Rect struct {
	TopRight Point
	BotLeft  Point
	Vertices []JVertex
}

// InRect checks if a vertex is in a rectangle
func (r *Rect) InRect(v JVertex) bool {
	for _, vertex := range r.Vertices {
		if vertex.ID == v.ID {
			return true
		}
	}
	return false
}

// StreetGraph is a graph of streets with vertices of type int and edges of type JVertex
type StreetGraph struct {
	graph.Graph[int, JVertex]
}

// GraphBuilder is a builder for a graph
type GraphBuilder struct {
	graph          StreetGraph
	vertices       []JVertex
	edges          []JEdge
	rectangleParts int
	bot, top       Point
}

// NewGraphBuilder returns a new graph builder
func NewGraphBuilder() *GraphBuilder {
	return &GraphBuilder{}
}

// WithVertices sets the vertices of the graph
func (gb *GraphBuilder) WithVertices(vertices []JVertex) *GraphBuilder {
	gb.vertices = vertices
	return gb
}

// WithEdges sets the edges of the graph and creates a hashmap for each edge
// it also sets the Data struct of each edge
func (gb *GraphBuilder) WithEdges(edges []JEdge) *GraphBuilder {
	// new edge slice
	nEdges := make([]JEdge, len(edges))

	for _, e := range edges {
		// Nil check may be redundant
		if e.Data.Map == nil {
			// Convert max speed to float64
			msi, err := strconv.Atoi(e.MaxSpeed)
			msf := float64(msi)
			if err != nil {
				msf = 50.0 // Default max speed, aka. 'the inchident'
			}

			// Create a new map
			hMap := utils.NewMap[string, *Vehicle]()

			// Add the Data struct to the edge
			e.Data.Map = &hMap
			e.Data.MaxSpeed = msf
			e.Data.Length = e.Length
		}
		nEdges = append(nEdges, e)
	}

	gb.edges = nEdges
	return gb
}

// WithRectangleParts sets the number of rectangle parts the graph should be divided into
func (gb *GraphBuilder) WithRectangleParts(n int) *GraphBuilder {
	gb.rectangleParts = n
	return gb
}

// FromJsonBytes unmarshals the graph JSON bytes into a graph
func (gb *GraphBuilder) FromJsonBytes(jBytes []byte) *GraphBuilder {
	jGraph, err := UnmarshalGraphJSON(jBytes)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal graph JSON.")
		panic(err)
	}

	return gb.WithVertices(jGraph.Graph.Vertices).WithEdges(jGraph.Graph.Edges)
}

// FromJsonFile reads the graph JSON file and unmarshals it into a graph
func (gb *GraphBuilder) FromJsonFile(jFile string) *GraphBuilder {
	jBytes, err := os.ReadFile(jFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read graph JSON file.")
		panic(err)
	}

	return gb.FromJsonBytes(jBytes)
}

// SetTopRightBottomLeftVertices returns the top right and bottom left vertices of the graph
func (gb *GraphBuilder) SetTopRightBottomLeftVertices() *GraphBuilder {
	if len(gb.vertices) == 0 {
		log.Error().Msg("No vertices set in graph. Use WithVertices() to set vertices.")
		return gb
	}

	// Get all vertices
	vertices := gb.vertices

	botX := 100.
	botY := 100.
	topX := 0.
	topY := 0.

	for _, vertex := range vertices {
		if vertex.X < botX {
			botX = vertex.X
		}
		if vertex.Y < botY {
			botY = vertex.Y
		}
		if vertex.X > topX {
			topX = vertex.X
		}
		if vertex.Y > topY {
			topY = vertex.Y
		}
	}

	bot := Point{
		X: botX,
		Y: botY,
	}
	top := Point{
		X: topX,
		Y: topY,
	}

	log.Debug().Msgf("Bottom left vertex: %v", bot)
	log.Debug().Msgf("Top right vertex: %v", top)

	gb.bot = bot
	gb.top = top

	return gb
}

// Build builds the graph
func (gb *GraphBuilder) Build() *StreetGraph {
	vertexHash := func(vertex JVertex) int {
		return vertex.ID
	}
	g := graph.New(vertexHash, graph.Directed())

	for _, vertex := range gb.vertices {
		_ = g.AddVertex(vertex)
	}

	for _, edge := range gb.edges {
		_ = g.AddEdge(
			edge.From,
			edge.To,
			graph.EdgeData(edge.Data))
	}

	gb.graph = StreetGraph{g}
	return &gb.graph
}

// -- End of GraphBuilder --

// VertexInGraph checks if a vertex is in a graph
func (g *StreetGraph) VertexInGraph(v JVertex) bool {
	_, err := (*g).Vertex(v.ID)
	return err == nil
}
