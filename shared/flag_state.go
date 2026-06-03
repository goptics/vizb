package shared

type flagState struct {
	Tag             string
	Name            string
	OutputFile      string
	MemUnit         string
	TimeUnit        string
	NumberUnit      string
	Description     string
	GroupPattern    string
	GroupRegex      string
	Sort            string
	Charts          []string
	ShowLabels      bool
	FilterRegex     string
	Scale           string
	TagAxis string
	Parser          string
	DataURL         string
}

var FlagState flagState = flagState{}
