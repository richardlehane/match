package wac

import (
	"bytes"
	"testing"
)

func equal(a []Result, b []Result) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v.Index != b[i].Index {
			return false
		}
	}
	return true
}

func loop(output chan Result) []Result {
	results := make([]Result, 0)
	for res := range output {
		results = append(results, res)
	}
	return results
}

func test(t *testing.T, a []byte, b []Seq, expect []Result) {
	wac := New(b)
	output := wac.Index(bytes.NewBuffer(a), make(chan struct{}))
	results := loop(output)
	if !equal(expect, results) {
		t.Errorf("Index fail; Expecting: %v, Got: %v", expect, results)
	}
}

func seq(max int, s string) Seq {
	return NewSeq(max, []byte(s))
}

// Tests (the test strings are taken from John Graham-Cumming's lua implementation: https://github.com/jgrahamc/aho-corasick-lua Copyright (c) 2013 CloudFlare)
func TestWikipedia(t *testing.T) {
	test(t, []byte("abccab"),
		[]Seq{seq(-1, "a"), seq(-1, "ab"), seq(-1, "bc"), seq(-1, "bca"), seq(-1, "c"), seq(-1, "caa")},
		[]Result{Result{[2]int{0, 0}, 0}, Result{[2]int{1, 0}, 0}, Result{[2]int{2, 0}, 1}, Result{[2]int{4, 0}, 2}, Result{[2]int{4, 0}, 3}, Result{[2]int{0, 0}, 4}, Result{[2]int{1, 0}, 4}})
}

func TestSimple(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "poto")},
		[]Result{})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "The")},
		[]Result{Result{[2]int{0, 0}, 0}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "pot")},
		[]Result{Result{[2]int{0, 0}, 4}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "pot ")},
		[]Result{Result{[2]int{0, 0}, 4}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "ot h")},
		[]Result{Result{[2]int{0, 0}, 5}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "andle")},
		[]Result{Result{[2]int{0, 0}, 15}})
}

func TestMultipleNonoverlapping(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "h")},
		[]Result{Result{[2]int{0, 0}, 1}, Result{[2]int{0, 0}, 8}, Result{[2]int{0, 0}, 14}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "ha"), seq(-1, "he")},
		[]Result{Result{[2]int{1, 0}, 1}, Result{[2]int{0, 0}, 8}, Result{[2]int{0, 0}, 14}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "pot"), seq(-1, "had")},
		[]Result{Result{[2]int{0, 0}, 4}, Result{[2]int{1, 0}, 8}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "pot"), seq(-1, "had"), seq(-1, "hod")},
		[]Result{Result{[2]int{0, 0}, 4}, Result{[2]int{1, 0}, 8}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "The"), seq(-1, "pot"), seq(-1, "had"), seq(-1, "hod"), seq(-1, "andle")},
		[]Result{Result{[2]int{0, 0}, 0}, Result{[2]int{1, 0}, 4}, Result{[2]int{2, 0}, 8}, Result{[2]int{4, 0}, 15}})
}

func TestOverlapping(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "Th"), seq(-1, "he pot"), seq(-1, "The"), seq(-1, "pot h")},
		[]Result{Result{[2]int{0, 0}, 0}, Result{[2]int{2, 0}, 0}, Result{[2]int{1, 0}, 1}, Result{[2]int{3, 0}, 4}})
}

func TestNesting(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "handle"), seq(-1, "hand"), seq(-1, "and"), seq(-1, "andle")},
		[]Result{Result{[2]int{1, 0}, 14}, Result{[2]int{2, 0}, 15}, Result{[2]int{0, 0}, 14}, Result{[2]int{3, 0}, 15}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "handle"), seq(-1, "hand"), seq(-1, "an"), seq(-1, "n")},
		[]Result{Result{[2]int{2, 0}, 15}, Result{[2]int{3, 0}, 16}, Result{[2]int{1, 0}, 14}, Result{[2]int{0, 0}, 14}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "dle"), seq(-1, "l"), seq(-1, "le")},
		[]Result{Result{[2]int{1, 0}, 18}, Result{[2]int{0, 0}, 17}, Result{[2]int{2, 0}, 18}})
}

func TestRandom(t *testing.T) {
	test(t, []byte("yasherhs"),
		[]Seq{seq(-1, "say"), seq(-1, "she"), seq(-1, "shr"), seq(-1, "he"), seq(-1, "her")},
		[]Result{Result{[2]int{1, 0}, 2}, Result{[2]int{3, 0}, 3}, Result{[2]int{4, 0}, 3}})
}

