package myjson

import "testing"


func BenchmarkCollabGraphAction(b *testing.B) {
	data := []byte(`{"actor":{"id":1},"repo":{"id":2}}`)

	b.ReportAllocs() // 👈 tells Go to measure allocations
	for i := 0; i < b.N; i++ {
		_, _ = CollabGraphAction(data)
	}
}
