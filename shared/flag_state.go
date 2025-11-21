package shared

type flagState struct {
	Name         string
	OutputFile   string
	MemUnit      string
	TimeUnit     string
	NumberUnit   string
	Description  string
	Format       string
	GroupPattern string
	Sort         string
	Charts       []string
	ShowLabels   bool
}

var FlagState flagState = flagState{}
