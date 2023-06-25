package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"pchpc/streets"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	graph, conn, err := streets.New()
	defer conn.Close()

	if err != nil {
		panic(err)
	}

	a := streets.Vertex{
		ID:    213322468,
		X:     0,
		Y:     0,
		Edges: nil,
	}

	b := streets.Vertex{
		ID:    270678741,
		X:     0,
		Y:     0,
		Edges: nil,
	}

	path, err := graph.FindPath(&a, &b)
	log.Info().Msgf("Path: %v", path)
}
