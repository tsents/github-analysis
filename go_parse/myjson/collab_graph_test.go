package myjson

import "testing"


// Easyjson improved by 3* the speed! and also removed allocs. now all on stack.
// NEW TESTING SHOWED THAT IT IS IN FACT *3 SLOWER IN PRACTICE.
// THIS IS BECAUSE THE UNMARSHELLING TAKES A LONG TIME, AND CALLS INSIDE FUNTIONS.
// Might be that jsoniter can stop for small json while easyjson needs to read more?
func BenchmarkCollabGraphAction(b *testing.B) {
	data := []byte(`{"actor":{"id":1},"repo":{"id":2}}`)

	b.ReportAllocs() // tells Go to measure allocations
	for i := 0; i < b.N; i++ {
		_, _ = CollabGraphAction(data)
	}
}
