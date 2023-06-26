package main

import (
	"os"
	"sync"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog"

	"github.com/rs/zerolog/log"

	"pchpc/streets"
)

func allVehiclesParked(vehicles []*streets.Vehicle) bool {
	for _, v := range vehicles {
		if !v.IsParked {
			return false
		}
	}
	return true
}

func Multi(graph streets.Graph) {
	a := streets.Vertex{
		ID: 28127535,
	}

	b := streets.Vertex{
		ID: 208640196,
	}

	// a := streets.Vertex{ID: 60347877}
	// b := streets.Vertex{ID: 73066996}

	path, err := graph.FindPath(&a, &b)
	if err != nil {
		panic(err)
	}

	log.Info().Msgf("Path N=%v", len(path.Vertices))
	vehicles := make([]*streets.Vehicle, 0)

	for i := 0; i < 100; i++ {
		v1 := streets.NewVehicle(path, 2.5, graph)
		log.Info().Msgf("Vehicle %d (%v) started", i, v1.ID)
		vehicles = append(vehicles, &v1)
	}

	var wg sync.WaitGroup

	for !allVehiclesParked(vehicles) {
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

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	// zerolog.SetGlobalLevel(zerolog.PanicLevel)

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
	Multi(graph)
}
