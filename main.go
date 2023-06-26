package main

import (
	"os"

	"github.com/gomodule/redigo/redis"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"pchpc/streets"
)

func makeStep(v1 *streets.Vehicle) {
	if v1.IsParked {
		log.Info().Msgf("Vehicle %s is parked (%d seconds)", v1.ID)
		return
	}
	v1.Step()
	v1.PrintInfo()
}

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
	vehicles := make([]*streets.Vehicle, 0)

	for i := 0; i < 3; i++ {
		v1 := streets.NewVehicle(path, 2.5, graph)
		log.Info().Msgf("Vehicle %d (%v) started", i, v1.ID)
		vehicles = append(vehicles, &v1)
	}

	for !vehicles[0].IsParked {
		for _, v1 := range vehicles {
			makeStep(v1)
		}
	}
}
