package streets

// Path is a struct for a path in the graph
type Path struct {
	StartVertex *Vertex
	EndVertex   *Vertex
	Vertices    []Vertex
}
