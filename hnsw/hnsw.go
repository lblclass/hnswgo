package hnsw

import (
	"container/heap"
	"math"
	"math/rand"
	"sync"

	"github.com/lblclass/hnswgo/models"
	hnswheap "github.com/lblclass/hnswgo/util/heap"
)

// HNSW is the main struct representing the graph.
type HNSW struct {
	Layers          []map[int][]int // Connections at each layer
	EnterPoint      int             // Entry point ID
	InitConnects    int
	MaxConnections  int     // Maximum connections per element
	EfConstruction  int     // Candidate list size
	NormalizationML float64 // Level normalization factor
	MaxLayers       int
	Elements        map[int]models.Element // Element data
	mu              sync.RWMutex
}

// NewHNSW initializes an HNSW graph.
func NewHNSW(maxConnections, efConstruction, initConnects, maxLayers int, nm float64) *HNSW {
	return &HNSW{
		Layers:          []map[int][]int{},
		EnterPoint:      -1,
		InitConnects:    initConnects,
		MaxConnections:  maxConnections,
		EfConstruction:  efConstruction,
		MaxLayers:       maxLayers,
		NormalizationML: nm,
		Elements:        make(map[int]models.Element),
		mu:              sync.RWMutex{},
	}
}

/*
Output: update hnsw inserting element q
1 W ← ∅ // list for the currently found nearest elements
2 ep ← get enter point for hnsw
3 L ← level of ep // top layer for hnsw
4 l ← ⌊-ln(unif(0..1))∙mL⌋ // new element’s level
5 for lc ← L … l+1
6   W ← SEARCH-LAYER(q, ep, ef=1, lc)
7   ep ← get the nearest element from W to q
8 for lc ← min(L, l) … 0
9   W ← SEARCH-LAYER(q, ep, efConstruction, lc)
10  neighbors ← SELECT-NEIGHBORS(q, W, M, lc) // alg. 3 or alg. 4
11  add bidirectionall connectionts from neighbors to q at layer lc
12  for each e ∈ neighbors // shrink connections if needed
13   eConn ← neighbourhood(e) at layer lc
14   if │eConn│ > Mmax // shrink connections of e 	// if lc = 0 then Mmax = Mmax0
15     eNewConn ← SELECT-NEIGHBORS(e, eConn, Mmax, lc)  // alg. 3 or alg. 4
16     set neighbourhood(e) at layer lc to eNewConn
17   ep ← W
18 if l > L
19 set enter point for hnsw to q
*/

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
			h.Layers = append(h.Layers, map[int][]int{
				q.ID: {},
			})
		}
		h.EnterPoint = q.ID
	} else {
		// 如果topLevel大于level，则ep需要有些变化。
		for lc := topLevel; lc > level; lc-- {
			ep = h.searchLayer(q, ep, 1, lc)[0]
		}
	}

	for lc := min(topLevel, level); lc >= 0; lc-- {
		neighbors := h.searchLayer(q, ep, h.EfConstruction, lc)
		selected := h.SelectNeighborsHeuristic(q, neighbors, h.InitConnects, lc, true, true)

		// Connect bidirectionally.
		for _, n := range selected {
			h.addConnection(q.ID, n, lc)
			h.addConnection(n, q.ID, lc)
		}
		ep = neighbors[0]
	}

}

// SEARCH-LAYER(q, ep, ef, lc)
// Input: query element q, enter points ep, number of nearest to q elements to return ef, layer number lc
// Output: ef closest neighbors to q
// 1 v ← ep // set of visited elements
// 2 C ← ep // set of candidates
// 3 W ← ep // dynamic list of found nearest neighbors
// 4 while │C│ > 0
// 5   c ← extract nearest element from C to q
// 6   f ← get furthest element from W to q
// 7   if distance(c, q) > distance(f, q)
// 8     break // all elements in W are evaluated
// 9   for each e ∈ neighbourhood(c) at layer lc // update C and W
// 10    if e ∉ v
// 11      v ← v ⋃ e
// 12    f ← get furthest element from W to q
// 13    if distance(e, q) < distance(f, q) or │W│ < ef
// 14      C ← C ⋃ e
// 15      W ← W ⋃ e
// 16      if │W│ > ef
// 17         remove furthest element from W to q
// 18 return W

// searchLayer finds nearest neighbors in the specified layer.
func (h *HNSW) searchLayer(q models.Element, entryPoint int, ef int, lc int) []int {
	V := map[int]bool{entryPoint: true} // set of visited elements
	qCandidate := models.Candidate{
		NodeID:   entryPoint,
		Distance: h.distance(q, h.Elements[entryPoint]),
	}
	C := make(hnswheap.SmallCandidates, 0)
	heap.Init(&C)
	heap.Push(&C, qCandidate)
	W := make(hnswheap.BigCandidates, 0)
	heap.Init(&W)
	heap.Push(&W, qCandidate)
	for C.Len() > 0 {
		nc := heap.Pop(&C).(models.Candidate)
		fc := W[0]
		if nc.Distance > fc.Distance {
			break
		}
		for _, el := range h.Layers[lc][nc.NodeID] {
			if _, ok := V[el]; ok {
				continue
			}
			V[el] = true
			fc := W[0]
			if (fc.Distance < h.distance(q, h.Elements[el])) || (W.Len() < ef) {
				tmpC := models.Candidate{NodeID: el, Distance: h.distance(q, h.Elements[el])}
				heap.Push(&C, tmpC)
				heap.Push(&W, tmpC)
				if W.Len() > ef {
					heap.Pop(&W)
				}
			}
		}
	}
	// Simplified: replace with real search logic.
	return []int{entryPoint}
}

