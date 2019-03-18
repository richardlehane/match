// Copyright 2019 Richard Lehane. All rights reserved.
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

package dwac

import (
	"io"
	"sync"
)

// Entanglement is an (OR) set of an (AND) set of bof/eof seqs that must be satisfied
// up to the first wild in that seq
// you only entangle with seqs that have fixed max offsets
// [2]int, [2]int
// [2]int
// [2]int, [2]int, [2]int
type Entanglement [][][2]int

// Result contains the index and offset of matches.
type Result struct {
	Index  [2]int // a double index: index of the Seq and index of the Choice
	Offset int64
	Length int
}

type Searcher struct {
	bofOnce       *sync.Once
	bofWac        Wac
	eofOnce       *sync.Once
	eofWac        Wac
	maxBof        int
	maxEof        int
	wildSeqs      []Seq          // separate out wild sequences to create a dynamic searcher for wildcard matching
	entanglements []Entanglement // same len as wildSeqs
}

func New(bofSeqs []Seq, eofSeqs []Seq, entanglements map[int]Entanglement) *Searcher {
	return nil
}

// Search returns a channel of results, these contain the indexes (a double index: index of the Seq and index of the Choice)
// and offsets (in the input byte slice) of matching sequences.
func (s *Searcher) Search(bof, eof io.ByteReader) chan Result {
	output := make(chan Result)
	// check bof
	// check eof
	// build wild matcher
	// check wild bof
	return output
}
