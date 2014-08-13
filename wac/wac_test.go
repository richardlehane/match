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
		if v != b[i] {
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
	output := wac.Index(bytes.NewBuffer(a), make(chan int), make(chan struct{}))
	results := loop(output)
	if !equal(expect, results) {
		t.Errorf("Index fail; Expecting: %v, Got: %v", expect, results)
	}
}

func seq(max int, s string) Seq {
	return Seq{max, []Choice{Choice{[]byte(s)}}}
}

// Tests (the test strings are taken from John Graham-Cumming's lua implementation: https://github.com/jgrahamc/aho-corasick-lua Copyright (c) 2013 CloudFlare)
func TestWikipedia(t *testing.T) {
	test(t, []byte("abccab"),
		[]Seq{seq(-1, "a"), seq(-1, "ab"), seq(-1, "bc"), seq(-1, "bca"), seq(-1, "c"), seq(-1, "caa")},
		[]Result{Result{[2]int{0, 0}, 0, 1, true}, Result{[2]int{1, 0}, 0, 2, true}, Result{[2]int{2, 0}, 1, 2, true}, Result{[2]int{4, 0}, 2, 1, true}, Result{[2]int{4, 0}, 3, 1, true}, Result{[2]int{0, 0}, 4, 1, true}, Result{[2]int{1, 0}, 4, 2, true}})
}

func TestSimple(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "poto")},
		[]Result{})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "The")},
		[]Result{Result{[2]int{0, 0}, 0, 3, true}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "pot")},
		[]Result{Result{[2]int{0, 0}, 4, 3, true}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "pot ")},
		[]Result{Result{[2]int{0, 0}, 4, 4, true}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "ot h")},
		[]Result{Result{[2]int{0, 0}, 5, 4, true}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "andle")},
		[]Result{Result{[2]int{0, 0}, 15, 5, true}})
}

func TestMultipleNonoverlapping(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "h")},
		[]Result{Result{[2]int{0, 0}, 1, 1, true}, Result{[2]int{0, 0}, 8, 1, true}, Result{[2]int{0, 0}, 14, 1, true}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "ha"), seq(-1, "he")},
		[]Result{Result{[2]int{1, 0}, 1, 2, true}, Result{[2]int{0, 0}, 8, 2, true}, Result{[2]int{0, 0}, 14, 2, true}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "pot"), seq(-1, "had")},
		[]Result{Result{[2]int{0, 0}, 4, 3, true}, Result{[2]int{1, 0}, 8, 3, true}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "pot"), seq(-1, "had"), seq(-1, "hod")},
		[]Result{Result{[2]int{0, 0}, 4, 3, true}, Result{[2]int{1, 0}, 8, 3, true}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "The"), seq(-1, "pot"), seq(-1, "had"), seq(-1, "hod"), seq(-1, "andle")},
		[]Result{Result{[2]int{0, 0}, 0, 3, true}, Result{[2]int{1, 0}, 4, 3, true}, Result{[2]int{2, 0}, 8, 3, true}, Result{[2]int{4, 0}, 15, 5, true}})
}

func TestOverlapping(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "Th"), seq(-1, "he pot"), seq(-1, "The"), seq(-1, "pot h")},
		[]Result{Result{[2]int{0, 0}, 0, 2, true}, Result{[2]int{2, 0}, 0, 3, true}, Result{[2]int{1, 0}, 1, 6, true}, Result{[2]int{3, 0}, 4, 5, true}})
}

func TestNesting(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "handle"), seq(-1, "hand"), seq(-1, "and"), seq(-1, "andle")},
		[]Result{Result{[2]int{1, 0}, 14, 4, true}, Result{[2]int{2, 0}, 15, 3, true}, Result{[2]int{0, 0}, 14, 6, true}, Result{[2]int{3, 0}, 15, 5, true}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "handle"), seq(-1, "hand"), seq(-1, "an"), seq(-1, "n")},
		[]Result{Result{[2]int{2, 0}, 15, 2, true}, Result{[2]int{3, 0}, 16, 1, true}, Result{[2]int{1, 0}, 14, 4, true}, Result{[2]int{0, 0}, 14, 6, true}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "dle"), seq(-1, "l"), seq(-1, "le")},
		[]Result{Result{[2]int{1, 0}, 18, 1, true}, Result{[2]int{0, 0}, 17, 3, true}, Result{[2]int{2, 0}, 18, 2, true}})
}

