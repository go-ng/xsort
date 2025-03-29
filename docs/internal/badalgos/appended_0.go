package sort

import (
	"fmt"

	"github.com/go-ng/container/heap"
	"github.com/go-ng/slices"
	"github.com/go-ng/sort"
)

type Interface[E any] sort.Interface[E]

func heapAppendSort[E any, S Interface[E]](s S, tailLength uint) {
	// Strategy:
	//
	// Basically this is a heap sort, which starts not from scratch, but
	// from a state where some elements are already pulled from the heap.
	//
	// But using of heap is a bit slow, so we combine this idea with
	// splitting the right part to two areas: the heap area and a sorted area.
	//
	// When possible we use the sorted area to pull the values to the left part,
	// otherwise we pull from the heap area. It allows to cheaply reduce amount
	// of calls to expensive heap.Pop.
	if tailLength == 0 {
		return
	}

	splitIdx := len(s) - int(tailLength)
	if splitIdx == 0 {
		sort.Sort(s)
		return
	}

	// Start with an empty heap and a sorted right part:

	rightPart := s[splitIdx:]
	sort.Sort(rightPart)
	if !s.Less(splitIdx, splitIdx-1) {
		return
	}

	h := s[splitIdx:splitIdx]

	// Find the first element we need to modify in the left part

	modifyStart := sort.Search(splitIdx, func(i int) bool {
		return s.Less(splitIdx, i)
	})
	if modifyStart >= splitIdx {
		// This case should be covered by "if !s.Less(splitIdx, splitIdx-1) {"
		// blow above.
		panic(fmt.Sprintf("should not happen: %d: %v", modifyStart, s))
	}

	// Going the the left part:
	// On each iteration we see if the right part has a lower value and pull it
	// if there is. If not, just continue (go to the next item of the left part).

	length := len(s)
	rightIdx := splitIdx
	hasRight := true
	hasHeap := false
	for idx := modifyStart; idx < splitIdx; idx++ {
		if !hasRight && !hasHeap {
			break
		}
		availFromHeap := hasHeap && s.Less(splitIdx, idx)
		availFromRight := hasRight && s.Less(rightIdx, idx)
		if !availFromRight && !availFromHeap {
			continue
		}
		old := s[idx]

		// pull from the heap only if it has the least value
		if availFromHeap && (!availFromRight || s.Less(splitIdx, rightIdx)) {
			s[idx] = heap.Pop(&h)
			hasHeap = len(h) > 0
		} else {
			// otherwise pull from the pre-sorted right part
			availFromRight = rightIdx+1 < length && s.Less(rightIdx+1, idx)
			s[idx] = s[rightIdx]
			rightIdx++
			hasRight = rightIdx < length
		}
		// put the old value to the pre-sorted right part (to avoid
		// pulling it back from the expensive heap if unnecessary)
		if !availFromRight {
			rightIdx--
			hasRight = true
			s[rightIdx] = old
			continue
		}
		heap.Push(&h, old)
		hasHeap = true
	}

	if len(h) == 0 {
		return
	}

	if len(h) == 1 && s.Less(splitIdx, rightIdx) {
		return
	}

	// At this moment we have a situation when left part if good to go,
	// but the right part has the heap on its left and an sorted slice
	// on its right. So either we just run traditional Sort or...

	if !shouldUseAppended(tailLength, uint(len(h))) {
		sort.Sort(rightPart)
		return
	}

	// ... or we swapping the parts of the right part and reusing the
	// same heapAppendSort.

	slices.Rotate(s[splitIdx:], int(tailLength)-len(h))
	heapAppendSort(s[splitIdx:], uint(len(h)))
}

// shouldUseAppended returns true if Appended is a more optimal
// sorter than Slice.
//
//   - totalSize is the size of the slice to be sorted.
//   - tailSize is the size of the unsorted right part (while the left
//     part is already sorted).
func shouldUseAppended(totalSize, tailSize uint) bool {
	// TODO: improve this
	switch {
	case totalSize <= 4:
		return false
	case totalSize <= 128:
		return 8*tailSize < totalSize
	case totalSize <= 4096:
		return tailSize*tailSize*4 < totalSize
	}
	return tailSize < 64
}
