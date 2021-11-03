# promrwconv

The `promrwconv` helps to convert Prometheus metrics
from [text-based exposition format](https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md#text-based-format)
to  [Prometheus Remote Write](https://github.com/prometheus/prometheus/tree/main/prompb) request

## ⚠️ Work in progress

The main API which is `MetricsToPromRWRequest` function is stable. API may extend to add more functionality. The functions still has a lot of allocations which should be avoided.

## Usage

```go
metrics := []byte(`# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 3.677e-05
go_gc_duration_seconds{quantile="0.25"} 4.3408e-05
go_gc_duration_seconds{quantile="0.5"} 6.4948e-05
go_gc_duration_seconds{quantile="0.75"} 9.2592e-05
go_gc_duration_seconds{quantile="1"} 0.027297214
go_gc_duration_seconds_sum 0.102844028
go_gc_duration_seconds_count 394`)

rwReq, rwReqErr := promrwconv.MetricsToPromRWRequest(metrics)
if rwReqErr != nil {
    return fmt.Error("failed to convert metrics to remote write request: %w", rwReqErr)
}

data, mErr := proto.Marshal(rwReq)
if mErr != nil {
    return fmt.Errorf("failed to marshal request to protobuf: %w", mErr)
}

encoded := snappy.Encode(nil, data)

req, reqErr := http.NewRequest("POST", "prometheus-remote-write-url-here", bytes.NewReader(encoded))
if reqErr != nil {
    return fmt.Errorf("failed to create http request: %w", reqErr)
}

req.Header.Set("Content-Type", "application/x-protobuf")
req.Header.Set("Content-Encoding", "snappy")
req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

// You are ready to send the request.
```

To learn more about Prometheus Remote Write protocol:

- [Integrating Long-Term Storage with Prometheus [A] - Julius Volz, Prometheus](https://www.youtube.com/watch?v=MuHkckZg5L0)
- [Things you wish you never knew about the Prometheus Remote Write API](https://drive.google.com/file/d/0B0tWC_gFU85NY1Zub3hTVUQzb0U/view?resourcekey=0-rbBZShSxVNRIV0dFfQRGig)

## Benchmarks

```text
goos: darwin
goarch: amd64
pkg: github.com/insightcat/promrwconv
cpu: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz
Benchmark_parseMetrics-12                   3013            345070 ns/op          182802 B/op       6032 allocs/op
Benchmark_parseLabels-12                 1629504               747.7 ns/op           464 B/op         17 allocs/op
Benchmark_parseMetricLine-12             3677604               326.5 ns/op           188 B/op          7 allocs/op
Benchmark_countMetricLines-12             125007              9373 ns/op            4097 B/op          1 allocs/op
```