package streets

// Vertex is a struct for a vertex in the graph
type Vertex struct {
	ID int
	// X  float32
	// Y  float32

	Edges []Edge
	Graph *Graph
}
