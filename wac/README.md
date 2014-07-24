Aho-Corasick multiple string matching algorithm with wildcards, choices and a leading offset. 

This algorithm allows for sequences that are composed of subsequences that can contain an arbitrary number of wildcards. Sequences can also be given a maximum offset that defines the maximum byte position of the first sub-sequence.

Subsequences are groups of choices - any can match for the subsequence to trigger a result.

The results returned are for the matches on subsequences (NOT the full sequences). The index of those subsequences and the offset is returned.
It is up to clients to verify that the complete sequence that they are interested in has matched.

Example usage:
    
    w := wac.New()
    seq := wac.Seq(1, []wac.Choice{wac.Choice{[]byte{'b'}}, wac.Choice{[]byte{'a','d'}}, wac.Choice{[]byte{'r', 'a'},[]byte{'l', 'a'}}})
    // close quit to finish search early
    quit := make(chan {}struct)
    for result := range w.Index(bytes.NewBuffer([]byte("abracadabra")), quit) {
      fmt.Println(result.Index, "-", result.Offset)
    }