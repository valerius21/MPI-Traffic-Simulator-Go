package streets

import (
	"errors"
	"os"
	"strconv"

	"github.com/aidarkhanov/nanoid"

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
	// ID is the ID of the graph. Root graph has ID 0
	ID string

	// RootGraph is the root graph of the graph, nil if the graph is the root graph
	RootGraph *StreetGraph

	// graph is the graph
	Graph graph.Graph[int, JVertex]
}

// GraphBuilder is a builder for a graph
type GraphBuilder struct {
	graph                StreetGraph
	vertices             []JVertex
	edges                []JEdge
	rectangleParts, pick int
	bot, top             Point
	rects                []Rect
	pickedRect           Rect
	id                   string
	root                 *StreetGraph
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

func (gb *GraphBuilder) NumberOfRects(n int) *GraphBuilder {
	gb.rectangleParts = n
	return gb
}

// DivideGraphsIntoRects divides the graph into n parts. Column-wise division.
func (gb *GraphBuilder) DivideGraphsIntoRects() *GraphBuilder {
	if gb.top == (Point{}) || gb.bot == (Point{}) {
		gb.SetTopRightBottomLeftVertices()
	}
	if gb.rectangleParts == 0 {
		gb.rectangleParts = 1
	}

	n := gb.rectangleParts
	rootBot := gb.bot
	rootTop := gb.top

	// Get all vertices
	vertices := gb.vertices

	rects := make([]Rect, n)

	xDelta := rootTop.X - rootBot.X

	for i := 0; i < n; i++ {
		botX := rootBot.X + (xDelta/float64(n))*float64(i)
		topX := rootBot.X + (xDelta/float64(n))*float64(i+1)

		rects[i] = Rect{
			TopRight: Point{
				X: topX,
				Y: rootTop.Y,
			},
			BotLeft: Point{
				X: botX,
				Y: rootBot.Y,
			},
			Vertices: make([]JVertex, 0),
		}

		for _, vertex := range vertices {
			isInYInterval := vertex.Y >= rootBot.Y && vertex.Y <= rootTop.Y
			isInXInterval := vertex.X >= botX && vertex.X <= topX

			if isInYInterval && isInXInterval {
				rects[i].Vertices = append(rects[i].Vertices, vertex)
			}
		}
	}

	gb.rects = rects

	return gb
}

// PickRect picks a rectangle from the graph
func (gb *GraphBuilder) PickRect(i int) *GraphBuilder {
	if gb.rects == nil {
		gb.DivideGraphsIntoRects()
	}

	n := len(gb.rects)

	if i >= n {
		log.Error().Msgf("Rectangle index out of bounds. Max index: %d", n-1)
		return gb
	}

	gb.pickedRect = gb.rects[i]

	return gb
}

// FilterForRect filters the graph for the picked rectangle
func (gb *GraphBuilder) FilterForRect() *GraphBuilder {
	rect := gb.pickedRect
	filteredEdges := make([]JEdge, 0)

	// filter for coordinates in rect
	for _, edge := range gb.edges {
		src := edge.From
		dst := edge.To

		srcInRect := false
		dstInRect := false

		for _, vertex := range rect.Vertices {
			if vertex.ID == src {
				srcInRect = true
			}
			if vertex.ID == dst {
				dstInRect = true
			}
		}

		if srcInRect && dstInRect {
			filteredEdges = append(filteredEdges, edge)
		}
	}

	// filter for vertices in rect
	filteredVertices := make([]JVertex, 0)

	for _, vertex := range gb.vertices {
		for _, rectVertex := range rect.Vertices {
			if vertex.ID == rectVertex.ID {
				filteredVertices = append(filteredVertices, vertex)
			}
		}
	}

	gb.edges = filteredEdges
	gb.vertices = filteredVertices

	return gb
}

func (gb *GraphBuilder) IsRoot() *GraphBuilder {
	gb.id = "root"
	return gb
}

func (gb *GraphBuilder) IsLeaf(root *StreetGraph) *GraphBuilder {
	gb.id = nanoid.New()
	gb.root = root
	return gb
}

func (gb *GraphBuilder) check() error {
	// Verify that the graph can be built
	if gb.vertices == nil {
		log.Error().Msg("No vertices set in graph. Use WithVertices() to set vertices.")
		return errors.New("no vertices set in graph")
	}

	if gb.edges == nil {
		log.Error().Msg("No edges set in graph. Use WithEdges() to set edges.")
		return errors.New("no edges set in graph")
	}

	if gb.bot == (Point{}) || gb.top == (Point{}) {
		log.Error().Msg("No top or bottom vertices set in graph. Use SetTopRightBottomLeftVertices() to set vertices.")
		return errors.New("no top or bottom vertices set in graph")
	}

	if gb.rectangleParts == 0 {
		log.Error().Msg("No rectangle parts set in graph. Use NumberOfRects() to set number of parts.")
		return errors.New("no rectangle parts set in graph")
	}

	if gb.rects == nil {
		log.Error().Msg("No rectangles set in graph. Use DivideGraphsIntoRects() to divide graph into rectangles.")
		return errors.New("no rectangles set in graph")
	}

	if gb.pickedRect.Vertices == nil {
		log.Error().Msg("No rectangle or Vertex picked in graph. Use PickRect() to pick a rectangle.")
		return errors.New("no rectangle/vertex picked in graph")
	}

	if gb.id == "" {
		log.Error().Msg("No id set in graph. Use IsRoot() or IsLeaf() to set id.")
		return errors.New("no id set in graph")
	}

	return nil
}

// Build builds the graph
func (gb *GraphBuilder) Build() (*StreetGraph, error) {
	if err := gb.check(); err != nil {
		return nil, err
	}

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

	gb.graph = StreetGraph{
		ID:        gb.id,
		RootGraph: gb.root,
		Graph:     g,
	}

	return &gb.graph, nil
}

// -- End of GraphBuilder --

// VertexInGraph checks if a vertex is in a graph
func (g *StreetGraph) VertexInGraph(v JVertex) bool {
	_, err := (*g).Graph.Vertex(v.ID)
	return err == nil
}
