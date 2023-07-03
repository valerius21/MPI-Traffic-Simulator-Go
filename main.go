package main

import (
	"flag"
	"math/rand"
	"os"
	"sync"

	"pchpc/utils"

	"github.com/rs/zerolog"

	"github.com/dominikbraun/graph"
	"github.com/rs/zerolog/log"

	"pchpc/streets"
)

// getVertices returns a list of vertices in the graph
func getVertices(g *graph.Graph[int, streets.GVertex]) ([]int, error) {
	edges, err := (*g).Edges()
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

	return keys, nil
}

// setVehicle creates a vehicle with a random path
func setVehicle(g *graph.Graph[int, streets.GVertex], speed float64) (streets.Vehicle, error) {
	vertices, err := getVertices(g)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get vertices.")
		return *new(streets.Vehicle), err
	}
	var path []int
	for len(path) < 2 {
		srcIdx := rand.Intn(len(vertices))
		src := vertices[srcIdx]
		destIdx := rand.Intn(len(vertices))
		dest := vertices[destIdx]
		path, _ = graph.ShortestPath(*g, src, dest)
	}
	v := streets.NewVehicle(speed, path, g)
	return v, nil
}

// main is the entry point of the program
func main() {
	// Flags
	n := flag.Int("n", 100, "Number of vehicles")
	useRoutines := flag.Bool("m", false, "Use goroutines")
	minSpeed := flag.Float64("min-speed", 5.5, "Minimum speed")
	maxSpeed := flag.Float64("max-speed", 8.5, "Maximum speed")

	flag.Parse()

	// Logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	runLogFile, _ := os.OpenFile(
		"main.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o664,
	)
	multi := zerolog.MultiLevelWriter(os.Stdout, runLogFile)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	// Init Graph
	g := streets.NewGraph()
	size, err := g.Size()
	if err != nil {
		panic(err)
	}

	log.Info().Msgf("Graph size: %d", size)

	// Create vehicles and drive
	var wg sync.WaitGroup
	j := 0

	for i := 0; i < *n; i++ {
		wg.Add(1)
		j++
		fn := func() {
			defer func() {
				j--
				wg.Done()
			}()
			speed := utils.RandomFloat64(*minSpeed, *maxSpeed)
			v, err := setVehicle(&g, speed)
			if err != nil {
				log.Error().Err(err).Msg("Failed to set vehicle.")
				return
			}
			log.Debug().Msgf("Vehicle: %s", v)
			for !v.IsParked {
				log.Info().Msgf("Active Vehicles: %d", j)
				v.Step()
				log.Debug().Msgf("Vehicle: %s", v)
				v.PrintInfo()
			}
			log.Debug().Msgf("Vehicle Parked %s", v.ID)
		}
		if *useRoutines {
			go fn()
		} else {
			fn()
		}
	}

	wg.Wait()
}
