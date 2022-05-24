package dwac

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

func test(t *testing.T, a []byte, b []Seq, w []Seq, expect []Result) {
	dwac := New(b)
	output, resume := dwac.Index(bytes.NewBuffer(a))
	results := make([]Result, 0)
	for res := range output {
		if res.Index[0] == -1 {
			resume <- w
			continue
		}
		results = append(results, res)
	}
	if !equal(expect, results) {
		t.Errorf("Index fail; Expecting: %v, Got: %v", expect, results)
	}
}

func seq(s string) Seq {
	return Seq{[]int64{64}, []Choice{Choice{[]byte(s)}}}
}

// Tests (the test strings are taken from John Graham-Cumming's lua implementation: https://github.com/jgrahamc/aho-corasick-lua Copyright (c) 2013 CloudFlare)
func TestWikipedia(t *testing.T) {
	test(t, []byte("abccab"),
		[]Seq{seq("a"), seq("ab"), seq("bc"), seq("bca"), seq("c"), seq("caa")},
		nil,
		[]Result{Result{[2]int{0, 0}, 0, 1}, Result{[2]int{1, 0}, 0, 2}, Result{[2]int{2, 0}, 1, 2}, Result{[2]int{4, 0}, 2, 1}, Result{[2]int{4, 0}, 3, 1}, Result{[2]int{0, 0}, 4, 1}, Result{[2]int{1, 0}, 4, 2}})
}

func TestSimple(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("poto")},
		nil,
		[]Result{})
	test(t, []byte("The pot had a handle The"),
		[]Seq{Seq{[]int64{0}, []Choice{Choice{[]byte("The")}}}},
		nil,
		[]Result{Result{[2]int{0, 0}, 0, 3}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("pot")},
		nil,
		[]Result{Result{[2]int{0, 0}, 4, 3}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("pot ")},
		nil,
		[]Result{Result{[2]int{0, 0}, 4, 4}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("ot h")},
		nil,
		[]Result{Result{[2]int{0, 0}, 5, 4}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("andle")},
		nil,
		[]Result{Result{[2]int{0, 0}, 15, 5}})
}

func TestMultipleNonoverlapping(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("h")},
		nil,
		[]Result{Result{[2]int{0, 0}, 1, 1}, Result{[2]int{0, 0}, 8, 1}, Result{[2]int{0, 0}, 14, 1}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("ha"), seq("he")},
		nil,
		[]Result{Result{[2]int{1, 0}, 1, 2}, Result{[2]int{0, 0}, 8, 2}, Result{[2]int{0, 0}, 14, 2}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("pot"), seq("had")},
		nil,
		[]Result{Result{[2]int{0, 0}, 4, 3}, Result{[2]int{1, 0}, 8, 3}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("pot"), seq("had"), seq("hod")},
		nil,
		[]Result{Result{[2]int{0, 0}, 4, 3}, Result{[2]int{1, 0}, 8, 3}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("The"), seq("pot"), seq("had"), seq("hod"), seq("andle")},
		nil,
		[]Result{Result{[2]int{0, 0}, 0, 3}, Result{[2]int{1, 0}, 4, 3}, Result{[2]int{2, 0}, 8, 3}, Result{[2]int{4, 0}, 15, 5}})
}

func TestOverlapping(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("Th"), seq("he pot"), seq("The"), seq("pot h")},
		nil,
		[]Result{Result{[2]int{0, 0}, 0, 2}, Result{[2]int{2, 0}, 0, 3}, Result{[2]int{1, 0}, 1, 6}, Result{[2]int{3, 0}, 4, 5}})
}

func TestNesting(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("handle"), seq("hand"), seq("and"), seq("andle")},
		nil,
		[]Result{Result{[2]int{1, 0}, 14, 4}, Result{[2]int{2, 0}, 15, 3}, Result{[2]int{0, 0}, 14, 6}, Result{[2]int{3, 0}, 15, 5}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("handle"), seq("hand"), seq("an"), seq("n")},
		nil,
		[]Result{Result{[2]int{2, 0}, 15, 2}, Result{[2]int{3, 0}, 16, 1}, Result{[2]int{1, 0}, 14, 4}, Result{[2]int{0, 0}, 14, 6}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("dle"), seq("l"), seq("le")},
		nil,
		[]Result{Result{[2]int{1, 0}, 18, 1}, Result{[2]int{0, 0}, 17, 3}, Result{[2]int{2, 0}, 18, 2}})
}

func TestRandom(t *testing.T) {
	test(t, []byte("yasherhs"),
		[]Seq{seq("say"), seq("she"), seq("shr"), seq("he"), seq("her")},
		nil,
		[]Result{Result{[2]int{1, 0}, 2, 3}, Result{[2]int{3, 0}, 3, 2}, Result{[2]int{4, 0}, 3, 3}})
}

func TestFailPartial(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("dlf"), seq("l")},
		nil,
		[]Result{Result{[2]int{1, 0}, 18, 1}})
}

