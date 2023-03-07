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

	"github.com/richardlehane/match/fwac"
)

type Dwac struct {
	maxOff int64
	root   *node
	p      *sync.Pool
}

func New(seqs []fwac.Seq) *Dwac {
	d := &Dwac{}
	d.root = &node{}
	d.maxOff = d.root.addGotos(seqs)
	d.root.addFails()
	d.p = &sync.Pool{New: preconsFn(seqs)}
	return d
}

// Dwac returns a channel of results, which are double indexes (of the Seq and of the Choice),
// and a resume channel, which is a slice of wild Seq indexes
func (d *Dwac) Index(rdr io.ByteReader) (<-chan fwac.Result, chan<- []fwac.Seq) {
	output, resume := make(chan fwac.Result), make(chan []fwac.Seq)
	go d.match(rdr, output, resume)
	return output, resume
}

func (dwac *Dwac) match(input io.ByteReader, results chan fwac.Result, resume chan []fwac.Seq) {
	var offset int64
	var resumeSignal = fwac.Result{Index: [2]int{-1, -1}}
	p := dwac.p.Get().(precons)
	curr := dwac.root
	var c byte
	var err error
	for c, err = input.ReadByte(); err == nil; c, err = input.ReadByte() {
		offset++
		if trans := curr.transit[c]; trans != nil {
			curr = trans
		} else {
			for curr != dwac.root {
				curr = curr.fail
				if trans := curr.transit[c]; trans != nil {
					curr = trans
					break
				}
			}
		}
		if curr.output != nil && (curr.outMax == -1 || curr.outMax >= offset-int64(curr.outMaxL)) {
			for _, o := range curr.output {
				if o.max == -1 || o.max >= offset-int64(o.length) {
					if o.subIndex == 0 || (p[o.seqIndex][o.subIndex-1] != 0 && offset-int64(o.length) >= p[o.seqIndex][o.subIndex-1]) {
						if p[o.seqIndex][o.subIndex] == 0 {
							p[o.seqIndex][o.subIndex] = offset
						}
						results <- fwac.Result{Index: [2]int{o.seqIndex, o.subIndex}, Offset: offset - int64(o.length), Length: o.length}
					}
				}
			}
		}
		if offset > int64(dwac.maxOff) && curr == dwac.root {
			break
		}
	}
	// return precons
	dwac.p.Put(clear(p))
	// if EOF not reached or other file read error, try the resume channel
	if err == nil {
		results <- resumeSignal
		seqs := <-resume
		if len(seqs) > 0 {
			root := &node{}
			root.addGotos(seqs)
			root.addFails()
			curr = root
			p = newPrecons(makeT(seqs))
			for c, err = input.ReadByte(); err == nil; c, err = input.ReadByte() {
				offset++
				if trans := curr.transit[c]; trans != nil {
					curr = trans
				} else {
					for curr != root {
						curr = curr.fail
						if trans := curr.transit[c]; trans != nil {
							curr = trans
							break
						}
					}
				}
				if curr.output != nil && (curr.outMax == -1 || curr.outMax >= offset-int64(curr.outMaxL)) {
					for _, o := range curr.output {
						if o.max == -1 || o.max >= offset-int64(o.length) {
							if o.subIndex == 0 || (p[o.seqIndex][o.subIndex-1] != 0 && offset-int64(o.length) >= p[o.seqIndex][o.subIndex-1]) {
								if p[o.seqIndex][o.subIndex] == 0 {
									p[o.seqIndex][o.subIndex] = offset
								}
								results <- fwac.Result{Index: [2]int{o.seqIndex, o.subIndex}, Offset: offset - int64(o.length), Length: o.length}
							}
						}
					}
				}
			}
		}
	}
	close(results)
}
