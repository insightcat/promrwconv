package promrwconv

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/prometheus/prompb"
)

var (
	// readerPool keeps pool of *bytes.Reader to reduce additional allocations
	// in parseMetrics and countMetricLines functions.
	readerPool = sync.Pool{New: func() interface{} { return &bytes.Reader{} }}
	// commentPrefix holds line prefix as byte slice with which the comment begins.
	commentPrefix = []byte("#")
	// labelsOpenMark holds mark labels after which labels begin.
	labelsOpenMark = []byte("{")
	// labelsCloseMark holds mark which closes the labels.
	labelsCloseMark = []byte("}")
	// spaceSep holds space separator as byte slice.
	spaceSep = []byte(" ")
	// commaSep holds comma separator as byte slice.
	commaSep = []byte(",")
	// equalSep holds equal separator as byte slice.
	equalSep = []byte("=")
)

// MetricsToPromRWRequest converts given bytes slice with text based Prometheus
// format to Prometheus remote write Request.
func MetricsToPromRWRequest(metrics []byte) (*prompb.WriteRequest, error) {
	return parseMetrics(metrics)
}

func parseMetrics(metrics []byte) (*prompb.WriteRequest, error) {
	lines, linesErr := countMetricLines(metrics)
	if linesErr != nil {
		return nil, fmt.Errorf("failed to parse metrics: %w", linesErr)
	}

	request := prompb.WriteRequest{
		Timeseries: make([]*prompb.TimeSeries, 0, lines),
	}

	reader := readerPool.Get().(*bytes.Reader)
	reader.Reset(metrics)
	defer readerPool.Put(reader)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if bytes.HasPrefix(scanner.Bytes(), commentPrefix) {
			continue
		}

		lStartIdx := bytes.Index(scanner.Bytes(), labelsOpenMark)
		if lStartIdx == -1 {
			ts, tsErr := parseMetric(scanner.Bytes())
			if tsErr != nil {
				return nil, tsErr
			}

			request.Timeseries = append(request.Timeseries, ts)

			continue
		}

		lEndIdx := bytes.Index(scanner.Bytes(), labelsCloseMark)

		metricParts := bytes.Split(scanner.Bytes(), scanner.Bytes()[lStartIdx:lEndIdx+1])
		if len(metricParts) != 2 {
			return nil, fmt.Errorf("invalid metrics format: '%s'", scanner.Bytes())
		}

		metricLine := bytes.NewBuffer(make([]byte, 0, len(metricParts[0])+len(metricParts[1])+1))
		metricLine.Write(metricParts[0])
		metricLine.WriteString(" ")
		metricLine.Write(bytes.TrimSpace(metricParts[1]))

		labelsLine := scanner.Bytes()[lStartIdx+1 : lEndIdx]

		ts, tsErr := parseMetric(metricLine.Bytes())
		if tsErr != nil {
			return nil, tsErr
		}

		labels, labelsErr := parseLabels(labelsLine)
		if labelsErr != nil {
			return nil, labelsErr
		}

		ts.Labels = append(ts.Labels, labels...)

		request.Timeseries = append(request.Timeseries, ts)
	}

	return &request, nil
}

func parseMetric(metricLine []byte) (*prompb.TimeSeries, error) {
	parts := bytes.Split(metricLine, spaceSep)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid metric line format: '%s'", metricLine)
	}

	val, parseValErr := strconv.ParseFloat(string(bytes.TrimSpace(parts[1])), 64)
	if parseValErr != nil {
		return nil, fmt.Errorf("invalid metric format: failed to convert byte slice to float64: %w", parseValErr)
	}

	ts := prompb.TimeSeries{
		Labels: []*prompb.Label{
			{
				Name:  "__name__",
				Value: string(parts[0]),
			},
		},
		Samples: []prompb.Sample{
			{
				Value:     val,
				Timestamp: time.Now().UTC().Unix(),
			},
		},
	}

	return &ts, nil
}

func parseLabels(labelsLine []byte) ([]*prompb.Label, error) {
	labelsKVs := bytes.Split(labelsLine, commaSep)
	labels := make([]*prompb.Label, 0, len(labelsKVs))

	for _, kv := range labelsKVs {
		labelKV := bytes.Split(kv, equalSep)
		if len(labelKV) != 2 {
			return nil, fmt.Errorf("invalid label format: %s", kv)
		}

		label := prompb.Label{
			Name:  string(labelKV[0]),
			Value: string(bytes.Trim(labelKV[1], `"`)),
		}

		labels = append(labels, &label)
	}

	return labels, nil
}

func countMetricLines(metrics []byte) (int, error) {
	count := 0

	reader := readerPool.Get().(*bytes.Reader)
	reader.Reset(metrics)
	defer readerPool.Put(reader)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if bytes.HasPrefix(scanner.Bytes(), commentPrefix) {
			continue
		}

		count++
	}

	if scanner.Err() != nil {
		return 0, fmt.Errorf("failed to count lines with metrics: %w", scanner.Err())
	}

	return count, nil
}
