// Copyright 2014 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package wac is a modified Aho-Corasick multiple string search algorithm.
//
// This algorithm allows for sequences that are composed of sub-sequences
// that can contain an arbitrary number of wildcards. Sequences can also be
// given a maximum offset that defines the maximum byte position of the first sub-sequence.
//
// The results returned are for the matches on subsequences (NOT the full sequences).
// The index of those subsequences and the offset is returned.
// It is up to clients to verify that the complete sequence that they are interested in has matched.
// A "progress" result is sent from offset 1024 onwards. This is to update clients on scanning progress and has index -1,-1.
// This result is sent on powers of two (1024, 2048, 4096, etc.)

// Example usage:
//
//     w := wac.New()
//     seq := wac.Seq(1, []wac.Choice{wac.Choice{[]byte{'b'}}, wac.Choice{[]byte{'a','d'}}, wac.Choice{[]byte{'r', 'a'}}})
//     for result := range w.Index(bytes.NewBuffer([]byte("abracadabra")), make(chan {}struct)) {
// 	     fmt.Println(result.Index, "-", result.Offset)
//     }

package wac

import (
	"fmt"
	"io"
	"strings"
)

type Choice [][]byte

// Seq is an ordered set of slices of byte slices (choices), with maximum offsets for each choice
type Seq struct {
	MaxOffsets []int
	Choices    []Choice
}

func (s Seq) String() string {
	str := "{Offsets:"
	for n, v := range s.MaxOffsets {
		if n > 0 {
			str += ","
		}
		str += fmt.Sprintf(" %d", v)
	}
	str += "; Choices:"
	for n, v := range s.Choices {
		if n > 0 {
			str += ","
		}
		str += " ["
		strs := make([]string, len(v))
		for i := range v {
			strs[i] = string(v[i])
		}
		str += strings.Join(strs, " | ")
		str += "]"
	}
	return str + "}"
}

// New creates an Wild Aho-Corasick tree
func New(seqs []Seq) *Wac {
	wac := new(Wac)
	zero := newNode()
	zero.addGotos(seqs, true)
	root := zero.addFails(true)
	root.addGotos(seqs, false)
	root.addFails(false)
	wac.zero, wac.root = zero, root
	return wac
}

type Wac struct {
	zero *node
	root *node
}

type node struct {
	val     byte
	transit *trans // the goto function
	fail    *node  // the fail function
	output  outs   // the output function
}

func newNode() *node { return &node{transit: newTrans(), output: make(outs, 0, 10)} }

type trans struct {
	keys  []byte
	gotos *[256]*node // the goto function is a pointer to an array of 256 nodes, indexed by the byte val
}

func (t *trans) put(b byte, n *node) {
	t.keys = append(t.keys, b)
	t.gotos[b] = n
}

func (t *trans) get(b byte) (*node, bool) {
	n := t.gotos[b]
	if n == nil {
		return n, false
	}
	return n, true
}

func newTrans() *trans { return &trans{keys: make([]byte, 0, 50), gotos: new([256]*node)} }

type out struct {
	max      int
	seqIndex int
	subIndex int
	length   int
	final    bool
}

type outs []out

func (outs outs) contains(out out) bool {
	for _, o := range outs {
		if o == out {
			return true
		}
	}
	return false
}

func (start *node) addGotos(seqs []Seq, zero bool) {
	// iterate through byte sequences adding goto links to the link matrix
	for id, seq := range seqs {
		for i, choice := range seq.Choices {
			// skip the first choice set if this isn't the zero tree and it is at 0 offset
			if !zero && i == 0 && seq.MaxOffsets[0] == 0 {
				continue
			}
			var f bool
			if i == len(seq.Choices)-1 {
				f = true
			}
			for _, byts := range choice {
				curr := start
				for _, byt := range byts {
					if t, ok := curr.transit.get(byt); ok {
						curr = t
					} else {
						node := newNode()
						node.val = byt
						curr.transit.put(byt, node)
						curr = node
					}
				}
				max := seq.MaxOffsets[i]
				curr.output = append(curr.output, out{max, id, i, len(byts), f})
			}
		}
	}
}

func (start *node) addFails(zero bool) *node {
	// root and its children fail to root
	start.fail = start
	for _, k := range start.transit.keys {
		start.transit.gotos[k].fail = start
	}
	// traverse tree in breadth first search adding fails
	queue := make([]*node, 0, 50)
	queue = append(queue, start)
	for len(queue) > 0 {
		pop := queue[0]
		for _, key := range pop.transit.keys {
			node := pop.transit.gotos[key]
			queue = append(queue, node)
			// starting from the node's parent, follow the fails back towards root,
			// and stop at the first fail that has a goto to the node's value
			fail := pop.fail
			_, ok := fail.transit.get(node.val)
			for fail != start && !ok {
				fail = fail.fail
				_, ok = fail.transit.get(node.val)
			}
			fnode, ok := fail.transit.get(node.val)
			if ok && fnode != node {
				node.fail = fnode
			} else {
				node.fail = start
			}
			// another traverse back to root following the fails. This time add any unique out functions to the node
			fail = node.fail
			for fail != start {
				for _, o := range fail.output {
					if !node.output.contains(o) {
						node.output = append(node.output, o)
					}
				}
				fail = fail.fail
			}
		}
		queue = queue[1:]
	}
	// for the zero tree, rewrite the fail links so they now point to the root of the main tree
	if zero {
		root := newNode()
		start.fail = root
		for _, k := range start.transit.keys {
			start.transit.gotos[k].fail = root
		}
		return root
	}
	return start
}

// Index returns a channel of results, these contain the indexes (in the list of sequences that made the tree)
// and offsets (in the input byte slice) of matching sequences.
// Has a quit channel that should be closed to signal quit.
func (wac *Wac) Index(input io.ByteReader, quit chan struct{}) chan Result {
	output := make(chan Result)
	go wac.match(input, output, quit)
	return output
}

// Result contains the index (in the list of sequences that made the tree) and offset of matches.
type Result struct {
	Index  [2]int
	Offset int
	Length int
	Final  bool
}

var progressResult = Result{Index: [2]int{-1, -1}}

func (wac *Wac) match(input io.ByteReader, results chan Result, quit chan struct{}) {
	var offset int
	precons := make(map[[2]int]int)
	curr := wac.zero
	for {
		select {
		case <-quit:
			close(results)
			return
		default:
		}
		c, err := input.ReadByte()
		if err != nil {
			break
		}
		offset++
		if trans, ok := curr.transit.get(c); ok {
			curr = trans
		} else {
			for curr != wac.root {
				curr = curr.fail
				if trans, ok := curr.transit.get(c); ok {
					curr = trans
					break
				}
			}
		}
		for _, o := range curr.output {
			if o.max == -1 || o.max >= offset-o.length {
				if o.subIndex == 0 || (precons[[2]int{o.seqIndex, o.subIndex - 1}] != 0 && offset-o.length >= precons[[2]int{o.seqIndex, o.subIndex - 1}]) {
					if precons[[2]int{o.seqIndex, o.subIndex}] == 0 {
						precons[[2]int{o.seqIndex, o.subIndex}] = offset
					}
					results <- Result{Index: [2]int{o.seqIndex, o.subIndex}, Offset: offset - o.length, Length: o.length, Final: o.final}
				}
			}
		}
		if offset >= 1024 && offset&(^offset+1) == offset {
			progressResult.Offset = offset
			results <- progressResult
		}
	}
	close(results)
}
