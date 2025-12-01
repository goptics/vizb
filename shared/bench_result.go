package shared

type Stat struct {
	Type  string  `json:"type"`
	Value float64 `json:"value,omitempty"`
}

type BenchmarkResult struct {
	Name  string `json:"name,omitempty"`
	XAxis string `json:"xAxis,omitempty"`
	YAxis string `json:"yAxis,omitempty"`
	Stats []Stat `json:"stats"`
}

type Benchmark struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CPU         struct {
		Name  string `json:"name,omitempty"`
		Cores int    `json:"cores,omitempty"`
	} `json:"cpu"`
	OS       string `json:"os,omitempty"`
	Arch     string `json:"arch,omitempty"`
	Pkg      string `json:"pkg,omitempty"`
	Settings struct {
		Charts []string `json:"charts"`
		Sort   struct {
			Enabled bool   `json:"enabled"`
			Order   string `json:"order"`
		} `json:"sort"`
		ShowLabels bool `json:"showLabels"`
	} `json:"settings"`
	Data []BenchmarkResult `json:"data"`
}
