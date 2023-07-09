package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"math/rand"
	"os"
	"sync"

	"github.com/dominikbraun/graph/draw"
	mpi "github.com/sbromberger/gompi"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

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

			var v streets.Vehicle

			if mpi.IsOn() {
				v = streets.Vehicle{}
				// TODO: continue here
			} else {
				vh, err := setVehicle(g, speed)
				if err != nil {
					log.Error().Err(err).Msg("Failed to set vehicle.")
					return
				}
				v = vh
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
	useMPI := flag.Bool("mpi", false, "Use MPI")

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

	if *useMPI {
		mpi.Start(true)
		defer mpi.Stop()
		if !mpi.IsOn() {
			log.Error().Msg("MPI is not on.")
			return
		}

		comm := mpi.NewCommunicator(nil)

		numTasks := comm.Size()
		taskID := comm.Rank()
		rectanglesTag := 1
		edgesTag := 2
		pathsTag := 3

		if taskID == 0 {
			// "chunkify"

			g := streets.NewGraph(*redisURL)

			rects, err := streets.DivideGraphsIntoRects(numTasks, &g)
			if err != nil {
				log.Error().Err(err).Msg("Failed to divide graph.")
				return
			}
			log.Debug().Msgf("MPI: Number of tasks: %d My rank: %d", numTasks, taskID)
			log.Debug().Msgf("MPI: Number of rects: %d", len(rects))

			// parse edges
			edges, err := g.Edges()
			rawEdges := make([]streets.RawEdge[int], len(edges))

			for _, e := range edges {
				rawEdges = append(rawEdges, streets.RawEdge[int]{
					Source: e.Source,
					Target: e.Target,
				})
			}

			if err != nil {
				log.Error().Err(err).Msg("Failed to get edges.")
				return
			}

			// send rects to other tasks
			for i := 1; i < numTasks; i++ {
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				err := enc.Encode(rects)
				if err != nil {
					log.Error().Err(err).Msg("Failed to encode rects.")
					return
				}
				comm.SendBytes(buf.Bytes(), i, rectanglesTag)

				buf.Reset()
				err = enc.Encode(rawEdges)
				if err != nil {
					log.Error().Err(err).Msg("Failed to encode edges.")
					return
				}
				comm.SendBytes(buf.Bytes(), i, edgesTag)
			}

			// create vehicle routes

			// total vehicles = n * numTasks
			// get random nodes
			verts, err := streets.GetVertices(&g)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get vertices.")
				return
			}

			paths := make([][]int, 0)
			for i := 0; i < numTasks*(*n); i++ {
				var path []int
				for len(path) < 2 {
					rand.Shuffle(len(verts), func(i, j int) {
						verts[i], verts[j] = verts[j], verts[i]
					})
					start := verts[0]
					end := verts[1]
					p, err := graph.ShortestPath(g, start.ID, end.ID)
					if err != nil {
						log.Debug().Err(err).Msg("Failed to get shortest path.")
						continue
					}
					path = p
				}
				if len(path) < 2 {
					panic("Path is too short")
				}
				paths = append(paths, path)
			}

			// send paths to other tasks
			for i := 1; i < numTasks; i++ {
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				err := enc.Encode(paths)
				if err != nil {
					log.Error().Err(err).Msg("Failed to encode paths.")
					return
				}
				comm.SendBytes(buf.Bytes(), i, pathsTag)
			}

		} else {
			myId := comm.Rank()
			// receive rects from task 0
			var buf bytes.Buffer
			dec := gob.NewDecoder(&buf)

			bbs, _ := comm.RecvBytes(0, rectanglesTag)
			buf.Write(bbs)

			var rects []streets.Rect
			err := dec.Decode(&rects)
			if err != nil {
				log.Error().Err(err).Msg("Failed to decode rects.")
				return
			}

			log.Debug().Msgf("MPI: Number of tasks: %d My rank: %d", numTasks, taskID)
			log.Debug().Msgf("MPI: Number of rects: %d", len(rects))

			buf.Reset()
			bbs, _ = comm.RecvBytes(0, edgesTag)
			buf.Write(bbs)

			var rawEdges []streets.RawEdge[int]
			err = dec.Decode(&rawEdges)

			myRect := rects[myId]

			// init subgraph
			g := streets.GraphFromRect(rawEdges, myRect)
			size, err := g.Size()
			if err != nil {
				log.Error().Err(err).Msg("Failed to get graph size.")
				return
			}
			log.Info().Msgf("Process %d: Graph size: %d", myId, size)

			// receive paths from task 0
			buf.Reset()
			bbs, _ = comm.RecvBytes(0, pathsTag)
			buf.Write(bbs)

			var paths [][]int
			err = dec.Decode(&paths)
			if err != nil {
				log.Error().Err(err).Msg("Failed to decode paths.")
				return
			}

			// TODO: check indices
			paths = paths[(myId-1)*(*n) : (myId+1)*(*n)]

			// TODO: check paths
			log.Info().Msgf("Process %d: Number of paths (%d-%d): %d", myId, len(paths), (myId-1)*(*n), (myId+1)*(*n))

			run(&g, n, minSpeed, maxSpeed, useRoutines)
		}

	} else {
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
}
