package hnsw

import (
	"container/heap"
	"math"
	"math/rand"
	"sync"

	hnswheap "github.com/lblclass/hnswgo/util/heap"

	"github.com/lblclass/hnswgo/models"
)

// HNSW is the main struct representing the graph.
type HNSW struct {
	Layers          []map[int]*hnswheap.CandidateHeap // Connections at each layer
	EnterPoint      int                               // Entry point ID
	M               int
	maxConnections  int     // Maximum connections per element
	EfConstruction  int     // Candidate list size
	NormalizationML float64 // Level normalization factor
	MaxLayers       int
	Elements        map[int]models.Element // Element data
	mu              sync.RWMutex
}

// NewHNSW initializes an HNSW graph.
func NewHNSW(efConstruction, M, maxLayers int, nm float64) *HNSW {
	return &HNSW{
		Layers:          []map[int]*hnswheap.CandidateHeap{},
		EnterPoint:      -1,
		M:               M,
		maxConnections:  2 * M,
		EfConstruction:  efConstruction,
		MaxLayers:       maxLayers,
		NormalizationML: nm,
		Elements:        make(map[int]models.Element),
		mu:              sync.RWMutex{},
	}
}

// Insert adds a new element into the HNSW graph.
func (h *HNSW) Insert(q models.Element) {
	level := h.generateLevel()
	if level > h.MaxLayers {
		level = h.MaxLayers
	}
	h.Elements[q.ID] = q
	topLevel := len(h.Layers) - 1
	ep := h.EnterPoint
	if topLevel <= level {
		// Add new layers if needed.
		for i := len(h.Layers); i <= level; i++ {
			tmp := hnswheap.CandidateHeap{
				Compare: "big",
			}
			heap.Init(&tmp)
			h.Layers = append(h.Layers, map[int]*hnswheap.CandidateHeap{
				q.ID: &tmp,
			})
		}
		h.EnterPoint = q.ID
	} else {
		// 如果topLevel大于level，则ep需要有些变化。
		for lc := topLevel; lc > level; lc-- {
			tmpEp := h.SearchLayer(q, ep, 1, lc)
			ep = tmpEp.Candidates[0].NodeID
		}
	}

	for lc := min(topLevel, level); lc >= 0; lc-- {
		tmp := hnswheap.CandidateHeap{
			Compare: "big",
		}
		heap.Init(&tmp)
		h.Layers[lc][q.ID] = &tmp
		tmpNeighbors := h.SearchLayer(q, ep, h.EfConstruction, lc)
		neighbors := tmpNeighbors.ExtractHeapData()
		selected := h.SelectNeighborsHeuristic(q, neighbors, h.M, lc, true, true)

		// Connect bidirectionally.
		for _, n := range selected {
			h.addConnection(n, q.ID, lc)
			h.addConnection(q.ID, n, lc)
		}
		ep = neighbors[0]
	}

}

// searchLayer finds nearest neighbors in the specified layer.
func (h *HNSW) SearchLayer(q models.Element, entryPoint int, ef int, lc int) *hnswheap.CandidateHeap {
	V := map[int]bool{entryPoint: true} // set of visited elements
	qCandidate := models.Candidate{
		NodeID:   entryPoint,
		Distance: h.Distance(q, h.Elements[entryPoint]),
	}
	C := hnswheap.NewSmallCandidatesHeap()
	heap.Init(C)
	heap.Push(C, qCandidate)
	W := hnswheap.NewBigCandidatesHeap()
	heap.Init(W)
	heap.Push(W, qCandidate)
	for C.Len() > 0 {
		nc := heap.Pop(C).(models.Candidate)
		fc := W.Candidates[0]
		if nc.Distance > fc.Distance {
			break
		}
		for i := 0; i < h.Layers[lc][nc.NodeID].Len(); i++ {
			vNode := h.Layers[lc][nc.NodeID].Candidates[i].NodeID
			if _, ok := V[vNode]; ok {
				continue
			}
			V[vNode] = true
			if (fc.Distance > h.Distance(q, h.Elements[vNode])) || (W.Len() < ef) {
				tmpC := models.Candidate{NodeID: vNode, Distance: h.Distance(q, h.Elements[vNode])}
				heap.Push(C, tmpC)
				heap.Push(W, tmpC)
				if W.Len() > ef {
					heap.Pop(W)
				}
			}
		}
	}
	// Simplified: replace with real search logic.
	return W

}

