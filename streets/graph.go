package streets

import (
	"math/big"

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
func NewGraph(redisURL string) graph.Graph[int, GVertex] {
	log.Info().Msg("Creating new graph.")
	vertexHash := func(vertex GVertex) int {
		return vertex.ID
	}
	g := graph.New(vertexHash, graph.Directed())

	info, conn, err := GetRedisInfo(redisURL)
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
		_ = g.AddVertex(vertex)
		//if err != nil {
		//	log.Debug().Err(err).Msg("Vertex already exists.")
		//	continue
		//}
	}

	for _, edge := range info.Edges {
		hMap := utils.NewMap[string, *Vehicle]()
		_ = g.AddEdge(
			edge.FromVertexID,
			edge.ToVertexID,
			graph.EdgeData(EdgeData{
				MaxSpeed: edge.MaxSpeed,
				Length:   edge.Length,
				Map:      &hMap,
			}))
		//if err != nil {
		//	log.Debug().Err(err).Msg("Edge already exists.")
		//	continue
		//}
	}

	return g
}

// GetVertices returns a list of vertices in the graph
func GetVertices(g *graph.Graph[int, GVertex]) ([]GVertex, error) {
	gg := *g
	edges, err := gg.Edges()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get edges.")
		return nil, err
	}

	vertices := make(map[int]bool, 0)

	for _, edge := range edges {
		src := edge.Source
		dst := edge.Target

		vertices[src] = false
		vertices[dst] = false
	}

	keys := make([]int, 0, len(vertices))
	for k := range vertices {
		keys = append(keys, k)
	}

	gVertices := make([]GVertex, 0, len(vertices))
	for _, key := range keys {
		vertex, err := gg.Vertex(key)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get vertex.")
			return nil, err
		}
		gVertices = append(gVertices, vertex)
	}

	return gVertices, nil
}

// GetTopRightBottomLeftVertices returns the top right and bottom left vertices of the graph
func GetTopRightBottomLeftVertices(gr *graph.Graph[int, GVertex]) (bot, top Point) {
	// Get all vertices
	vertices, err := GetVertices(gr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get vertices.")
		return bot, top
	}

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

	bot = Point{
		X: botX,
		Y: botY,
	}
	top = Point{
		X: topX,
		Y: topY,
	}

	log.Debug().Msgf("Bottom left vertex: %v", bot)
	log.Debug().Msgf("Top right vertex: %v", top)

	return bot, top
}

// Point is a point in 2D space
type Point struct {
	X, Y float64
}

type FloatPoint struct {
	X, Y *big.Float
}

// Rect is a rectangle in 2D space, holding the top right and bottom left points
// and the vertices of the rectangle
type Rect struct {
	TopRight Point
	BotLeft  Point
	Vertices []GVertex
}

// BRect is a __big__ rectangle in 2D space, holding the top right and bottom left points
// and the vertices of the rectangle
type BRect struct {
	TopRight FloatPoint
	BotLeft  FloatPoint
	Vertices []GVertex
}

func (b *BRect) toRect() Rect {
	tX, _ := b.TopRight.X.Float64()
	tY, _ := b.TopRight.Y.Float64()
	bX, _ := b.BotLeft.X.Float64()
	bY, _ := b.BotLeft.Y.Float64()
	return Rect{
		TopRight: Point{
			X: tX,
			Y: tY,
		},
		BotLeft: Point{
			X: bX,
			Y: bY,
		},
		Vertices: b.Vertices,
	}
}

// DivideGraph divides the graph into n parts. Column-wise division.
func DivideGraph(n int, gr *graph.Graph[int, GVertex]) ([]Rect, error) {
	rootBot, rootTop := GetTopRightBottomLeftVertices(gr)
	// Get all vertices
	vertices, err := GetVertices(gr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get vertices.")
		return nil, err
	}

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
			Vertices: make([]GVertex, 0),
		}

		for _, vertex := range vertices {
			isInYInterval := vertex.Y >= rootBot.Y && vertex.Y <= rootTop.Y
			isInXInterval := vertex.X >= botX && vertex.X <= topX

			if isInYInterval && isInXInterval {
				rects[i].Vertices = append(rects[i].Vertices, vertex)
			}
		}
	}

	return rects, nil
}
