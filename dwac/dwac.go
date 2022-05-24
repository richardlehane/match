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
	"fmt"
	"io"
	"strings"
	"sync"
)

type Dwac struct {
	maxOff  int
	root    *node
	precons *sync.Pool
}

func New(seqs []Seq) *Dwac {
	return nil
}

// Choice represents the different byte slices that can occur at each position of the Seq
type Choice [][]byte

// Seq is an ordered set of slices of Choices, with maximum offsets for each choice
type Seq struct {
	MaxOffsets []int64 // maximum offsets for each choice. Can be -1 for wildcard.
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

// Resume is indexes to wild sequences that should be searched in the dynamic phase
type Resume []Seq

// Result contains the index and offset of matches.
type Result struct {
	Index  [2]int // a double index: index of the Seq and index of the Choice
	Offset int64
	Length int
}

// Dwac returns a channel of results, which are double indexes (of the Seq and of the Choice),
// and a resume channel, which is a list of wild Seq indexes
func (s *Dwac) Index(rdr io.ByteReader) (<-chan Result, chan<- Resume) {
	output, resume := make(chan Result), make(chan Resume)

	// check ac
	// wait on resume, then build wild ac
	// check wild ac
	return output, resume
}
