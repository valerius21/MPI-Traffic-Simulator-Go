package streets_test

import (
	"sync"
	"testing"

	"github.com/rs/zerolog"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"

	"github.com/gammazero/deque"
	"github.com/stretchr/testify/assert"
	"pchpc/streets"
)

func TestVehicle_GetCurrentEdge(t *testing.T) {
	vertex1 := streets.Vertex{1, nil, nil}
	vertex2 := streets.Vertex{2, nil, nil}
	edge := streets.Edge{
		ID:           0,
		FromVertexID: 1,
		ToVertexID:   2,
		Length:       100,
		MaxSpeed:     0,
		Q:            &deque.Deque[*streets.Vehicle]{},
		Graph:        nil,
	}
	g := streets.Graph{
		Vertices: []streets.Vertex{vertex1, vertex2},
		Edges:    []streets.Edge{edge},
		Rdb:      nil,
	}
	vertex1.Graph = &g
	vertex2.Graph = &g
	edge.Graph = &g
	q := deque.Deque[float64]{}
	q.PushBack(0.0)
	s := streets.Vehicle{
		ID:         "vh1",
		Speed:      10,
		PathLength: &q,
		Path: streets.Path{
			StartVertex: &vertex1,
			EndVertex:   &vertex2,
			Vertices:    []streets.Vertex{vertex1, vertex2},
		},
		Graph: &g,
	}

	vh1 := s
	edge.PushVehicle(&vh1)
	testID := vh1.GetCurrentEdge().ID

	assert.Equal(t, edge.ID, testID, "Expected vh1 to be on edge")
}

func BenchmarkMakeStep(b *testing.B) {
	zerolog.SetGlobalLevel(zerolog.PanicLevel)
	graph, conn, err := streets.New()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			log.Panic().Err(err).Msg("Failed to close Redis connection")
		}
	}(conn)

	aVertex := streets.Vertex{
		ID: 28127535,
	}

	bVertex := streets.Vertex{
		ID: 208640196,
	}

	// a := streets.Vertex{ID: 60347877}
	// b := streets.Vertex{ID: 73066996}

	path, err := graph.FindPath(&aVertex, &bVertex)
	if err != nil {
		panic(err)
	}

	log.Info().Msgf("Path N=%v", len(path.Vertices))
	vehicles := make([]*streets.Vehicle, 0)

	for i := 0; i < b.N; i++ {
		// fmt.Println(i)
		v1 := streets.NewVehicle(path, 2.5, graph)
		log.Info().Msgf("Vehicle %d (%v) started", i, v1.ID)
		vehicles = append(vehicles, &v1)
	}

	var wg sync.WaitGroup

	for !vehicles[0].IsParked {
		for _, v1 := range vehicles {
			wg.Add(1)
			v1 := v1
			go func() {
				streets.MakeStep(v1)
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
