package main

import (
	"pchpc/streets"
)

func main() {
	//v := vehicles.New(255, 255)
	// format the vehicle output
	//fmt.Printf("%+v", v)

	graph, conn, err := streets.New()
	defer conn.Close()

	if err != nil {
		panic(err)
	}

	a := streets.Vertex{
		ID:    51,
		X:     0,
		Y:     0,
		Edges: nil,
	}

	b := streets.Vertex{
		ID:    91,
		X:     0,
		Y:     0,
		Edges: nil,
	}

	graph.FindPath(&a, &b)
}
