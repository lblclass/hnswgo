package hnswheap

import (
	"github.com/lblclass/hnswgo/models"
)

// 定义一个类型，满足 heap.Interface
type BigCandidates []models.Candidate

func (bc BigCandidates) Len() int { return len(bc) }

// 实现 sort.Interface 的 Less 方法，将比较逻辑反转即可实现大顶堆
func (bc BigCandidates) Less(i, j int) bool {
	return bc[i].Distance > bc[j].Distance // 大于号表示最大值优先
}

// 实现 sort.Interface 的 Swap 方法
func (bc BigCandidates) Swap(i, j int) {
	bc[i], bc[j] = bc[j], bc[i]
}

// 实现 heap.Interface 的 Pusbc 方法
func (bc *BigCandidates) Push(x any) {
	*bc = append(*bc, x.(models.Candidate))
}

// 实现 heap.Interface 的 Pop 方法
func (bc *BigCandidates) Pop() any {
	old := *bc
	n := len(old)
	x := old[n-1]
	*bc = old[0 : n-1]
	return x
}

// PriorityQueue implements a min-heap for sorting candidates.
type SmallCandidates []models.Candidate

func (sc SmallCandidates) Len() int { return len(sc) }
func (sc SmallCandidates) Less(i, j int) bool {
	return sc[i].Distance < sc[j].Distance
}
func (sc SmallCandidates) Swap(i, j int) { sc[i], sc[j] = sc[j], sc[i] }

func (sc *SmallCandidates) Push(x interface{}) {
	*sc = append(*sc, x.(models.Candidate))
}

func (sc *SmallCandidates) Pop() interface{} {
	old := *sc
	n := len(old)
	item := old[n-1]
	*sc = old[0 : n-1]
	return item
}
