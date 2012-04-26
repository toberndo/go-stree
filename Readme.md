# go-stree

go-stree is a package for Go that can be used to process a large number of intervals.
The main purpose of this module is to solve the following problem: given a set of intervals, how to find all overlapping intervals at a certain point or within a certain range.

It offers three different algorithms:
- *stree*: implemented as segment tree
- *serial*: simple sequential algorithm, mainly for testing purposes
- *mtree*: implemented as segment tree with parallel processing

All three algorithms implement the following interface:
```go
// Main interface to access tree
type Tree interface {
  // Push new interval to stack
  Push(from, to int)
  // Push array of intervals to stack
  PushArray(from, to []int)
  // Clear the interval stack
  Clear()
  // Build segment tree out of interval stack
  BuildTree()
  // Print tree recursively to sdout
  Print()
  // Transform tree to array
  Tree2Array() []SegmentOverlap
  // Query interval
  Query(from, to int) []Interval
  // Query interval array
  QueryArray(from, to []int) []Interval
}
```

## Installation

    go get github.com/toberndo/go-stree/stree

## Example

```go
import (
  "fmt"
  "github.com/toberndo/go-stree/stree"
)

func main() {
    tree := stree.NewTree()
    tree.Push(1, 1)
    tree.Push(2, 3)
    tree.Push(5, 7)
    tree.Push(4, 6)
    tree.Push(6, 9)
    tree.BuildTree()
    fmt.Println(tree.Query(3, 5))
}
```

The serial algorithm resides in the same package:

```go
import (
  "github.com/toberndo/go-stree/stree"
)

func main() {
    serial := stree.NewSerial()
}
```

A parallel version of the segment tree is in the sub package *multi*:

```go
import (
  "github.com/toberndo/go-stree/stree/multi"
)

func main() {
    mtree := multi.NewMTree()
}
```

## Segment tree

A [segment tree](http://en.wikipedia.org/wiki/Segment_tree) is a data structure that can be used to run range queries on large sets of intervals. This can be used for example to analyze data of gene sequences.
The usage is as in the example above: we build a new tree object, push intervals to the data structure, build the tree and can then run certain queries on the tree. The segment tree is a static structure which means we cannot add further intervals once the tree is built. Rebuilding the tree is then required.

![segment tree example](http://assets.yarkon.de/images/Segment_tree_instance.gif)

## Serial

The sequential algorithm simply traverses the array of intervals to search for overlaps. It builds up a dynamic structure where intervals can be added at any time. The interface is equal to the segment tree, but tree specific methods like BuildTree(), Print() and Tree2Array() are not supported.

## API

See http://go.pkgdoc.org/github.com/toberndo/go-stree/stree

## Performance

To test performance execute the following command in directories *stree* and *multi*:
    go test -test.bench "." -test.cpu 4
As a short summary: the performance depends highly on the quality of the test data. Parallelism does not always improve performance, in some scenarios the stree algorithm is faster. In the optimal case mtree version with parallel support performs 20% better on a dual core machine than single threaded stree version.

## Licence

Use of this source code is governed by a BSD-style license that can be found in the LICENSE file.

## About

written by Thomas Oberndörfer <toberndo@yarkon.de>
follow me on [Twitter](https://twitter.com/#!/toberndo)