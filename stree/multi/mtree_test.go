// Copyright 2012 Thomas Oberndörfer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multi

import (
	"fmt"
	. "github.com/toberndo/go-stree/stree"
	"math"
	"math/rand"
	"testing"
)

func TestTreeEqualMTree(t *testing.T) {
	tree := NewTree()
	mtree := NewMTree()
	for i := 0; i < 1000; i++ {
		min := rand.Int()
		max := rand.Int()
		if min > max {
			min, max = max, min
		}
		tree.Push(min, max)
		mtree.Push(min, max)
	}
	tree.BuildTree()
	mtree.BuildTree()
	tSegArray := tree.Tree2Array()
	mSegArray := mtree.Tree2Array()
	fail := false
	if len(tSegArray) != len(mSegArray) {
		fail = true
		fmt.Printf("unequal array length")
		goto Fail
	}
outer:
	for i, seg := range tSegArray {
		mseg := mSegArray[i]
		if seg.Segment != mseg.Segment {
			fail = true
			fmt.Printf("unequal segment")
			break
		}

		if len(seg.Interval) != len(mseg.Interval) {
			fail = true
			fmt.Printf("unequal interval len\n")
			fmt.Println(len(seg.Interval))
			fmt.Println(len(mseg.Interval))
			break
		}
		for _, intrvl := range seg.Interval {
			count := 0
			for _, mintrvl := range mseg.Interval {
				if intrvl == mintrvl {
					count++
				}
			}
			if count != 1 {
				fail = true
				fmt.Printf("unequal count\n")
				break outer
			}
		}
	}
Fail:
	if fail {
		t.Errorf("Trees not equal")
	}

}

func TestMinimalTree(t *testing.T) {
	tree := NewMTree()
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
		tree := NewMTree()
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

func BenchmarkBuildTree100000(b *testing.B) {
	tree := NewTree()
	buildTree(b, tree, 100000)
}

func BenchmarkBuildMulti100000(b *testing.B) {
	tree := NewMTree()
	buildTree(b, tree, 100000)
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

func BenchmarkInsertNodesMulti100000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tree := NewMTree().(*mtree)
		pushRandom(tree, 100000)
		var endpoint []int
		endpoint, tree.min, tree.max = Endpoints(tree.base)
		b.StartTimer()
		tree.root = tree.insertNodes(endpoint, 0)
		for i := 0; i < tree.numG; i++ {
			<-tree.done
		}
	}
}

func BenchmarkInsertIntervalsMulti100000(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		tree := NewMTree().(*mtree)
		pushRandom(tree, 100000)
		var endpoint []int
		endpoint, tree.min, tree.max = Endpoints(tree.base)
		tree.root = tree.insertNodes(endpoint, 0)
		for i := 0; i < tree.numG; i++ {
			<-tree.done
		}
		b.StartTimer()
		tree.insertIntervalM()
	}
}

var tree Tree
var multi Tree

func init() {
	tree = NewTree()
	multi = NewMTree()
	for j := 0; j < 100000; j++ {
		min := rand.Int()
		max := rand.Int()
		if min > max {
			min, max = max, min
		}
		tree.Push(min, max)
		multi.Push(min, max)
	}
	tree.BuildTree()
	multi.BuildTree()
}

func BenchmarkQueryTree(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree.Query(0, 100000)
	}
}

func BenchmarkQueryMulti(b *testing.B) {
	for i := 0; i < b.N; i++ {
		multi.Query(0, 100000)
	}
}

func BenchmarkQueryTreeMax(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree.Query(0, math.MaxInt32)
	}
}

func BenchmarkQueryMulitMax(b *testing.B) {
	for i := 0; i < b.N; i++ {
		multi.Query(0, math.MaxInt32)
	}
}

func BenchmarkQueryArray(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree.QueryArray([]int{0, 100000000, 200000000, 300000000, 400000000, 500000000, 600000000, 700000000, 800000000, 900000000},
			[]int{50000000, 150000000, 250000000, 350000000, 450000000, 550000000, 650000000, 750000000, 850000000, 950000000})
	}
}

func BenchmarkQueryArrayMulti(b *testing.B) {
	for i := 0; i < b.N; i++ {
		multi.QueryArray([]int{0, 100000000, 200000000, 300000000, 400000000, 500000000, 600000000, 700000000, 800000000, 900000000},
			[]int{50000000, 150000000, 250000000, 350000000, 450000000, 550000000, 650000000, 750000000, 850000000, 950000000})
	}
}
