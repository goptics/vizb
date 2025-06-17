package shared

type flagState struct {
	Name        string
	OutputFile  string
	MemUnit     string
	TimeUnit    string
	Description string
	Separator   string
	ShowVersion bool
}

var FlagState flagState = flagState{}
