package streets

// Edge is a struct for an edge in the graph
type Edge struct {
	ID int

	FromVertexID int
	ToVertexID   int

	Length   float64
	MaxSpeed float64

	Graph *Graph
}
