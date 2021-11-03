package promrwconv

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/prometheus/prometheus/prompb"
)

func Test_parseMetricLine(t *testing.T) {

}

func Test_parseLabels(t *testing.T) {
	type tcase struct {
		labels  []byte
		wantErr error
		want    []*prompb.Label
	}

	tests := map[string]tcase{
		"OK": {
			labels:  []byte(`cpu="8",mode="nice"`),
			wantErr: nil,
			want:    []*prompb.Label{{Name: "cpu", Value: "8"}, {Name: "mode", Value: "nice"}},
		},
		"Curly_Braces": {
			labels:  []byte(`{cpu="8",mode="nice"}`),
			wantErr: fmt.Errorf(`curly bracers should be trimmed: %s`, `{cpu="8",mode="nice"}`),
			want:    nil,
		},
		"Error": {
			labels:  []byte(`cpu="8",mode="`),
			wantErr: fmt.Errorf("invalid label format: %s", `cpu="8",mode="`),
			want:    nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseLabels(tc.labels)
			if tc.wantErr != nil {
				if !reflect.DeepEqual(err, tc.wantErr) {
					t.Fatalf("parseLabels() error = %v, wantErr = %v", err, tc.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("parseLabels() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("parseLabels() got = %v, want = %v", got, tc.want)
			}
		})
	}
}

func Test_countMetricLines(t *testing.T) {
	type tcase struct {
		metricsFile string
		wantErr     error
		want        int
	}

	tests := map[string]tcase{
		"OK": {
			metricsFile: "testdata/test_metrics_count.txt",
			wantErr:     nil,
			want:        4,
		},
		"0_Lines": {
			metricsFile: "testdata/test_metrics_count_0.txt",
			wantErr:     nil,
			want:        0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			f, openErr := os.Open(tc.metricsFile)
			if openErr != nil {
				t.Fatalf("failed to open test file with metrics: %v", openErr)
			}

			metrics, readErr := io.ReadAll(f)
			if readErr != nil {
				t.Fatalf("failed to read test file with metrics: %v", readErr)
			}

			got, err := countMetricLines(metrics)
			if tc.wantErr != nil {
				if !reflect.DeepEqual(err, tc.wantErr) {
					t.Fatalf("countMetricLines() error = %v, wantErr %v", err, tc.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("countMetricLines() unexpected error: %v", err)
			}

			if got != tc.want {
				t.Errorf("countMetricLines() got = %v, want %v", got, tc.want)
			}
		})
	}
}
