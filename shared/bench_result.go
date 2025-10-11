package shared

type Stat struct {
	Type     string  `json:"type"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit"`
	NotPerOp bool    `json:"notPerOp"`
}

type BenchmarkResult struct {
	Name     string `json:"name"`
	Workload string `json:"workload"`
	Subject  string `json:"subject"`
	Stats    []Stat `json:"stats"`
}