func TestMany(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("handle"), seq("andle"), seq("ndle"), seq("dle"), seq("le"), seq("e")},
		nil,
		[]Result{Result{[2]int{5, 0}, 2, 1}, Result{[2]int{0, 0}, 14, 6}, Result{[2]int{1, 0}, 15, 5}, Result{[2]int{2, 0}, 16, 4}, Result{[2]int{3, 0}, 17, 3}, Result{[2]int{4, 0}, 18, 2}, Result{[2]int{5, 0}, 19, 1}})
	test(t, []byte("The pot had a handle"),
		[]Seq{seq("handle"), seq("handl"), seq("hand"), seq("han"), seq("ha"), seq("a")},
		nil,
		[]Result{Result{[2]int{4, 0}, 8, 2}, Result{[2]int{5, 0}, 9, 1}, Result{[2]int{5, 0}, 12, 1}, Result{[2]int{4, 0}, 14, 2}, Result{[2]int{5, 0}, 15, 1}, Result{[2]int{3, 0}, 14, 3}, Result{[2]int{2, 0}, 14, 4}, Result{[2]int{1, 0}, 14, 5}, Result{[2]int{0, 0}, 14, 6}})
}

func TestLong(t *testing.T) {
	test(t, []byte("macintosh"),
		[]Seq{seq("acintosh"), seq("in")},
		nil,
		[]Result{Result{[2]int{1, 0}, 3, 2}, Result{[2]int{0, 0}, 1, 8}})
	test(t, []byte("macintosh"),
		[]Seq{seq("acintosh"), seq("in"), seq("tosh")},
		nil,
		[]Result{Result{[2]int{1, 0}, 3, 2}, Result{[2]int{0, 0}, 1, 8}, Result{[2]int{2, 0}, 5, 4}})
	test(t, []byte("macintosh"),
		[]Seq{seq("acintosh"), seq("into"), seq("to"), seq("in")},
		nil,
		[]Result{Result{[2]int{3, 0}, 3, 2}, Result{[2]int{1, 0}, 3, 4}, Result{[2]int{2, 0}, 5, 2}, Result{[2]int{0, 0}, 1, 8}})
}

func TestOffset(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{Seq{[]int64{0}, []Choice{Choice{[]byte("pot")}}}, Seq{[]int64{18}, []Choice{Choice{[]byte("l")}}}},
		nil,
		[]Result{Result{[2]int{1, 0}, 18, 1}})
}

func TestChoices(t *testing.T) {
	test(t, []byte("The pot had a handle"),
		[]Seq{
			Seq{[]int64{0, 18, -1}, []Choice{Choice{[]byte("The")}, Choice{[]byte("pot")}, Choice{[]byte("l")}}},
			Seq{[]int64{-1}, []Choice{Choice{[]byte("The")}}},
			Seq{[]int64{8, -1}, []Choice{Choice{[]byte("had")}, Choice{[]byte("ndle")}}},
		},
		nil,
		[]Result{
			Result{[2]int{0, 0}, 0, 3},
			Result{[2]int{1, 0}, 0, 3},
			Result{[2]int{0, 1}, 4, 3},
			Result{[2]int{2, 0}, 8, 3},
			Result{[2]int{0, 2}, 18, 1},
			Result{[2]int{2, 1}, 16, 4},
		})
}

// Benchmarks
func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = New([]Seq{seq("handle"), seq("handl"), seq("hand"), seq("han"), seq("ha"), seq("a")})
	}
}

func BenchmarkIndex(b *testing.B) {
	b.StopTimer()
	dwac := New([]Seq{seq("handle"), seq("handl"), seq("hand"), seq("han"), seq("ha"), seq("a")})
	reader := bytes.NewBuffer([]byte("The pot had a handle"))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		results, resume := dwac.Index(reader)
		for res := range results {
			if res.Index[0] == -1 {
				resume <- nil
				continue
			}
		}
	}
}

// following benchmark code is from <http://godoc.org/code.google.com/p/ahocorasick> for comparison
func benchmarkValue(n int) []byte {
	input := make([]byte, n)
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			input[i] = 'a'
		} else {
			input[i] = 'b'
		}
	}
	return input
}

func hardTree() []Seq {
	ret := make([]Seq, 0, 2500)
	str := ""
	for i := 0; i < 2500; i++ {
		// We add a 'q' to the end to make sure we never actually match
		ret = append(ret, seq(str+string([]byte{'a', byte(i % 26), 'q'})))
		if i%26 == 25 {
			str = str + string([]byte{'a', byte(len(str) % 2)})
		}
	}
	return ret
}

func BenchmarkMatchingNoMatch(b *testing.B) {
	b.StopTimer()
	reader := bytes.NewBuffer(benchmarkValue(b.N))
	dwac := New([]Seq{seq("abababababababd"),
		seq("abababb"),
		seq("abababababq")})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		results, resume := dwac.Index(reader)
		for res := range results {
			if res.Index[0] == -1 {
				resume <- nil
				continue
			}
		}
	}
}

func BenchmarkMatchingManyMatches(b *testing.B) {
	b.StopTimer()
	reader := bytes.NewBuffer(benchmarkValue(b.N))
	dwac := New([]Seq{seq("ab"),
		seq("ababababababab"),
		seq("ababab"),
		seq("ababababab")})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		results, resume := dwac.Index(reader)
		for res := range results {
			if res.Index[0] == -1 {
				resume <- nil
				continue
			}
		}
	}
}

func BenchmarkMatchingHardTree(b *testing.B) {
	b.StopTimer()
	reader := bytes.NewBuffer(benchmarkValue(b.N))
	dwac := New([]Seq{seq("abababababababd"),
		seq("abababb"),
		seq("abababababq")})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		results, resume := dwac.Index(reader)
		for res := range results {
			if res.Index[0] == -1 {
				resume <- nil
				continue
			}
		}
	}
}
