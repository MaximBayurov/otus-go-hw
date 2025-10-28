//go:build bench_standalone

package hw10programoptimization

import (
	"archive/zip"
	"testing"
)

func BenchmarkGetDomainStat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		r, _ := zip.OpenReader("testdata/users.dat.zip")
		data, _ := r.File[0].Open()

		b.StartTimer()
		_, _ = GetDomainStat(data, "biz")
		b.StopTimer()

		r.Close()
	}
}
