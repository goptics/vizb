package shared

type flagState struct {
	Name        string
	OutputFile  string
	MemUnit     string
	TimeUnit    string
	AllocUnit   string
	Description string
	Format      string
}

var FlagState flagState = flagState{}
