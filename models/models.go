package models

// Element represents an element in the HNSW graph.
type Element struct {
	ID         int
	Embeddings []float64
	Msg        string
}

// Candidate represents a node and its distance to the query point.
type Candidate struct {
	NodeID   int
	Distance float64
}
