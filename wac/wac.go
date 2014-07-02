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

// Example usage:
//
//     w := wac.New()
//     seq := wac.NewSeq(1, []byte{'b'}, []byte{'a','d'}}, []byte{'r', 'a'})
//     w.Add(seq)
//     for result := range w.Index(bytes.NewBuffer([]byte("abracadabra"))) {
// 	   fmt.Println(result.Index, "-", result.Offset)
//     }

package wac

// Todo... :
// a save-able/load-able design; use less memory by only storing relevant trans nodes in a sorted slice.
// New plan: single AC. But with a modified output function that checks for preconditions before sending.
// Preconditions are max offset for head; and entanglements for tail. Store entanglements in a bool slice. With a reset slice that just contains indexes of entanglements.
// Trans function takes offset, preceding-matches, and the byte. When building the WAC if we have multiple choices from the same byte, we go with the simplest
// Output function takes preceding matches and the results channel.

import "io"

type Wac struct {
	Preconditions []bool
}

type Seq struct {
	Head Head
	Tail [][]byte
}

type Head struct {
	max int
	val []byte
}

type AC []Node

type Node struct {
	Val   byte
	Trans Transition // the goto function
	Fail  [2]int     // the fail function
	Out   Out        // the output function
}

type Trans struct {
	Val byte
	Idx [2]int // the goto function is a pointer to an array of 256 nodes, indexed by the byte val
}

type Transition func(byte) (*Node, bool)



type Out []struct{[4]int // precondition index, node indexes (sequence and subsequence), and len}

func (t *trans) put(b byte, ac *Ac) {
	t.keys = append(t.keys, b)
	t.gotos[b] = ac
}

func (t *trans) get(b byte) (*Ac, bool) {
	node := t.gotos[b]
	if node == nil {
		return node, false
	}
	return node, true
}

func newTrans() *trans { return &trans{keys: make([]byte, 0, 50), gotos: new([256]*Ac)} }

func (o out) contains(i int) bool {
	for _, v := range o {
		if v[0] == i {
			return true
		}
	}
	return false
}

func newNode() *Ac { return &Ac{trans: newTrans(), out: make(out, 0, 10)} }

func NewSeq(max int, byts ...[][]byte) Seq {

}

// New creates an Aho-Corasick tree from a slice of byte slices
func New(seqs []Seq) *Ac {
	root := newNode()
	root.addGotos(seqs, false)
	root.addFails()
	return root
}

func (root *Ac) addGotos(seqs [][]byte, fixed bool) {
	// iterate through byte sequences adding goto links to the link matrix
	for id, seq := range seqs {
		curr := root
		for _, seqByte := range seq {
			if trans, ok := curr.trans.get(seqByte); ok {
				curr = trans
			} else {
				node := newNode()
				node.val = seqByte
				if fixed {
					node.fail = root
				}
				curr.trans.put(seqByte, node)
				curr = node
			}
		}
		curr.out = append(curr.out, [2]int{id, len(seq)})
	}
}

func (root *Ac) addFails() {
	// root and its children fail to root
	root.fail = root
	for _, k := range root.trans.keys {
		root.trans.gotos[k].fail = root
	}
	// traverse tree in breadth first search adding fails
	queue := make([]*Ac, 0, 50)
	queue = append(queue, root)
	for len(queue) > 0 {
		pop := queue[0]
		for _, key := range pop.trans.keys {
			node := pop.trans.gotos[key]
			queue = append(queue, node)
			// starting from the node's parent, follow the fails back towards root,
			// and stop at the first fail that has a goto to the node's value
			fail := pop.fail
			_, ok := fail.trans.get(node.val)
			for fail != root && !ok {
				fail = fail.fail
				_, ok = fail.trans.get(node.val)
			}
			fnode, ok := fail.trans.get(node.val)
			if ok && fnode != node {
				node.fail = fnode
			} else {
				node.fail = root
			}
			// another traverse back to root following the fails. This time add any unique out functions to the node
			fail = node.fail
			for fail != root {
				for _, id := range fail.out {
					if !node.out.contains(id[0]) {
						node.out = append(node.out, id)
					}
				}
				fail = fail.fail
			}
		}
		queue = queue[1:]
	}
}

// Index returns a channel of results, these contain the indexes (in the list of sequences that made the tree)
// and offsets (in the input byte slice) of matching sequences.
// Has a quit channel that should be closed to signal quit.
func (ac *Ac) Index(input io.ByteReader, quit chan struct{}) chan Result {
	output := make(chan Result, 20)
	go ac.match(input, output, quit)
	return output
}

// Result contains the index (in the list of sequences that made the tree) and offset of matches.
type Result struct {
	Index  [2]int
	Offset int
}

func (ac *Ac) match(input io.ByteReader, results chan Result, quit chan struct{}) {
	var offset int
	root := ac
	curr := root

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
		if trans, ok := curr.trans.get(c); ok {
			curr = trans
		} else {
			for curr != root {
				curr = curr.fail
				if trans, ok := curr.trans.get(c); ok {
					curr = trans
					break
				}
			}
		}
		for _, id := range curr.out {
			results <- Result{Index: id[0], Offset: offset - id[1]}
		}
	}
	close(results)
}
