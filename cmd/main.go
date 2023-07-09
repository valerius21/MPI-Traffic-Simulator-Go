package main

import (
	"flag"
	"math/rand"
	"os"
	"sync"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/dominikbraun/graph/draw"

	"github.com/rs/zerolog"

	"pchpc/utils"

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

// run creates vehicles and drives them
func run(g *graph.Graph[int, streets.GVertex], n *int, minSpeed *float64, maxSpeed *float64, useRoutines *bool) {
	// Create vehicles and drive
	var wg sync.WaitGroup
	p := mpb.New(mpb.WithWaitGroup(&wg))
	j := 0

	total := *n
	bar := p.AddBar(int64(total),
		mpb.PrependDecorators(decor.Name("Vehicles arrived: "),
			decor.Percentage(decor.WCSyncSpace)),
		mpb.AppendDecorators(
			// replace ETA decorator with "done" message, OnComplete event
			decor.OnComplete(
				// ETA decorator with ewma age of 30
				decor.EwmaETA(decor.ET_STYLE_GO, 30, decor.WCSyncWidth), "done",
			),
		),
	)

	for i := 0; i < total; i++ {
		wg.Add(1)
		j++

		fn := func() {
			defer func() {
				j--
				wg.Done()
				bar.Increment()
			}()
			speed := utils.RandomFloat64(*minSpeed, *maxSpeed)
			v, err := setVehicle(g, speed)
			if err != nil {
				log.Error().Err(err).Msg("Failed to set vehicle.")
				return
			}
			log.Debug().Msgf("Vehicle: %s", v.String())
			for !v.IsParked {
				log.Debug().Msgf("Active Vehicles: %d of %d", j, *n)
				v.Step()
				log.Debug().Msgf("Vehicle: %s", v.String())
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

	// wg.Wait()
	p.Wait()
}

// saveGraph saves the graph to a file in the current working directory
func saveGraph(g *graph.Graph[int, streets.GVertex]) error {
	file, err := os.Create("graph.gv")
	if err != nil {
		return err
	}
	return draw.DOT(*g, file)
}

// main is the entry point of the program
func main() {
	// Flags
	n := flag.Int("n", 100, "Number of vehicles")
	useRoutines := flag.Bool("m", false, "Use goroutines")
	minSpeed := flag.Float64("min-speed", 5.5, "Minimum speed")
	maxSpeed := flag.Float64("max-speed", 8.5, "Maximum speed")
	dbPath := flag.String("dbFile", "assets/db.sqlite", "Path to the database")
	redisURL := flag.String("redisURL", "localhost:6379", "URL to the redis server")
	exportGraph := flag.Bool("export", false, "Export graph to graph.gv (current working directory)")
	debug := flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()

	// Init DB
	utils.SetDBPath(*dbPath)

	// Logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	runLogFile, _ := os.OpenFile(
		"main.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0o664,
	)
	multi := zerolog.MultiLevelWriter(os.Stdout, runLogFile)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()

	// Init Graph
	g := streets.NewGraph(*redisURL)
	ed, err := g.Edges()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get edges.")
		return
	}

	log.Debug().Msgf("Edges: %d", len(ed))

	// save graph async
	if *exportGraph {
		err := saveGraph(&g)
		if err != nil {
			log.Error().Err(err).Msg("Failed to save graph.")
		}
	}

	run(&g, n, minSpeed, maxSpeed, useRoutines)
}
