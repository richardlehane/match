Aho-Corasick multiple string matching algorithm with wildcards. 

This algorithm allows for sequences that are composed of subsequences that can contain wildcards including wildcards with min/max offsets.
The results returned are for the matches on subsequences (NOT full sequences). The index of those subsequences and the offset is returned.
It is up to clients to verify that the complete sequence that they are interested in has matched.

Example usage:
    
    w := wac.New()
    seq := wac.NewSeq(wac.Sub{0, 0, []byte{'a','b'}}, wac.Sub{2, 4, []byte{'a','d'}}, wac.Sub{0, -1, []byte{'r', 'a'}})
    w.Add(seq)
    for result := range w.Index(bytes.NewBuffer([]byte("abracadabra"))) {
	  fmt.Println(result.Index, "-", result.Offset)
    }

Install with `go get github.com/richardlehane/wac`