// selectNeighbors selects M nearest neighbors from the candidates.
func (h *HNSW) SelectNeighborsSimple(q models.Element, candidates []int) []int {
	// candidates is already ordered at select
	return candidates[:min(len(candidates), h.M)]
}

// addConnection adds a connection to the graph.
func (h *HNSW) addConnection(from, to, layer int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ft := h.Distance(h.Elements[from], h.Elements[to])
	toCandidate := models.Candidate{
		NodeID:   to,
		Distance: ft,
	}
	if h.Layers[layer][from].Len() < h.maxConnections {
		heap.Push(h.Layers[layer][from], toCandidate)
	} else {
		if ft < h.Layers[layer][from].Candidates[0].Distance {
			heap.Pop(h.Layers[layer][from])
			heap.Push(h.Layers[layer][from], toCandidate)
		}
	}
}

// generateLevel determines the level for a new element.
func (h *HNSW) generateLevel() int {
	return int(math.Floor(-math.Log(rand.Float64()) * h.NormalizationML))
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (h *HNSW) Distance(e1, e2 models.Element) float64 {
	sum := 0.0
	for i := range e1.Embeddings {
		diff := e1.Embeddings[i] - e2.Embeddings[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// SelectNeighborsHeuristic implements Algorithm 4.
func (h *HNSW) SelectNeighborsHeuristic(
	q models.Element,
	candidates []int,
	M int,
	layer int,
	extendCandidates bool,
	keepPrunedConnections bool,
) []int {
	R := make(map[int]bool) // Result set
	W := hnswheap.NewSmallCandidatesHeap()
	heap.Init(W)

	// Add initial candidates to the queue.
	for _, c := range candidates {
		dist := h.Distance(q, h.Elements[c])
		heap.Push(W, models.Candidate{NodeID: c, Distance: dist})
	}

	// Extend candidates by their neighbors if needed.
	if extendCandidates {
		neighbors := map[int]bool{}
		for _, c := range candidates {
			for _, neighborStruct := range h.Layers[layer][c].Candidates {
				neighbor := neighborStruct.NodeID
				if _, exists := neighbors[neighbor]; !exists {
					dist := h.Distance(q, h.Elements[neighbor])
					heap.Push(W, models.Candidate{NodeID: neighbor, Distance: dist})
					neighbors[neighbor] = true
				}
			}
		}
	}

	Wd := hnswheap.NewSmallCandidatesHeap() // Discarded candidates
	heap.Init(Wd)

	// Process candidates.
	for W.Len() > 0 && len(R) < M {
		e := heap.Pop(W).(models.Candidate)
		closer := true
		for r := range R {
			if h.Distance(h.Elements[e.NodeID], h.Elements[r]) < e.Distance {
				closer = false
				break
			}
		}
		if closer {
			R[e.NodeID] = true
		} else {
			heap.Push(Wd, e)
		}
	}

	// Add pruned connections if required.
	if keepPrunedConnections {
		for Wd.Len() > 0 && len(R) < M {
			e := heap.Pop(Wd).(models.Candidate)
			R[e.NodeID] = true
		}
	}

	// Convert result set to slice.
	result := make([]int, 0, len(R))
	for r := range R {
		result = append(result, r)
	}
	return result
}

func (h *HNSW) KNNSearch(q models.Element, K int) []int {
	// W := make(hnswheap.SmallCandidates, 0)
	ep := h.EnterPoint
	totalLayers := len(h.Layers)
	for lc := (totalLayers - 1); lc >= 1; lc-- {
		W := h.SearchLayer(q, ep, 1, lc)
		ep = W.Candidates[0].NodeID
	}
	W := h.SearchLayer(q, ep, K, 0)
	return W.TopKMinVal(K)
}
