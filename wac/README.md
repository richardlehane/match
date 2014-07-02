Aho-Corasick multiple string matching algorithm with wildcards. 

This algorithm allows for sequences that are composed of sub-sequences that can contain an arbitrary number of wildcards. Sequences can also be given a maximum offset that defines the maximum byte position of the first sub-sequence.

The results returned are for the matches on subsequences (NOT the full sequences). The index of those subsequences and the offset is returned.
It is up to clients to verify that the complete sequence that they are interested in has matched.

Example usage:
    
    w := wac.New()
    seq := wac.NewSeq(1, []byte{'b'}, []byte{'a','d'}}, []byte{'r', 'a'})
    w.Add(seq)
    for result := range w.Index(bytes.NewBuffer([]byte("abracadabra"))) {
	  fmt.Println(result.Index, "-", result.Offset)
    }