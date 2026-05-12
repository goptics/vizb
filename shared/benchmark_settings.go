package shared

type benchmarkSettings struct {
	Name        string
	Description string
	MemUnit     string
	TimeUnit    string
	NumberUnit  string
	Sort        string
	Charts      []string
	ShowLabels  bool
	FilterRegex string
	Scale       string
}

var BenchSettings benchmarkSettings

func NewBenchmark(data []BenchmarkData) Benchmark {
	b := Benchmark{
		Name:        BenchSettings.Name,
		Description: BenchSettings.Description,
		Pkg:         Pkg,
		OS:          OS,
		Arch:        Arch,
		Data:        data,
	}
	b.CPU.Name = CPU
	b.CPU.Cores = CPUCount
	b.Settings.Charts = BenchSettings.Charts
	b.Settings.ShowLabels = BenchSettings.ShowLabels
	b.Settings.Scale = BenchSettings.Scale
	b.Settings.Sort.Enabled = BenchSettings.Sort != ""
	if b.Settings.Sort.Enabled {
		b.Settings.Sort.Order = BenchSettings.Sort
	} else {
		b.Settings.Sort.Order = "asc"
	}
	return b
}
