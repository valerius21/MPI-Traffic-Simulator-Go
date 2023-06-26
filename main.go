package main

import (
	"os"

	"github.com/dominikbraun/graph"
	"github.com/dominikbraun/graph/draw"
	"github.com/rs/zerolog/log"

	"pchpc/streets"
)

func createDot(g graph.Graph[int, streets.GVertex]) {
	file, _ := os.Create("./mygraph.gv")
	_ = draw.DOT(g, file)
}

func main() {
	g := streets.NewGraph()
	size, err := g.Size()
	if err != nil {
		panic(err)
	}

	log.Info().Msgf("Graph size: %d", size)

	_, err = g.Edges()
	if err != nil {
		panic(err)
	}

	path, err := graph.ShortestPath(g, 2617388513, 1247500404)
	if err != nil {
		panic(err)
	}

	vh1 := streets.NewVehicle(2.5, path, g)

	log.Info().Msgf("Vehicle: %s", vh1)
}
