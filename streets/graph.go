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
func GetTopRightBottomLeftVertices(gr *graph.Graph[int, GVertex]) (bot, top GVertex) {
	bot = GVertex{
		ID: -1,
		X:  999.,
		Y:  999.,
	}
	top = GVertex{
		ID: -1,
		X:  -999.,
		Y:  -999.,
	}
	// Get all vertices
	vertices, err := GetVertices(gr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get vertices.")
		return bot, top
	}

	for _, vertex := range vertices {

		if vertex.X < bot.X && vertex.Y < bot.Y {
			bot = vertex
		}
		if vertex.X > top.X && vertex.Y > top.Y {
			top = vertex
		}
	}

	log.Debug().Msgf("Bottom left vertex: %v", bot)
	log.Debug().Msgf("Top right vertex: %v", top)

	return bot, top
}

// Point is a point in 2D space
type Point struct {
	X, Y float64
}

type RatPoint struct {
	X, Y *big.Rat
}

// Rect is a rectangle in 2D space, holding the top right and bottom left points
// and the vertices of the rectangle
type Rect struct {
	TopRight Point
	BotLeft  Point
	Vertices []GVertex
}

// DivideGraph divides the graph into n parts. Column-wise division.
func DivideGraph(n int, gr *graph.Graph[int, GVertex]) ([]Rect, error) {
	fbot, ftop := GetTopRightBottomLeftVertices(gr)

	bot := RatPoint{
		X: new(big.Rat).SetFloat64(fbot.X),
		Y: new(big.Rat).SetFloat64(fbot.Y),
	}

	top := RatPoint{
		X: new(big.Rat).SetFloat64(ftop.X),
		Y: new(big.Rat).SetFloat64(ftop.Y),
	}

	// Get all vertices
	vertices, err := GetVertices(gr)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get vertices.")
		return nil, err
	}
	rects := make([]Rect, n)

	width := new(big.Rat).Sub(top.X, bot.X)
	deltaX := new(big.Rat).Quo(width, big.NewInt(int64(n)))

	for i := 0; i < n; i++ {
		multiplicativeFactor := new(big.Rat).SetInt64(int64(i))
		topRightX := new(big.Rat).Add(bot.X, deltaX)
		botLeftX := new(big.Rat).Add(bot.X, new(big.Rat).Mul(big.NewInt(int64(i)), deltaX))

		rects[i] = Rect{
			TopRight: Point{
				X: topRightX,
				Y: top.Y,
			},
			BotLeft: Point{
				X: botLeftX,
				Y: bot.Y,
			},
			Vertices: make([]GVertex, 0, len(vertices)),
		}

		for _, vertex := range vertices {
			if vertex.X.Cmp(rects[i].TopRight.X) <= 0 && vertex.X.Cmp(rects[i].BotLeft.X) >= 0 {
				rects[i].Vertices = append(rects[i].Vertices, vertex)
			}
		}
	}

	return rects, nil
}
