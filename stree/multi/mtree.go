// Copyright 2012 Thomas Oberndörfer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package implements a segment tree that uses parallel
// processing with multiple concurrent goroutines
package multi

import (
	. "github.com/toberndo/go-stree/stree"
	"math"
	"runtime"
	"sync"
)

const (
	// number of goroutines = 2 ** P_LEVEL
	P_LEVEL = 6 // 64 goroutines
)

// number of goroutines for tree walker
var NUM_WORKER int = runtime.NumCPU() * 2

type mtree struct {
	// Number of intervals
	count int
	root  *mnode
	// Interval stack
	base []Interval
	// Min value of all intervals
	min int
	// Max value of all intervals
	max int
	// channel to signal goroutine is done
	done chan bool
	// channel to limit number of running goroutines
	sem chan int
	// max number of goroutines used
	numG int
	// fallback to single processing if low number of intervals
	single bool
}

type mnode struct {
	// A segment is a interval represented by the node
	segment     Segment
	left, right *mnode
	// All intervals that overlap with segment
	overlap []*Interval
	// lock node for concurrent write access
	lock sync.Mutex
}

func (n *mnode) Segment() Segment {
	return n.segment
}

func (n *mnode) Left() Node {
	return n.left
}

func (n *mnode) Right() Node {
	return n.right
}

// Overlap transforms []*Interval to []Interval
func (n *mnode) Overlap() []Interval {
	if n.overlap == nil {
		return nil
	}
	interval := make([]Interval, len(n.overlap))
	for i, pintrvl := range n.overlap {
		interval[i] = *pintrvl
	}
	return interval
}

// NewMTree returns a Tree interface with underlying parallel segment tree implementation
func NewMTree() Tree {
	t := new(mtree)
	t.Clear()
	return t
}

// Push new interval to stack
func (t *mtree) Push(from, to int) {
	t.base = append(t.base, Interval{t.count, Segment{from, to}})
	t.count++
}

// Push array of intervals to stack
func (t *mtree) PushArray(from, to []int) {
	for i := 0; i < len(from); i++ {
		t.Push(from[i], to[i])
	}
}

// Clear the interval stack
func (t *mtree) Clear() {
	t.count = 0
	t.root = nil
	t.base = make([]Interval, 0, 100)
	t.min = 0
	t.max = 0
	// max number of goroutines = 2 ** P_LEVEL
	t.numG = int(math.Pow(2, P_LEVEL))
	// buffered channels
	t.done = make(chan bool, t.numG)
	t.sem = make(chan int, t.numG)
	// default: parallel processing
	t.single = false
}

// Build segment tree out of interval stack
func (t *mtree) BuildTree() {
	if len(t.base) == 0 {
		panic("No intervals in stack to build tree. Push intervals first")
	}
	var endpoint []int
	// attempts to parallelize the creation of endpoint array
	// only showed decrease in performance
	endpoint, t.min, t.max = Endpoints(t.base)
	// number of endpoints must be at least 10 times higher than number of
	// goroutines to justify effort and avoid locking situation
	if len(endpoint) < t.numG*10 {
		t.single = true
	}
	// create tree nodes from interval endpoints, uses goroutines if t.single == false
	t.root = t.insertNodes(endpoint, 0)
	if !t.single {
		// wait for goroutines to finish
		t.wait()
		// insert intervals using multi processing
		t.insertIntervalM()
	} else {
		// fall back for single processing
		for i := range t.base {
			t.insertInterval(t.root, &t.base[i])
		}
	}
}

func (t *mtree) wait() {
	for i := 0; i < t.numG; i++ {
		<-t.done
	}
}

func (t *mtree) Print() {
	Print(t.root)
}

func (t *mtree) Tree2Array() []SegmentOverlap {
	return Tree2Array(t.root)
}

// insertNodes builds tree structure from given endpoints
// starts with single processing, at P_LEVEL level of tree the children
// are created in seperate goroutines
func (t *mtree) insertNodes(endpoint []int, level int) *mnode {
	var n *mnode
	//fmt.Printf("Level: %d\n", level)
	if len(endpoint) == 1 {
		n = &mnode{segment: Segment{endpoint[0], endpoint[0]}}
		n.left = nil
		n.right = nil
	} else if len(endpoint) == 2 {
		n = &mnode{segment: Segment{endpoint[0], endpoint[1]}}
		if endpoint[1] != t.max {
			n.left = &mnode{segment: Segment{endpoint[0], endpoint[0]}}
			n.right = &mnode{segment: Segment{endpoint[1], endpoint[1]}}
		}
	} else {
		n = &mnode{segment: Segment{endpoint[0], endpoint[len(endpoint)-1]}}
		center := len(endpoint) / 2
		level++
		if level == P_LEVEL && !t.single {
			t.insertNodesAsync(&n.left, endpoint[:center+1], level)
			t.insertNodesAsync(&n.right, endpoint[center+1:], level)
		} else {
			n.left = t.insertNodes(endpoint[:center+1], level)
			n.right = t.insertNodes(endpoint[center+1:], level)
		}
	}
	return n
}

// insertNodesAsync starts new goroutine for creation of tree branch
func (t *mtree) insertNodesAsync(ppNode **mnode, endpoint []int, level int) {
	go func() {
		*ppNode = t.insertNodes(endpoint, level)
		t.done <- true
	}()
}

// Insert intervals with multiple goroutines
func (t *mtree) insertIntervalM() {
	for i := range t.base {
		// create new goroutines as long as space in buffer
		t.sem <- 1
		go func(index int) {
			t.insertInterval(t.root, &t.base[index])
			// release one entry in buffer when goroutine finishes
			<-t.sem
		}(i)
	}
	// wait for running goroutines to finish
	for i := 0; i < t.numG; i++ {
		t.sem <- 1
	}
}