func TestRandom(t *testing.T) {
	test(t, []byte("yasherhs"),
		[]Seq{seq(-1, "say"), seq(-1, "she"), seq(-1, "shr"), seq(-1, "he"), seq(-1, "her")},
		[]Result{Result{[2]int{1, 0}, 2, 3, true}, Result{[2]int{3, 0}, 3, 2, true}, Result{[2]int{4, 0}, 3, 3, true}})
}

func TestFailPartial(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "dlf"), seq(-1, "l")},
		[]Result{Result{[2]int{1, 0}, 18, 1, true}})
}

func TestMany(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "handle"), seq(-1, "andle"), seq(-1, "ndle"), seq(-1, "dle"), seq(-1, "le"), seq(-1, "e")},
		[]Result{Result{[2]int{5, 0}, 2, 1, true}, Result{[2]int{0, 0}, 14, 6, true}, Result{[2]int{1, 0}, 15, 5, true}, Result{[2]int{2, 0}, 16, 4, true}, Result{[2]int{3, 0}, 17, 3, true}, Result{[2]int{4, 0}, 18, 2, true}, Result{[2]int{5, 0}, 19, 1, true}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(-1, "handle"), seq(-1, "handl"), seq(-1, "hand"), seq(-1, "han"), seq(-1, "ha"), seq(-1, "a")},
		[]Result{Result{[2]int{4, 0}, 8, 2, true}, Result{[2]int{5, 0}, 9, 1, true}, Result{[2]int{5, 0}, 12, 1, true}, Result{[2]int{4, 0}, 14, 2, true}, Result{[2]int{5, 0}, 15, 1, true}, Result{[2]int{3, 0}, 14, 3, true}, Result{[2]int{2, 0}, 14, 4, true}, Result{[2]int{1, 0}, 14, 5, true}, Result{[2]int{0, 0}, 14, 6, true}})
}

func TestLong(t *testing.T) {
	test(t, []byte("macintosh"),
		[]Seq{seq(-1, "acintosh"), seq(-1, "in")},
		[]Result{Result{[2]int{1, 0}, 3, 2, true}, Result{[2]int{0, 0}, 1, 8, true}})
	test(t, []byte("macintosh"),
		[]Seq{seq(-1, "acintosh"), seq(-1, "in"), seq(-1, "tosh")},
		[]Result{Result{[2]int{1, 0}, 3, 2, true}, Result{[2]int{0, 0}, 1, 8, true}, Result{[2]int{2, 0}, 5, 4, true}})
	test(t, []byte("macintosh"),
		[]Seq{seq(-1, "acintosh"), seq(-1, "into"), seq(-1, "to"), seq(-1, "in")},
		[]Result{Result{[2]int{3, 0}, 3, 2, true}, Result{[2]int{1, 0}, 3, 4, true}, Result{[2]int{2, 0}, 5, 2, true}, Result{[2]int{0, 0}, 1, 8, true}})
}

func TestOffset(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq(0, "pot"), seq(18, "l")},
		[]Result{Result{[2]int{1, 0}, 18, 1, true}})
}

func TestChoices(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{
			Seq{0, []Choice{Choice{[]byte("The")}, Choice{[]byte("pot")}, Choice{[]byte("l")}}},
			Seq{0, []Choice{Choice{[]byte("The")}}},
			Seq{8, []Choice{Choice{[]byte("had")}, Choice{[]byte("ndle")}}},
		},
		[]Result{
			Result{[2]int{0, 0}, 0, 3, false},
			Result{[2]int{1, 0}, 0, 3, true},
			Result{[2]int{0, 1}, 4, 3, false},
			Result{[2]int{2, 0}, 8, 3, false},
			Result{[2]int{0, 2}, 18, 1, true},
			Result{[2]int{2, 1}, 16, 4, true},
		})
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
		r := ac.Index(input, make(chan int), make(chan struct{}))
		for _ = range r {
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
	r := ac.Index(reader, make(chan int), make(chan struct{}))
	for _ = range r {
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
	r := ac.Index(reader, make(chan int), make(chan struct{}))
	for _ = range r {
	}
}

func BenchmarkMatchingHardTree(b *testing.B) {
	b.StopTimer()
	reader := bytes.NewBuffer(benchmarkValue(b.N))
	ac := New(hardTree())
	b.StartTimer()
	r := ac.Index(reader, make(chan int), make(chan struct{}))
	for _ = range r {
	}
}
