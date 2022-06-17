// This file is available under CC-0 1.0 license.
//
// See file `CC0-LICENSE`.

package xsort

import (
	"fmt"
	"math/rand"
	stdsort "sort"
	"strings"
	"testing"

	"github.com/go-ng/sort"
)

func intsEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for idx := range a {
		if a[idx] != b[idx] {
			return false
		}
	}
	return true
}

func prepareTestCase(initial []byte, tailLenght uint) ([]int, []string, []string, string) {
	splitIdx := len(initial) - int(tailLenght)
	stdsort.Slice(initial[:splitIdx], func(i, j int) bool {
		return initial[i] < initial[j]
	})
	var leftStrs, rightStrs []string
	for _, v := range initial[:len(initial)-int(tailLenght)] {
		leftStrs = append(leftStrs, fmt.Sprintf("%d", v))
	}
	for _, v := range initial[len(initial)-int(tailLenght):] {
		rightStrs = append(rightStrs, fmt.Sprintf("%d", v))
	}
	slice := make([]int, len(initial))
	for idx, v := range initial {
		slice[idx] = int(v)
	}
	return slice, leftStrs, rightStrs, fmt.Sprintf("%s | %s (tailLength: %d)", strings.Join(leftStrs, " "), strings.Join(rightStrs, " "), tailLenght)
}

func testAppended(t *testing.T, initial []byte, tailLenght uint) {
	s, leftStrs, rightStrs, testName := prepareTestCase(initial, tailLenght)
	c := make([]int, len(s))
	copy(c, s)
	t.Run(testName, func(t *testing.T) {
		Appended(stdsort.IntSlice(s), uint(tailLenght))
		stdsort.Slice(c, func(i, j int) bool {
			return c[i] < c[j]
		})
		if !intsEqual(c, s) {
			t.Fatalf("%v != %v; testCase < %s , %s >", c, s, strings.Join(leftStrs, ","), strings.Join(rightStrs, ","))
		}
	})
}

func TestAppended(t *testing.T) {
	testAppended(t, []byte{1, 3, 5, 7, 11, 13, 12, 6, 4, 8}, 4)
	testAppended(t, []byte{0, 0, 2, 5, 8, 8, 9, 10, 10, 11, 11, 15, 11, 12, 8, 14}, 4)
}

func FuzzAppended(f *testing.F) {
	f.Fuzz(func(t *testing.T, initial, _ []byte) {
		tailLenght := uint(rand.Intn(len(initial) + 1))
		testAppended(t, initial, uint(tailLenght))
	})
}

func testAppended2(t *testing.T, initial []byte, tailLenght uint) {
	s, leftStrs, rightStrs, testName := prepareTestCase(initial, tailLenght)
	c := make([]int, len(s))
	copy(c, s)
	t.Run(testName, func(t *testing.T) {
		groupInsertAppendSort(stdsort.IntSlice(s), tailLenght)
		stdsort.Slice(c, func(i, j int) bool {
			return c[i] < c[j]
		})
		if !intsEqual(c, s) {
			t.Fatalf("%v != %v; testCase < %s , %s >", c, s, strings.Join(leftStrs, ","), strings.Join(rightStrs, ","))
		}
	})
}

func TestAppended2(t *testing.T) {
	testAppended2(t, []byte{1, 3, 5, 7, 11, 13, 12, 6, 4, 8}, 4)
	testAppended2(t, []byte{0, 0, 2, 5, 8, 8, 9, 10, 10, 11, 11, 15, 11, 12, 8, 14}, 4)
	testAppended2(t, []byte{49, 255, 127}, 2)
	testAppended2(t, []byte{65, 76, 173, 37, 67, 145}, 5)
}

func FuzzAppended2(f *testing.F) {
	f.Fuzz(func(t *testing.T, initial, _ []byte) {
		tailLenght := uint(rand.Intn(len(initial) + 1))
		testAppended2(t, initial, uint(tailLenght))
	})
}

func testAppended3(t *testing.T, initial []byte, tailLenght uint) {
	s, leftStrs, rightStrs, testName := prepareTestCase(initial, tailLenght)
	c := make([]int, len(s))
	copy(c, s)
	t.Run(testName, func(t *testing.T) {
		groupInsertAppendSortWithBuf(stdsort.IntSlice(s), make([]int, tailLenght))
		stdsort.Slice(c, func(i, j int) bool {
			return c[i] < c[j]
		})
		if !intsEqual(c, s) {
			t.Fatalf("%v != %v; testCase < %s , %s >", c, s, strings.Join(leftStrs, ","), strings.Join(rightStrs, ","))
		}
	})
}