// selectNeighbors selects M nearest neighbors from the candidates.
func (h *HNSW) SelectNeighborsSimple(q models.Element, candidates []int) []int {
	// candidates is already ordered at select
	return candidates[:min(len(candidates), h.InitConnects)]
}

// addConnection adds a connection to the graph.
func (h *HNSW) addConnection(from, to, layer int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Layers[layer][from] = append(h.Layers[layer][from], to)
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

func (h *HNSW) distance(e1, e2 models.Element) float64 {
	sum := 0.0
	for i := range e1.Embeddings {
		diff := e1.Embeddings[i] - e2.Embeddings[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// SELECT-NEIGHBORS-HEURISTIC(q, C, M, lc, extendCandidates, keepPrunedConnections)
// Input: base element q, candidate elements C, number of neighbors to
// return M, layer number lc, flag indicating whether or not to extend
// candidate list extendCandidates, flag indicating whether or not to add
// discarded elements keepPrunedConnections
// Output: M elements selected by the heuristic
// 1 R ← ∅
// 2 W ← C // working queue for the candidates
// 3 if extendCandidates // extend candidates by their neighbors
// 4   for each e ∈ C
// 5     for each eadj ∈ neighbourhood(e) at layer lc
// 6       if eadj ∉ W
// 7         W ← W ⋃ eadj
// 8 Wd ← ∅ // queue for the discarded candidates
// 9 while │W│ > 0 and │R│< M
// 10  e ← extract nearest element from W to q
// 11  if e is closer to q compared to any element from R
// 12    R ← R ⋃ e
// 13  else
// 14    Wd ← Wd ⋃ e
// 15 if keepPrunedConnections // add some of the discarded // connections from Wd
// 16   while │Wd│> 0 and │R│< M
// 17     R ← R ⋃ extract nearest element from Wd to q
// 18 return R

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
	W := make(hnswheap.SmallCandidates, 0, len(candidates))
	heap.Init(&W)

	// Add initial candidates to the queue.
	for _, c := range candidates {
		dist := h.distance(q, h.Elements[c])
		heap.Push(&W, models.Candidate{NodeID: c, Distance: dist})
	}

	// Extend candidates by their neighbors if needed.
	if extendCandidates {
		neighbors := map[int]bool{}
		for _, c := range candidates {
			for _, neighbor := range h.Layers[layer][c] {
				if _, exists := neighbors[neighbor]; !exists {
					dist := h.distance(q, h.Elements[neighbor])
					heap.Push(&W, models.Candidate{NodeID: neighbor, Distance: dist})
					neighbors[neighbor] = true
				}
			}
		}
	}

	Wd := make(hnswheap.SmallCandidates, 0) // Discarded candidates
	heap.Init(&Wd)

	// Process candidates.
	for W.Len() > 0 && len(R) < M {
		e := heap.Pop(&W).(models.Candidate)
		closer := true
		for r := range R {
			if h.distance(h.Elements[e.NodeID], h.Elements[r]) < e.Distance {
				closer = false
				break
			}
		}
		if closer {
			R[e.NodeID] = true
		} else {
			heap.Push(&Wd, e)
		}
	}

	// Add pruned connections if required.
	if keepPrunedConnections {
		for Wd.Len() > 0 && len(R) < M {
			e := heap.Pop(&Wd).(models.Candidate)
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

// K-NN-SEARCH(hnsw, q, K, ef)
// Input: multilayer graph hnsw, query element q, number of nearest
// neighbors to return K, size of the dynamic candidate list ef
// Output: K nearest elements to q
// 1 W ← ∅ // set for the current nearest elements
// 2 ep ← get enter point for hnsw
// 3 L ← level of ep // top layer for hnsw
// 4 for lc ← L … 1
// 5   W ← SEARCH-LAYER(q, ep, ef=1, lc)
// 6   ep ← get nearest element from W to q
// 7 W ← SEARCH-LAYER(q, ep, ef, lc =0)
// 8 return K nearest elements from W to q

func (h *HNSW) KNNSearch(q models.Element, K int) []int {
	// W := make(hnswheap.SmallCandidates, 0)
	ep := h.EnterPoint
	totalLayers := len(h.Layers)
	for lc := (totalLayers - 1); lc >= 0; lc-- {
		W := h.searchLayer(q, ep, 1, lc)
		ep = W[0]
	}
	W := h.searchLayer(q, ep, h.EfConstruction, 0)
	return W
}