func TestFailPartial(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "dlf"), seq(-1, "l")},
		[]Result{Result{[2]int{1, 0}, 18}})
}

func TestMany(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "handle"), seq(-1, "andle"), seq(-1, "ndle"), seq(-1, "dle"), seq(-1, "le"), seq(-1, "e")},
		[]Result{Result{[2]int{5, 0}, 2}, Result{[2]int{0, 0}, 14}, Result{[2]int{1, 0}, 15}, Result{[2]int{2, 0}, 16}, Result{[2]int{3, 0}, 17}, Result{[2]int{4, 0}, 18}, Result{[2]int{5, 0}, 19}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "handle"), seq(-1, "handl"), seq(-1, "hand"), seq(-1, "han"), seq(-1, "ha"), seq(-1, "a")},
		[]Result{Result{[2]int{4, 0}, 8}, Result{[2]int{5, 0}, 9}, Result{[2]int{5, 0}, 12}, Result{[2]int{4, 0}, 14}, Result{[2]int{5, 0}, 15}, Result{[2]int{3, 0}, 14}, Result{[2]int{2, 0}, 14}, Result{[2]int{1, 0}, 14}, Result{[2]int{0, 0}, 14}})
}

func TestLong(t *testing.T) {
	test(t, []byte("macintosh"),
		[]Seq{seq(-1, "acintosh"), seq(-1, "in")},
		[]Result{Result{[2]int{1, 0}, 3}, Result{[2]int{0, 0}, 1}})
	test(t, []byte("macintosh"),
		[]Seq{seq(-1, "acintosh"), seq(-1, "in"), seq(-1, "tosh")},
		[]Result{Result{[2]int{1, 0}, 3}, Result{[2]int{0, 0}, 1}, Result{[2]int{2, 0}, 5}})
	test(t, []byte("macintosh"),
		[]Seq{seq(-1, "acintosh"), seq(-1, "into"), seq(-1, "to"), seq(-1, "in")},
		[]Result{Result{[2]int{3, 0}, 3}, Result{[2]int{1, 0}, 3}, Result{[2]int{2, 0}, 5}, Result{[2]int{0, 0}, 1}})
}

// Benchmarks
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New([]Seq{seq(-1, "handle"), seq(-1, "handl"), seq(-1, "hand"), seq(-1, "han"), seq(-1, "ha"), seq(-1, "a")})
	}
}

func BenchmarkIndex(b *testing.B) {
	b.StopTimer()
	ac := New([]Seq{seq(-1, "handle"), seq(-1, "handl"), seq(-1, "hand"), seq(-1, "han"), seq(-1, "ha"), seq(-1, "a")})
	input := bytes.NewBuffer([]byte("The pot had a handle"))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for _ = range ac.Index(input, make(chan struct{})) {
		}
	}
}

// following benchmark code is from <http://godoc.org/code.google.com/p/ahocorasick> for comparison
func benchmarkValue(n int) []byte {
	input := []byte{}
	for i := 0; i < n; i++ {
		var b byte
		if i%2 == 0 {
			b = 'a'
		} else {
			b = 'b'
		}
		input = append(input, b)
	}
	return input
}

func hardTree() []Seq {
	ret := make([]Seq, 0, 2500)
	str := ""
	for i := 0; i < 2500; i++ {
		// We add a 'q' to the end to make sure we never actually match
		ret = append(ret, seq(-1, str+string('a'+(i%26))+"q"))
		if i%26 == 25 {
			str = str + string('a'+len(str)%2)
		}
	}
	return ret
}

func BenchmarkMatchingNoMatch(b *testing.B) {
	b.StopTimer()
	reader := bytes.NewBuffer(benchmarkValue(b.N))
	ac := New([]Seq{seq(-1, "abababababababd"),
		seq(-1, "abababb"),
		seq(-1, "abababababq")})
	b.StartTimer()
	for _ = range ac.Index(reader, make(chan struct{})) {
	}
}

func BenchmarkMatchingManyMatches(b *testing.B) {
	b.StopTimer()
	reader := bytes.NewBuffer(benchmarkValue(b.N))
	ac := New([]Seq{seq(-1, "ab"),
		seq(-1, "ababababababab"),
		seq(-1, "ababab"),
		seq(-1, "ababababab")})
	b.StartTimer()
	for _ = range ac.Index(reader, make(chan struct{})) {
	}
}

func BenchmarkMatchingHardTree(b *testing.B) {
	b.StopTimer()
	reader := bytes.NewBuffer(benchmarkValue(b.N))
	ac := New(hardTree())
	b.StartTimer()
	for _ = range ac.Index(reader, make(chan struct{})) {
	}
}
