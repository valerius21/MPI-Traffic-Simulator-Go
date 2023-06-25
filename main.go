package main

import (
	"os"

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

	a := streets.Vertex{
		ID: 28127535,
	}

	b := streets.Vertex{
		ID: 208640196,
	}

	path, err := graph.FindPath(&a, &b)
	log.Info().Msgf("Path N=%v", len(path.Vertices))
}
