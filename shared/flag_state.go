package shared

type flagState struct {
	Tag          string
	Name         string
	OutputFile   string
	MemUnit      string
	TimeUnit     string
	NumberUnit   string
	Description  string
	GroupPattern string
	GroupRegex   string
	Sort         string
	Charts       []string
	Group        []string
	ShowLabels   bool
	FilterRegex  string
	Scale        string
	TagAxis      string
	Parser       string
	DataURL      string
	ChartSpecs   []string
}

var FlagState flagState = flagState{}
