package promrwconv

import (
	"io"
	"os"
	"testing"
)

func Benchmark_countMetricLines(b *testing.B) {
	f, openErr := os.Open("testdata/metrics.txt")
	if openErr != nil {
		b.Fatalf("failed to open test file with metrics: %v", openErr)
	}

	metrics, readErr := io.ReadAll(f)
	if readErr != nil {
		b.Fatalf("failed to read test file with metrics: %v", readErr)
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		countMetricLines(metrics)
	}
}

func Benchmark_parseLabels(b *testing.B) {
	labels := []byte(`collector="uname",collector2="uname1",collector1="uname1"`)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parseLabels(labels)
	}
}

func BenchmarkName(b *testing.B) {
	metric := []byte(`node_time_zone_offset_seconds 7200`)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parseMetric(metric)
	}
}
