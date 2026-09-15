package producer

import (
	"fmt"
	"testing"
)

func BenchmarkProcessFile(b *testing.B) {
	for _, workers := range []int{1, 2, 4, 8} {
		b.Run(fmt.Sprintf("workers-%d", workers), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := ProcessFile("testdata/large.csv", workers)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}