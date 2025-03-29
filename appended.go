// This file is available under CC-0 1.0 license.
//
// See file `CC0-LICENSE`.

package xsort

import (
	"fmt"

	"github.com/go-ng/slices"
	"github.com/go-ng/sort"
)

type Interface[E any] sort.Interface[E]

// Appended sort a slice in assumption that the beginning of the slice is
// already sorted, and only tailLength new unsorted elements are added
// to the end of the slice. The function contains an optimization for
// a case where tailLength is low and it is possible to resort the slice
// faster than just through a full resorting.
//
// For example if there is a slice of length 65536 with only 512 unsorted
// elements in the end, then `Appended` on my laptop works 7 times faster
// than just using `Slice` on the whole slice.
//
// The algorithm is good if the unsorted right part of the array is much
// smaller than the sorted left part (see `docs/force_appended_benchmark.txt`),
// but otherwise it is worse than a simple quick sort. So it fallbacks to
// quicksort if the unsorted part is too big.
//
// Roughly:
//
// T: O(k*ln(n) + n + k^2) -- thus if `k` is too high then: O(k^2)
//
// S: O(1) [if without `s`]
func Appended[E any, S Interface[E]](s S, tailLength uint) {
	if tailLength == 0 {
		return
	}

	if !shouldUseAppended(uint(len(s)), tailLength) {
		if tailLength > uint(len(s)) {
			panic(fmt.Sprintf("tailLength (%d) cannot be greater than the length of the provided slice (%d)", tailLength, len(s)))
		}

		sort.Sort(s)
		return
	}

	groupInsertAppendSort(s, tailLength)
}

// AppendedWithBuf is the same as Appended but:
// * Much faster.
// * Requires a buffer.
//
// The buffer length should be exactly the same as the length
// of the unsorted tail.
//
// For example if there is a slice of length 65536 with only 512 unsorted
// elements in the end, then `Appended` on my laptop works ~30 times faster
// than just using `Slice` on the whole slice.
//
// T: O(k*ln(n) + n)
//
// S: O(k) [if without `s`]
func AppendedWithBuf[E any, S Interface[E]](s S, buf []E) {
	tailLength := uint(len(buf))
	if tailLength == 0 {
		return
	}

	if !shouldUseAppendedWithBuf(uint(len(s)), tailLength) {
		if tailLength > uint(len(s)) {
			panic(fmt.Sprintf("tailLength (%d) cannot be greater than the length of the provided slice (%d)", tailLength, len(s)))
		}

		sort.Sort(s)
		return
	}

	groupInsertAppendSortWithBuf(s, buf)
}

func groupInsertAppendSort[E any, S Interface[E]](s S, tailLength uint) {
	// Strategy:
	//
	// This is basically an insert search, which:
	// * Moves the array by pieces to avoid k iterations through the whole array.
	// * Finds the place to insert through a binary search.
	//
	// See also groupInsertAppendSortWithBuf, it implements the same idea, but
	// it is easier to read. The difference is groupInsertAppendSortWithBuf
	// stores the unsorted values in an external storage, which allows avoiding
	// slice rotations, and just do the "move" (/copy) directly.
	length := len(s)
	if int(tailLength) > length {
		panic(fmt.Errorf("tail is longer than the slice: %d > %d", tailLength, len(s)))
	}
	splitIdx := uint(length) - tailLength
	if splitIdx == 0 {
		sort.Sort(s)
		return
	}
	rightPart := s[splitIdx:]
	sort.Slice(rightPart, func(i, j int) bool {
		return rightPart.Less(j, i)
	})

	unsortedStartIdx := splitIdx
	unsortedEnd := length
	for unsortedCount := tailLength; unsortedCount > 0; unsortedCount-- {
		leftIdx := sort.Search(int(unsortedStartIdx), func(i int) bool {
			return s.Less(int(unsortedStartIdx), i)
		})

		if leftIdx == int(unsortedStartIdx) {
			if unsortedStartIdx == 0 {
				slices.Reverse(s[0:unsortedCount])
				break
			}
			if leftIdx > 0 {
				leftIdx--
			}
			if s.Less(int(unsortedStartIdx), int(unsortedStartIdx)-1) {
				slices.Rotate(s[leftIdx:leftIdx+int(unsortedCount)+1], -2)
				unsortedStartIdx = uint(leftIdx)
			} else {
				slices.Rotate(s[leftIdx+1:leftIdx+int(unsortedCount)+1], -1)
				unsortedStartIdx = uint(leftIdx) + 1
			}
		} else {
			slices.Rotate(s[leftIdx+1:unsortedEnd], unsortedEnd-int(unsortedStartIdx))
			s[leftIdx], s[leftIdx+1] = s[leftIdx+1], s[leftIdx]
			slices.Rotate(s[leftIdx:leftIdx+int(unsortedCount)+1], -2)
			unsortedStartIdx = uint(leftIdx)
		}
		unsortedEnd = int(unsortedStartIdx) + int(unsortedCount) - 1
	}
}