func TestAppended3(t *testing.T) {
	testAppended3(t, []byte{1, 3, 5, 7, 11, 13, 12, 6, 4, 8}, 4)
	testAppended3(t, []byte{0, 0, 2, 5, 8, 8, 9, 10, 10, 11, 11, 15, 11, 12, 8, 14}, 4)
	testAppended3(t, []byte{49, 255, 127}, 2)
	testAppended3(t, []byte{65, 76, 173, 37, 67, 145}, 5)
}

func FuzzAppended3(f *testing.F) {
	f.Fuzz(func(t *testing.T, initial, _ []byte) {
		tailLenght := uint(rand.Intn(len(initial) + 1))
		testAppended3(t, initial, uint(tailLenght))
	})
}

type intSlice []int

func (s intSlice) Less(i, j int) bool {
	return s[i] < s[j]
}

func BenchmarkAppended(b *testing.B) {
	for totalSize := 1; totalSize <= 2*1024*1024; {
		for tailSize := 1; tailSize <= totalSize; {
			b.Run(fmt.Sprintf("total-%d/tail-%d", totalSize, tailSize), func(b *testing.B) {
				csCount := 1000/(totalSize+1) + 20

				rng := rand.New(rand.NewSource(0))
				in := make([][]int, csCount)
				for idx := range in {
					in[idx] = make([]int, totalSize)
					s := in[idx]
					for idx := range s {
						s[idx] = rng.Intn(totalSize)
					}
					sort.Slice(s[:len(s)-tailSize], func(i, j int) bool {
						return s[i] < s[j]
					})
				}

				cs := make([]intSlice, csCount)
				for idx := range cs {
					cs[idx] = make([]int, totalSize)
				}

				b.Run("Slice", func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						idx := i % csCount
						if idx == 0 {
							b.StopTimer()
							for idx := range cs {
								copy(cs[idx], in[idx])
							}
							b.StartTimer()
						}
						s := cs[idx]
						sort.Slice(s, func(i, j int) bool {
							return s[i] < s[j]
						})
					}
				})
				b.Run("Sort", func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						idx := i % csCount
						if idx == 0 {
							b.StopTimer()
							for idx := range cs {
								copy(cs[idx], in[idx])
							}
							b.StartTimer()
						}
						sort.Sort(cs[idx])
					}
				})
				b.Run("sort.Slice", func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						idx := i % csCount
						if idx == 0 {
							b.StopTimer()
							for idx := range cs {
								copy(cs[idx], in[idx])
							}
							b.StartTimer()
						}
						s := cs[idx]
						sort.Slice(s, func(i, j int) bool {
							return s[i] < s[j]
						})
					}
				})
				b.Run("sort.Sort", func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						idx := i % csCount
						if idx == 0 {
							b.StopTimer()
							for idx := range cs {
								copy(cs[idx], in[idx])
							}
							b.StartTimer()
						}
						stdsort.Sort(stdsort.IntSlice(cs[idx]))
					}
				})
				b.Run("Appended", func(b *testing.B) {
					if tailSize > 200000 {
						b.Skip()
						return
					}
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						idx := i % csCount
						if idx == 0 {
							b.StopTimer()
							for idx := range cs {
								copy(cs[idx], in[idx])
							}
							b.StartTimer()
						}
						c := cs[idx]
						Appended(c, uint(tailSize))
					}
				})
				b.Run("AppendedWithBuf", func(b *testing.B) {
					buf := make([]int, uint(tailSize))
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						idx := i % csCount
						if idx == 0 {
							b.StopTimer()
							for idx := range cs {
								copy(cs[idx], in[idx])
							}
							b.StartTimer()
						}
						c := cs[idx]
						AppendedWithBuf(c, buf)
					}
				})
			})
			if tailSize == 0 {
				tailSize = 1
			} else {
				tailSize = tailSize*5/4 + 1
			}
		}
		if totalSize == 0 {
			totalSize = 1
		} else {
			totalSize *= 2
		}
	}
}
