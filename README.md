# `promrwconv` - Prometheus remote write protocol conversion library.

The `promrwconv` helps to convert Prometheus metrics from [text-based exposition format](https://github.com/prometheus/docs/blob/master/content/docs/instrumenting/exposition_formats.md#text-based-format) to  [Prometheus Remote Write](https://github.com/prometheus/prometheus/tree/main/prompb) request

## ⚠️ Work in progress - API may change

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

req, reqErr := promrwconv.MetricsToPromRWRequest(metrics)
if reqErr != nil {
	return fmt.Error("failed to convert metrics to remote write request: %w", reqErr)
}

data, mErr := proto.Marshal(req)
if mErr != nil {
    return fmt.Errorf("failed to marshal request to protobuf: %w", mErr)
}

encoded := snappy.Encode(nil, data)

req, httpReqErr := http.NewRequest("POST", "prometheus-remote-write-url-here", bytes.NewReader(encoded))
if httpReqErr != nil {
    return fmt.Errorf("failed to create http request: %w", httpReqErr)
}

// You are ready to send the request.
```

To learn more about Prometheus Remote Write protocol:
- [Integrating Long-Term Storage with Prometheus [A] - Julius Volz, Prometheus](https://www.youtube.com/watch?v=MuHkckZg5L0)
- [Things you wish you never knew about the Prometheus Remote Write API](https://drive.google.com/file/d/0B0tWC_gFU85NY1Zub3hTVUQzb0U/view?resourcekey=0-rbBZShSxVNRIV0dFfQRGig)