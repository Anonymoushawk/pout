package server

import (
	"math/rand"
	"time"
)

// `TimeGraph` represents the plotting variables used in the connections graph.
type TimeGraph struct {
	DataMin  float64
	DataMax  float64
	DataX    []float64
	DataY    []float64
	ScatterY []float64
}

// `TimeGraph.UpdateConnectionGraph` appends the required data to the connection graph.
func (graph *TimeGraph) UpdateConnectionGraph(c *Client) {
	// Update the server connection graph to include this connection.
	graph.DataX = append(graph.DataX, float64(time.Now().Unix()))
	graph.DataY = append(graph.DataY, rand.Float64())
	graph.ScatterY = append(graph.ScatterY, rand.Float64())

	graph.DataMin = graph.DataX[0]
	graph.DataMax = graph.DataX[len(graph.DataX)-1]
}
