// Copyright 2012 Thomas Oberndörfer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stree

// serial is a structure that allows to query intervals
// with a sequential algorithm
type serial struct {
	stree
}

// NewSerial returns a Tree interface with underlying serial algorithm
func NewSerial() Tree {
	t := new(serial)
	t.Clear()
	return t
}

func (t *serial) BuildTree() {
	panic("BuildTree() not supported for serial data structure")
}

func (t *serial) Print() {
	panic("Print() not supported for serial data structure")
}

func (t *serial) Tree2Array() []SegmentOverlap {
	panic("Tree2Array() not supported for serial data structure")
}

// Query interval by looping through the interval stack
func (t *serial) Query(from, to int) []Interval {
	result := make([]Interval, 0, 10)
	for _, intrvl := range t.base {
		if !intrvl.Segment.Disjoint(from, to) {
			result = append(result, intrvl)
		}
	}
	return result
}

// Query interval array by looping through the interval stack
func (t *serial) QueryArray(from, to []int) []Interval {
	result := make([]Interval, 0, 10)
	for i, fromvalue := range from {
		result = append(result, t.Query(fromvalue, to[i])...)
	}
	return result
}
