package buffer

import "testing"

func TestReuse(t *testing.T) {
	b := Get()
	if len(*b) != DefaultBufferSize {
		t.Fatal(len(*b))
	}
	*b = (*b)[:1]
	Put(b)
	next := Get()
	if len(*next) != DefaultBufferSize {
		t.Fatal(len(*next))
	}
	Put(next)
}
func BenchmarkReuse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := Get()
		Put(p)
	}
}
