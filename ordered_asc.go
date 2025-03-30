package xsort

import (
	"sort"

	"golang.org/x/exp/constraints"
)

// OrderedAsc is a generic for a slice of "ordered" items that
// adds methods for sorting in the ascending order.
type OrderedAsc[E constraints.Ordered] []E

var _ sort.Interface = (OrderedAsc[int])(nil)

func (s OrderedAsc[E]) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s OrderedAsc[E]) Len() int {
	return len(s)
}

func (s OrderedAsc[E]) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
