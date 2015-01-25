package wac

import "sort"

type transitionFunc func() transition

type transition interface {
	get(byte) (*node, bool)
	put(byte, transitionFunc) *node
	iter(int) *node
	finalise()
}

// The default transition type is a simple 256 array of pointers. This is fast but consumes a lot of memory.

func newTrans() transition { return &trans{keys: make([]byte, 0, 1), gotos: new([256]*node)} }

type trans struct {
	keys  []byte
	gotos *[256]*node // the goto function is a pointer to an array of 256 nodes, indexed by the byte val
}

func (t *trans) put(b byte, fn transitionFunc) *node {
	if t.gotos[b] == nil {
		n := newNode(fn)
		n.val = b
		t.keys = append(t.keys, b)
		t.gotos[b] = n
		return n
	}
	return t.gotos[b]
}

func (t *trans) get(b byte) (*node, bool) {
	n := t.gotos[b]
	if n == nil {
		return n, false
	}
	return n, true
}

func (t *trans) finalise() {}

func (t *trans) iter(n int) *node {
	if n >= len(t.keys) {
		return nil
	}
	return t.gotos[t.keys[n]]
}

// The low memory transition uses a slice of nodes with binary search. It is modelled on: https://code.google.com/p/ahocorasick/source/browse/aho.go

func newTransLM() transition { return &transLM{} }

type link struct {
	b byte
	n *node
}

type transLM []*link

func (t transLM) Len() int {
	return len(t)
}
func (t transLM) Less(i, j int) bool {
	return t[i].b < t[j].b
}
func (t transLM) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t *transLM) put(b byte, fn transitionFunc) *node {
	for _, l := range *t {
		if l.b == b {
			return l.n
		}
	}
	n := newNode(fn)
	n.val = b
	*t = append(*t, &link{b, n})
	return n
}

func (t *transLM) get(b byte) (*node, bool) {
	top, bottom := len(*t), 0
	for top > bottom {
		i := (top-bottom)/2 + bottom
		b2 := (*t)[i].b
		if b2 > b {
			top = i
		} else if b2 < b {
			bottom = i + 1
		} else {
			return (*t)[i].n, true
		}
	}
	return nil, false
}

func (t *transLM) finalise() {
	sort.Sort(*t)
}

func (t *transLM) iter(n int) *node {
	if n >= len(*t) {
		return nil
	}
	return (*t)[n].n
}