func groupInsertAppendSortWithBuf[E any, S Interface[E]](s S, buf []E) {
	// Strategy:
	//
	// This is basically an insert search, which:
	// * Moves the array by pieces to avoid k iterations through the whole array.
	// * Finds the place to insert through a binary search.
	//
	// See also groupInsertAppendSort: "groupInsertAppendSort" is an in-place
	// implementation of the same idea, which on each iteration moves the bunch
	// of unsorted items closer to the left (thus it also have "k^2" in the time
	// complexity).
	tailLength := len(buf)
	length := len(s)
	if int(tailLength) > length {
		panic(fmt.Errorf("tail is longer than the slice: %d > %d", tailLength, len(s)))
	}
	splitIdx := length - tailLength
	if splitIdx == 0 {
		sort.Sort(s)
		return
	}
	rightPart := s[splitIdx:]
	sort.Sort(rightPart)
	copy(buf, rightPart)

	unsortedStartIdx := splitIdx
	unsortedEnd := length
	for unsortedCount := tailLength; unsortedCount > 0; unsortedCount-- {
		s[unsortedStartIdx] = buf[unsortedCount-1]
		leftIdx := sort.Search(int(unsortedStartIdx), func(i int) bool {
			return s.Less(unsortedStartIdx, i)
		})

		copyTo := leftIdx + unsortedCount
		// 1 3 5 7 9 4 6
		copy(s[copyTo:unsortedEnd], s[leftIdx:])
		// 1 3 5 7 9 7 9
		s[leftIdx+unsortedCount-1] = buf[unsortedCount-1]
		// 1 3 5 7 6 7 9

		unsortedStartIdx = leftIdx
		unsortedEnd = int(unsortedStartIdx) + int(unsortedCount) - 1
	}
}

// shouldUseAppended returns true if Appended is a more optimal
// sorter than Slice.
//
//   - totalSize is the size of the slice to be sorted.
//   - tailSize is the size of the unsorted right part (while the left
//     part is already sorted).
func shouldUseAppended(totalSize, tailSize uint) bool {
	// 32: 8
	// 64: 16
	// 128: 32-64
	// 256: 64
	// 512: 128
	// 1024: 128-256
	// ...
	// 32768: 1024-2048
	// 65536: 2048
	// 131072: 2048-4096
	// 262144: 4096
	// 524288: 4096-8192
	// 1048576: 8192
	switch {
	case totalSize < 512: // k is too small an the "k^2" is not dominating yet
		return tailSize*4 < totalSize
	default:
		return tailSize*tailSize/64 < totalSize // now "k^2" is dominating
	}
}

// shouldUseAppendedWithBuf returns true if AppendedWithBuf is a more optimal
// sorter than Slice.
//
//   - totalSize is the size of the slice to be sorted.
//   - tailSize is the size of the unsorted right part (while the left
//     part is already sorted).
func shouldUseAppendedWithBuf(totalSize, tailSize uint) bool {
	switch {
	case totalSize < 10:
		return false
	case totalSize < 64:
		return tailSize*3 < totalSize
	case totalSize < 256:
		return tailSize*2 < totalSize
	default:
		return tailSize*5 < totalSize*3
	}
}
