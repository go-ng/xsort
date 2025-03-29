# About

[![GoDoc](https://godoc.org/github.com/go-ng/xsort?status.svg)](https://pkg.go.dev/github.com/go-ng/xsort?tab=doc)

This package is a collection of extra primitives related to sorting. Currently, it has only two functions:
* [`Appended(s []T, tailLength uint)`](https://pkg.go.dev/github.com/go-ng/xsort#Appended)
* [`AppendedWithBuf(s []T, buf []T)`](https://pkg.go.dev/github.com/go-ng/xsort#AppendedWithBuf)

These functions allow to quickly re-sort a previously sorted slice, which has few unsorted elements appended. `Appended` works pretty well if the tail is small, but limitations of the in-place algorithm makes it ineffective is the tail is comparable by size with the total size of the slice. While `AppendedWithBuf` is effective with 

![latencies](https://raw.githubusercontent.com/go-ng/docs/main/xsort/force_appended.png "latencies")

As one may see resorting 100 items in a slice of length 1048576 using `AppendedWithBuf` is **almost 100 times more effective** than just full resorting using `sort.Slice` or `sort.Sort`. But if the amount of unsorted items is too high then it might be less effective. So for convenience to automatically avoid spending more time than simple through `Sort`/`Slice` these functions has heuristic fallback logic for big tails, so the actual performance for a slice of size 1048576 looks like here:

![latencies](https://raw.githubusercontent.com/go-ng/docs/main/xsort/appended.png "latencies")

But using of these functions does not make sense if `tailLength` always equals to `len(s)`.

P.S.: Also for comparison here is performance for a slice of size 256:

![latencies](https://raw.githubusercontent.com/go-ng/docs/main/xsort/appended_256.png "latencies")


# Quick start

In-place:
```go
package main

import (
	"fmt"
	"sort"

	"github.com/go-ng/xsort"
)

func main() {
	s := sort.IntSlice{-4, -2, 1, 3, 4, 5, 9}

	s = append(s, 2, -3)
	xsort.Appended(s, 2)

	fmt.Println(s) // output: [-4 -3 -2 1 2 3 4 5 9]
}
```
[(sandbox)](https://play.golang.com/p/vF7ZE6KMquw)

With buffer (faster):
```go
package main

import (
	"fmt"
	"sort"

	"github.com/go-ng/xsort"
)

func main() {
	s := sort.IntSlice{-4, -2, 1, 3, 4, 5, 9}

	a := []int{2, -3}
	s = append(s, a...)
	xsort.AppendedWithBuf(s, a)

	fmt.Println(s) // output: [-4 -3 -2 1 2 3 4 5 9]
}
```
[(sandbox)](https://play.golang.com/p/DF_c82-ImFn)