// Inserts interval into given tree structure, write access locked
func (t *mtree) insertInterval(node *mnode, intrvl *Interval) {
	switch node.segment.CompareTo(&intrvl.Segment) {
	case SUBSET:
		node.lock.Lock()
		// interval of node is a subset of the specified interval or equal
		if node.overlap == nil {
			node.overlap = make([]*Interval, 0, 10)
		}
		node.overlap = append(node.overlap, intrvl)
		node.lock.Unlock()
	case INTERSECT_OR_SUPERSET:
		// interval of node is a superset, have to look in both children
		if node.left != nil {
			t.insertInterval(node.left, intrvl)
		}
		if node.right != nil {
			t.insertInterval(node.right, intrvl)
		}
	case DISJOINT:
		// nothing to do
	}
}

// A tree walker for querying intervals
type twalker struct {
	// number of goroutines
	num int
	// wait until goroutines are finished
	wait *sync.WaitGroup
	// queue where each buffer entry represents an available goroutine
	queue chan byte
	// result map of intervals
	result chan *map[int]Interval
}

// init with max number of goroutines
func (t *twalker) init(num int) {
	t.num = num
	t.wait = new(sync.WaitGroup)
	t.queue = make(chan byte, num)
	t.result = make(chan *map[int]Interval, num)
}

// collect results from goroutines
func (t *twalker) collect(result *map[int]Interval) {
	// wait for all to finish
	t.wait.Wait()
	for i := 0; i < t.num; i++ {
		// the number of started goroutines might be lower than the max number of goroutines
		// therefore this construct to break out if t.result is empty
		select {
		case rmap := <-t.result:
			for key, value := range *rmap {
				(*result)[key] = value
			}
		default:
			break
		}
	}
}

// Query interval with parallel tree walker
func (t *mtree) Query(from, to int) []Interval {
	if t.root == nil {
		panic("Can't run query on empty tree. Call BuildTree() first")
	}
	result := make(map[int]Interval)
	tw := new(twalker)
	tw.init(NUM_WORKER)
	querySingle(t.root, from, to, &result, tw, false)
	tw.collect(&result)
	sl := make([]Interval, 0, len(result))
	for _, intrvl := range result {
		sl = append(sl, intrvl)
	}
	return sl
}

// querySingle traverses tree in parallel to search for overlaps
func querySingle(node *mnode, from, to int, result *map[int]Interval, tw *twalker, back bool) {
	if !node.segment.Disjoint(from, to) {
		for _, pintrvl := range node.overlap {
			(*result)[pintrvl.Id] = *pintrvl
		}
		if node.right != nil {
			// buffered channel tw.queue is a safe counter to limit number of started goroutines
			select {
			case tw.queue <- 1:
				// create new map for result
				newMap := make(map[int]Interval)
				// increment counter of wait group
				tw.wait.Add(1)
				// start new query in goroutine
				go querySingle(node.right, from, to, &newMap, tw, true)
			default:
				// pass-through result map of parent
				querySingle(node.right, from, to, result, tw, false)
			}
		}
		if node.left != nil {
			select {
			case tw.queue <- 1:
				newMap := make(map[int]Interval)
				tw.wait.Add(1)
				go querySingle(node.left, from, to, &newMap, tw, true)
			default:
				querySingle(node.left, from, to, result, tw, false)
			}
		}
	}
	// if back is true then this method was called with go
	if back {
		// pass the result in the channel
		tw.result <- result
		// let wait group know that we are done
		tw.wait.Done()
	}
}

// Query interval array in parallel
func (t *mtree) QueryArray(from, to []int) []Interval {
	if t.root == nil {
		panic("Can't run query on empty tree. Call BuildTree() first")
	}
	result := make(map[int]Interval)
	tw := new(twalker)
	tw.init(NUM_WORKER)
	queryMulti(t.root, from, to, &result, tw, false)
	tw.collect(&result)
	sl := make([]Interval, 0, len(result))
	for _, intrvl := range result {
		sl = append(sl, intrvl)
	}
	return sl
}

// queryMulti traverses tree parallel in search of overlaps with multiple intervals
func queryMulti(node *mnode, from, to []int, result *map[int]Interval, tw *twalker, back bool) {
	hitsFrom := make([]int, 0, 2)
	hitsTo := make([]int, 0, 2)
	for i, fromvalue := range from {
		if !node.segment.Disjoint(fromvalue, to[i]) {
			for _, pintrvl := range node.overlap {
				(*result)[pintrvl.Id] = *pintrvl
			}
			hitsFrom = append(hitsFrom, fromvalue)
			hitsTo = append(hitsTo, to[i])
		}
	}
	// search in children only with overlapping intervals of parent
	if len(hitsFrom) != 0 {
		if node.right != nil {
			// buffered channel tw.queue is a safe counter to limit number of started goroutines
			select {
			case tw.queue <- 1:
				// create new map for result
				newMap := make(map[int]Interval)
				// increment counter of wait group
				tw.wait.Add(1)
				// start new query in goroutine
				go queryMulti(node.right, from, to, &newMap, tw, true)
			default:
				// pass-through result map of parent
				queryMulti(node.right, from, to, result, tw, false)
			}
		}
		if node.left != nil {
			select {
			case tw.queue <- 1:
				newMap := make(map[int]Interval)
				tw.wait.Add(1)
				go queryMulti(node.left, from, to, &newMap, tw, true)
			default:
				queryMulti(node.left, from, to, result, tw, false)
			}
		}
	}
	// if back is true then this method was called with go
	if back {
		// pass the result in the channel
		tw.result <- result
		// let wait group know that we are done
		tw.wait.Done()
	}
}
