package main

import (
	"os"

	"github.com/gammazero/deque"

	"pchpc/vehicles"

	"github.com/gomodule/redigo/redis"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"pchpc/streets"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	graph, conn, err := streets.New()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			log.Panic().Err(err).Msg("Failed to close Redis connection")
		}
	}(conn)

	if err != nil {
		panic(err)
	}

	//a := streets.Vertex{
	//	ID: 28127535,
	//}
	//
	//b := streets.Vertex{
	//	ID: 208640196,
	//}

	a := streets.Vertex{ID: 60347877}
	b := streets.Vertex{ID: 73066996}

	path, err := graph.FindPath(&a, &b)
	log.Info().Msgf("Path N=%v", len(path.Vertices))

	v1 := vehicles.New(path, 2.5, graph)

	for i := 0; i < 30; i++ {
		if v1.IsParked {
			log.Info().Msgf("Vehicle %s is parked (%d seconds)", v1.ID, i)
			break
		}
		v1.Step()
		v1.PrintInfo()
	}

	var q deque.Deque[vehicles.Vehicle]

	for i := 0; i < 5; i++ {
		v := vehicles.New(path, 2.5, graph)
		q.PushBack(v)
	}

	for i := 0; i < q.Len(); i++ {
		vv := q.At(i)
		log.Info().Msgf("Vehicle %s (%d)", vv.ID, i)
	}
}
