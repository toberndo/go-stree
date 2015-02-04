// Copyright 2012 Thomas Oberndörfer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package stree

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
)

func TestTreeEqualSerial(t *testing.T) {
	tree := NewTree()
	serial := NewSerial()
	for i := 0; i < 100000; i++ {
		min := rand.Int()
		max := rand.Int()
		if min > max {
			min, max = max, min
		}
		tree.Push(min, max)
		serial.Push(min, max)
	}
	tree.BuildTree()
	treeresult := tree.Query(0, 1000000)
	//fmt.Println(treeresult)
	serialresult := serial.Query(0, 1000000)
	//fmt.Println(serialresult)
	treemap := make(map[int]Segment)
	fail := false
	if len(treeresult) != len(serialresult) {
		fail = true
		fmt.Printf("unequal result length")
		goto Fail
	}
	for _, value := range treeresult {
		treemap[value.Id] = value.Segment
	}
	for _, intrvl := range serialresult {
		if treemap[intrvl.Id] != intrvl.Segment {
			fail = true
			fmt.Printf("result interval mismatch")
			break
		}
	}
Fail:
	if fail {
		t.Errorf("Result not equal")
	}
}

func TestMinimalTree(t *testing.T) {
	tree := NewTree()
	tree.Push(3, 7)
	tree.BuildTree()
	fail := false
	result := tree.Query(1, 2)
	if len(result) != 0 {
		fail = true
	}
	result = tree.Query(2, 3)
	if len(result) != 1 {
		fail = true
	}
	if fail {
		t.Errorf("fail query minimal tree")
	}
}

func TestMinimalTree2(t *testing.T) {
	tree := NewTree()
	tree.Push(1, 1)
	tree.BuildTree()
	if result := tree.Query(1, 1); len(result) != 1 {
		t.Errorf("fail query minimal tree for (1, 1)")
	}
	if result := tree.Query(1, 2); len(result) != 1 {
		t.Errorf("fail query minimal tree for (1, 2)")
	}
	if result := tree.Query(2, 3); len(result) != 0 {
		t.Errorf("fail query minimal tree for (2, 3)")
	}
}

func TestNormalTree(t *testing.T) {
	tree := NewTree()
	tree.Push(1, 1)
	tree.Push(2, 3)
	tree.Push(5, 7)
	tree.Push(4, 6)
	tree.Push(6, 9)
	tree.BuildTree()
	if result := tree.Query(3, 5); len(result) != 3 {
		t.Errorf("fail query multiple tree for (3, 5)")
	}
	qvalid := map[int]int{
		0: 0,
		1: 1,
		2: 1,
		3: 1,
		4: 1,
		5: 2,
		6: 3,
		7: 2,
		8: 1,
		9: 1,
	}
	for i := 0; i <= 9; i++ {
		if result := tree.Query(i, i); len(result) != qvalid[i] {
			t.Errorf("fail query multiple tree for (%d, %d)", i, i)
		}
	}
}

func BenchmarkSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree := NewTree()
		tree.Push(1, 1)
		tree.Push(2, 3)
		tree.Push(5, 7)
		tree.Push(4, 6)
		tree.Push(6, 9)
		tree.Push(9, 14)
		tree.Push(10, 13)
		tree.Push(11, 11)
		tree.BuildTree()
	}
}

var tree Tree
var ser Tree

func init() {
	tree = NewTree()
	ser = NewSerial()
	for j := 0; j < 100000; j++ {
		min := rand.Int()
		max := rand.Int()
		if min > max {
			min, max = max, min
		}
		tree.Push(min, max)
		ser.Push(min, max)
	}
	tree.BuildTree()
}

func BenchmarkBuildTree1000(b *testing.B) {
	tree := NewTree()
	buildTree(b, tree, 1000)
}

func BenchmarkBuildTree100000(b *testing.B) {
	tree := NewTree()
	buildTree(b, tree, 100000)
}

func BenchmarkQueryTree(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree.Query(0, 100000)
	}
}

func BenchmarkQuerySerial(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ser.Query(0, 100000)
	}
}

func BenchmarkQueryTreeMax(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree.Query(0, math.MaxInt32)
	}
}

func BenchmarkQuerySerialMax(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ser.Query(0, math.MaxInt32)
	}
}

func BenchmarkQueryTreeArray(b *testing.B) {
	from := []int{0, 1000000, 2000000, 3000000, 4000000, 5000000, 6000000, 7000000, 8000000, 9000000}
	to := []int{10, 1000010, 2000010, 3000010, 4000010, 5000010, 6000010, 7000010, 8000010, 9000010}
	for i := 0; i < b.N; i++ {
		tree.QueryArray(from, to)
	}
}

func BenchmarkQuerySerialArray(b *testing.B) {
	from := []int{0, 1000000, 2000000, 3000000, 4000000, 5000000, 6000000, 7000000, 8000000, 9000000}
	to := []int{10, 1000010, 2000010, 3000010, 4000010, 5000010, 6000010, 7000010, 8000010, 9000010}
	for i := 0; i < b.N; i++ {
		ser.QueryArray(from, to)
	}
}

func buildTree(b *testing.B, tree Tree, count int) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tree.Clear()
		pushRandom(tree, count)
		b.StartTimer()
		tree.BuildTree()
	}
}

func pushRandom(tree Tree, count int) {
	for j := 0; j < count; j++ {
		min := rand.Int()
		max := rand.Int()
		if min > max {
			min, max = max, min
		}
		tree.Push(min, max)
	}
}

func BenchmarkEndpoints100000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tree := NewTree().(*stree)
		pushRandom(tree, 100000)
		b.StartTimer()
		Endpoints(tree.base)
	}
}

func BenchmarkInsertNodes100000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tree := NewTree().(*stree)
		pushRandom(tree, 100000)
		var endpoint []int
		endpoint, tree.min, tree.max = Endpoints(tree.base)
		//fmt.Println(len(endpoint))
		b.StartTimer()
		tree.root = tree.insertNodes(endpoint)
	}
}

func BenchmarkInsertIntervals100000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tree := NewTree().(*stree)
		pushRandom(tree, 100000)
		var endpoint []int
		endpoint, tree.min, tree.max = Endpoints(tree.base)
		tree.root = tree.insertNodes(endpoint)
		b.StartTimer()
		for i := range tree.base {
			insertInterval(tree.root, &tree.base[i])
		}
	}
}
