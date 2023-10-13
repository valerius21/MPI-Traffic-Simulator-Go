package streets

import (
	"errors"
	"fmt"

	"github.com/dominikbraun/graph"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

// StreetGraph is a graph of streets with vertices of type int and edges of type JVertex
type StreetGraph struct {
	// ID is the ID of the graph. Root graph has ID 0
	ID string

	// RootGraph is the root graph of the graph, nil if the graph is the root graph
	RootGraph *StreetGraph

	// graph is the graph
	Graph graph.Graph[int, JVertex]
}

// convertEdgeToJEdge converts an edge to a JEdge
func convertEdgeToJEdge(edge *graph.Edge[int]) (JEdge, error) {
	edgeData, err := GetEdgeData(*edge)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get edge data.")
		return JEdge{}, err
	}
	return JEdge{
		From:     edge.Source,
		To:       edge.Target,
		Length:   edgeData.Length,
		MaxSpeed: fmt.Sprintf("%.2f", edgeData.MaxSpeed),
		Name:     edgeData.Name,
		ID:       edgeData.ID,
		Data:     edgeData,
	}, nil
}

// produceRootGraph produces a root graph
func produceRootGraph(filePath string) *StreetGraph {
	gb := NewGraphBuilder().FromJsonFile(filePath).WithRectangleParts(1)
	gb = gb.SetTopRightBottomLeftVertices().DivideGraphsIntoRects()
	gb = gb.PickRect(0).FilterForRect().IsRoot()
	g, err := gb.Build()
	if err != nil {
		log.Error().Msgf("Error creating root graph: %v", err)
		panic(err)
		return nil
	}
	return g
}

// produceLeafGraph produces a leaf graph
func produceLeafGraph(index, parts int, rootGraph *StreetGraph) *StreetGraph {
	vertices, err := rootGraph.GetVertices()
	if err != nil {
		log.Error().Msgf("Error getting vertices from root graph: %v", err)
		panic(err)
		return nil
	}
	gEdges, err := rootGraph.Graph.Edges()
	if err != nil {
		log.Error().Msgf("Error getting edges from root graph: %v", err)
		panic(err)
		return nil
	}

	edges := make([]JEdge, len(gEdges))
	for i, edge := range gEdges {
		edges[i], err = convertEdgeToJEdge(&edge)
		if err != nil {
			log.Error().Msgf("Error converting edge to JEdge: %v", err)
			panic(err)
			return nil
		}
	}

	gb := NewGraphBuilder().WithVertices(vertices).WithEdges(edges).WithRectangleParts(parts)
	gb = gb.PickRect(index).FilterForRect().IsLeaf()

	g, err := gb.Build()
	if err != nil {
		log.Error().Msgf("Error creating leaf graph: %v", err)
		panic(err)
	}

	return g
}

// DefaultGraph creates a default graph
func DefaultGraph(filePath string, nRects int) (root *StreetGraph, leafs []*StreetGraph) {
	root = produceRootGraph(filePath)
	if nRects < 2 {
		return root, nil
	}

	leafs = make([]*StreetGraph, nRects)
	for i := 0; i < nRects; i++ {
		l := produceLeafGraph(i, root)
		leafs[i] = l
	}
	return root, leafs
}

// VertexInGraph checks if a vertex is in a graph
func (g *StreetGraph) VertexInGraph(v JVertex) bool {
	_, err := (*g).Graph.Vertex(v.ID)
	return err == nil
}

// GetVertices gets all vertices in a graph
func (g *StreetGraph) GetVertices() ([]JVertex, error) {
	gr := (*g).Graph
	edges, err := gr.Edges()
	if err != nil {
		log.Error().Msgf("Error getting edges from graph: %v", err)
		return nil, err
	}

	vertices := make([]JVertex, 0)
	for _, edge := range edges {
		dstID := edge.Target
		srcID := edge.Source

		dst, err := gr.Vertex(dstID)
		if err != nil {
			log.Error().Msgf("Error getting vertex from graph: %v", err)
			return nil, err
		}

		src, err := gr.Vertex(srcID)

		if err != nil {
			log.Error().Msgf("Error getting vertex from graph: %v", err)
			return nil, err
		}

		if !slices.Contains(vertices, dst) {
			vertices = append(vertices, dst)
		}
		if !slices.Contains(vertices, src) {
			vertices = append(vertices, dst, src)
		}
	}

	return vertices, nil
}

// GetEdgeData returns the data of an edge
func GetEdgeData(edge graph.Edge[int]) (Data, error) {
	if data, ok := edge.Properties.Data.(Data); ok {
		return data, nil
	}
	return Data{}, errors.New("edge data is not of type Data")
}
