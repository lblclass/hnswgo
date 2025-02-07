package hnswheap

import (
	"container/heap"

	"github.com/lblclass/hnswgo/models"
)

const (
	SMALL = "small"
	BIG   = "big"
)

// CandidateHeap is a generic heap for Candidates
type CandidateHeap struct {
	Candidates []models.Candidate
	Compare    string // default min
}

// Len is the number of elements in the collection
func (ch CandidateHeap) Len() int { return len(ch.Candidates) }

// Less reports whether the element with index i should sort before the element with index j
func (ch CandidateHeap) Less(i, j int) bool {
	res := 0.0
	if ch.Compare == BIG {
		res = (ch.Candidates[i].Distance - ch.Candidates[j].Distance)
	} else {
		res = (ch.Candidates[j].Distance - ch.Candidates[i].Distance)
	}
	return res > 0.0
}

// Swap swaps the elements with indexes i and j
func (ch CandidateHeap) Swap(i, j int) {
	ch.Candidates[i], ch.Candidates[j] = ch.Candidates[j], ch.Candidates[i]
}

// Push adds an element to the heap
func (ch *CandidateHeap) Push(x interface{}) {
	ch.Candidates = append(ch.Candidates, x.(models.Candidate))
}

// Pop removes and returns the last element of the heap
func (ch *CandidateHeap) Pop() interface{} {
	old := ch.Candidates
	n := len(old)
	item := old[n-1]
	ch.Candidates = old[0 : n-1]
	return item
}

// ExtractHeapData returns all node IDs in the heap without affecting the heap structure
func (ch *CandidateHeap) ExtractHeapData() []int {
	dataCopy := make([]int, len(ch.Candidates))
	for i, item := range ch.Candidates {
		dataCopy[i] = item.NodeID
	}
	return dataCopy
}

// get top value
func (ch *CandidateHeap) topValue(K int, topType string) []int {
	res := make([]int, K)
	if topType == ch.Compare {
		for i := 0; i < K; i++ {
			tmpRes := ch.Pop().(models.Candidate)
			res[i] = tmpRes.NodeID
		}
	} else {
		if K < ch.Len() {
			for j := 0; j < (ch.Len() - K); j++ {
				heap.Pop(ch)
			}
		}
		index := K
		if K > ch.Len() {
			index = ch.Len()
		}
		for j := (index - 1); j > 0; j-- {
			tmpRes := heap.Pop(ch).(models.Candidate)
			res[j] = tmpRes.NodeID
		}
	}
	return res
}

// get top K min value
func (ch *CandidateHeap) TopKMinVal(K int) []int {
	return ch.topValue(K, SMALL)
}

func (ch *CandidateHeap) TopKmaxVal(K int) []int {
	return ch.topValue(K, BIG)
}

// NewCandidateHeap creates a new CandidateHeap with a custom comparison function
func NewCandidateHeap(compare string) *CandidateHeap {
	return &CandidateHeap{
		Compare: compare,
	}
}

// Create a max-heap for BigCandidates based on Distance
func NewBigCandidatesHeap() *CandidateHeap {
	return NewCandidateHeap("big")
}

// Create a min-heap for SmallCandidates based on Distance
func NewSmallCandidatesHeap() *CandidateHeap {
	return NewCandidateHeap("small")
}
