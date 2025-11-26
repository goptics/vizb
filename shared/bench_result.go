package shared

type Stat struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Per   string  `json:"per"`
}

type BenchmarkResult struct {
	Name  string `json:"name"`
	XAxis string `json:"xAxis"`
	YAxis string `json:"yAxis"`
	Stats []Stat `json:"stats"`
}

type Benchmark struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CPU         struct {
		Name  string `json:"name"`
		Cores int    `json:"cores"`
	} `json:"cpu"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Pkg      string `json:"pkg"`
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
