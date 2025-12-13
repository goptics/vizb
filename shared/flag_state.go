package shared

type flagState struct {
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
	ShowLabels   bool
	FilterRegex  string
}

var FlagState flagState = flagState{}
