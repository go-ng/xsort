package xsort

import (
	"sort"

	"golang.org/x/exp/constraints"
)

// OrderedDesc is a generic for a slice of "ordered" items that
// adds methods for sorting in the descending order.
type OrderedDesc[E constraints.Ordered] []E

var _ sort.Interface = (OrderedDesc[int])(nil)

func (s OrderedDesc[E]) Less(i, j int) bool {
	return s[i] > s[j]
}

func (s OrderedDesc[E]) Len() int {
	return len(s)
}

func (s OrderedDesc[E]) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
