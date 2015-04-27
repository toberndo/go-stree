// Copyright 2012 Thomas Oberndörfer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package implements a segment tree and serial algorithm to query intervals
package stree

import (
	"fmt"
	"reflect"
	"sort"
)

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
	// Print tree recursively to stdout
	Print()
	// Transform tree to array
	Tree2Array() []SegmentOverlap
	// Query interval
	Query(from, to int) []Interval
	// Query interval array
	QueryArray(from, to []int) []Interval
}

type stree struct {
	// Number of intervals
	count int
	root  *node
	// Interval stack
	base []Interval
	// Min value of all intervals
	min int
	// Max value of all intervals
	max int
}

// Interface to provide unified access to nodes
type Node interface {
	Segment() Segment
	Left() Node
	Right() Node
	Overlap() []Interval
}

type node struct {
	// A segment is a interval represented by the node
	segment     Segment
	left, right *node
	// All intervals that overlap with segment
	overlap []*Interval
}

func (n *node) Segment() Segment {
	return n.segment
}

func (n *node) Left() Node {
	return n.left
}

func (n *node) Right() Node {
	return n.right
}

// Overlap transforms []*Interval to []Interval
func (n *node) Overlap() []Interval {
	if n.overlap == nil {
		return nil
	}
	interval := make([]Interval, len(n.overlap))
	for i, pintrvl := range n.overlap {
		interval[i] = *pintrvl
	}
	return interval
}

type Interval struct {
	Id int // unique
	Segment
}

type Segment struct {
	From int
	To   int
}

// Represents overlapping intervals of a segment
type SegmentOverlap struct {
	Segment  Segment
	Interval []Interval
}

// Node receiver for tree traversal
type NodeReceive func(Node)

// NewTree returns a Tree interface with underlying segment tree implementation
func NewTree() Tree {
	t := new(stree)
	t.Clear()
	return t
}

// Push new interval to stack
func (t *stree) Push(from, to int) {
	t.base = append(t.base, Interval{t.count, Segment{from, to}})
	t.count++
}

// Push array of intervals to stack
func (t *stree) PushArray(from, to []int) {
	for i := 0; i < len(from); i++ {
		t.Push(from[i], to[i])
	}
}

// Clear the interval stack
func (t *stree) Clear() {
	t.count = 0
	t.root = nil
	t.base = make([]Interval, 0, 100)
	t.min = 0
	t.max = 0
}

// Build segment tree out of interval stack
func (t *stree) BuildTree() {
	if len(t.base) == 0 {
		panic("No intervals in stack to build tree. Push intervals first")
	}
	var endpoints []int
	endpoints, t.min, t.max = Endpoints(t.base)
	// Create tree nodes from interval endpoints
	t.root = t.insertNodes(elementaryIntervals(endpoints))
	for i := range t.base {
		insertInterval(t.root, &t.base[i])
	}
}

func (t *stree) Print() {
	Print(t.root)
}

func (t *stree) Tree2Array() []SegmentOverlap {
	return Tree2Array(t.root)
}

// Endpoints returns a slice with all endpoints (sorted, unique)
func Endpoints(base []Interval) (result []int, min, max int) {
	baseLen := len(base)
	endpoints := make([]int, baseLen*2)
	for i, interval := range base {
		endpoints[i] = interval.From
		endpoints[i+baseLen] = interval.To
	}
	result = Dedup(endpoints)
	min = result[0]
	max = result[len(result)-1]
	return
}

// Dedup removes duplicates from a given slice
func Dedup(sl []int) []int {
	sort.Sort(sort.IntSlice(sl))
	unique := make([]int, 0, len(sl))
	prev := sl[0] + 1
	for _, val := range sl {
		if val != prev {
			unique = append(unique, val)
			prev = val
		}
	}
	return unique
}

// elementaryIntervals creates a slice of elementary intervals
// from a sorted slice of endpoints
// Input: [p1, p2, ..., pn]
// Output: [{p1 : p2}, {p2 : p2},... , {pn : pn}]
func elementaryIntervals(endpoints []int) []Segment {
	if len(endpoints) == 1 {
		return []Segment{Segment{endpoints[0], endpoints[0]}}
	}

	intervals := make([]Segment, len(endpoints)*2-1)
	for i := 0; i < len(endpoints); i++ {
		intervals[i*2] = Segment{endpoints[i], endpoints[i]}
		if i < len(endpoints)-1 { // don't store {pn, pn+1}
			intervals[i*2+1] = Segment{endpoints[i], endpoints[i+1]}
		}
	}
	return intervals
}

