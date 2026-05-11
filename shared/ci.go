package shared

import "time"

type BenchmarkResult struct {
	Name        string  `json:"name"`
	Pkg         string  `json:"pkg,omitempty"`
	NsPerOp     float64 `json:"ns_per_op"`
	MBPerSec    float64 `json:"mb_per_sec,omitempty"`
	BytesPerOp  float64 `json:"bytes_per_op"`
	AllocsPerOp float64 `json:"allocs_per_op"`
}

type Run struct {
	Version    string            `json:"version"`
	Tag        string            `json:"tag,omitempty"`
	Date       time.Time         `json:"date"`
	Branch     string            `json:"branch,omitempty"`
	Goos       string            `json:"goos,omitempty"`
	Goarch     string            `json:"goarch,omitempty"`
	CPU        string            `json:"cpu,omitempty"`
	Benchmarks []BenchmarkResult `json:"benchmarks"`
}

type HistoryMeta struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type History struct {
	Meta HistoryMeta `json:"meta"`
	Runs []Run       `json:"runs"`
}
