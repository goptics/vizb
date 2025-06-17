package shared

type flagState struct {
	Name        string
	OutputFile  string
	MemUnit     string
	TimeUnit    string
	Description string
	Separator   string
}

var FlagState flagState = flagState{}