// insertNodes builds the tree structure from the elementary intervals
func (t *stree) insertNodes(leaves []Segment) *node {
	var n *node
	if len(leaves) == 1 {
		n = &node{segment: leaves[0]}
		n.left = nil
		n.right = nil
	} else {
		n = &node{segment: Segment{leaves[0].From, leaves[len(leaves)-1].To}}
		center := len(leaves) / 2
		n.left = t.insertNodes(leaves[:center])
		n.right = t.insertNodes(leaves[center:])
	}

	return n
}

// Disjoint returns true if Segment does not overlap with interval
func (s *Segment) Disjoint(from, to int) bool {
	if from > s.To || to < s.From {
		return true
	}
	return false
}

func (s *Segment) subsetOf(other *Segment) bool {
	return other.From <= s.From && other.To >= s.To
}

func (s *Segment) intersectsWith(other *Segment) bool {
	return other.From <= s.To && s.From <= other.To ||
		s.From <= other.To && other.From <= s.To
}

// Inserts interval into given tree structure
func insertInterval(node *node, intrvl *Interval) {
	if node.segment.subsetOf(&intrvl.Segment) {

		// interval of node is a subset of the specified interval or equal
		if node.overlap == nil {
			node.overlap = make([]*Interval, 0, 10)
		}
		node.overlap = append(node.overlap, intrvl)
	} else {
		if node.left != nil && node.left.segment.intersectsWith(&intrvl.Segment) {
			insertInterval(node.left, intrvl)
		}
		if node.right != nil && node.right.segment.intersectsWith(&intrvl.Segment) {
			insertInterval(node.right, intrvl)
		}
	}
}

// Query interval
func (t *stree) Query(from, to int) []Interval {
	if t.root == nil {
		panic("Can't run query on empty tree. Call BuildTree() first")
	}
	result := make(map[int]Interval)
	querySingle(t.root, from, to, &result)
	// transform map to slice
	sl := make([]Interval, 0, len(result))
	for _, intrvl := range result {
		sl = append(sl, intrvl)
	}
	return sl
}

// querySingle traverse tree in search of overlaps
func querySingle(node *node, from, to int, result *map[int]Interval) {
	if !node.segment.Disjoint(from, to) {
		for _, pintrvl := range node.overlap {
			(*result)[pintrvl.Id] = *pintrvl
		}
		if node.right != nil {
			querySingle(node.right, from, to, result)
		}
		if node.left != nil {
			querySingle(node.left, from, to, result)
		}
	}
}

// Query interval array
func (t *stree) QueryArray(from, to []int) []Interval {
	if t.root == nil {
		panic("Can't run query on empty tree. Call BuildTree() first")
	}
	result := make(map[int]Interval)
	queryMulti(t.root, from, to, &result)
	sl := make([]Interval, 0, len(result))
	for _, intrvl := range result {
		sl = append(sl, intrvl)
	}
	return sl
}

// queryMulti traverse tree in search of overlaps with multiple intervals
func queryMulti(node *node, from, to []int, result *map[int]Interval) {
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
			queryMulti(node.right, hitsFrom, hitsTo, result)
		}
		if node.left != nil {
			queryMulti(node.left, hitsFrom, hitsTo, result)
		}
	}
}

// Traverse tree recursively call enter when entering node, resp. leave
func traverse(node Node, enter, leave NodeReceive) {
	if reflect.ValueOf(node).IsNil() {
		return
	}
	if enter != nil {
		enter(node)
	}
	traverse(node.Right(), enter, leave)
	traverse(node.Left(), enter, leave)
	if leave != nil {
		leave(node)
	}
}

// Print tree recursively to sdout
func Print(root Node) {
	traverse(root, func(node Node) {
		fmt.Printf("\nSegment: (%d,%d)", node.Segment().From, node.Segment().To)
		for _, intrvl := range node.Overlap() {
			fmt.Printf("\nInterval %d: (%d,%d)", intrvl.Id, intrvl.From, intrvl.To)
		}
	}, nil)
}

// Tree2Array transforms tree to array
func Tree2Array(root Node) []SegmentOverlap {
	array := make([]SegmentOverlap, 0, 50)
	traverse(root, func(node Node) {
		seg := SegmentOverlap{Segment: node.Segment(), Interval: node.Overlap()}
		array = append(array, seg)
	}, nil)
	return array
}
