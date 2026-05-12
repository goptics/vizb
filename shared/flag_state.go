package shared

type flagState struct {
	OutputFile   string
	GroupPattern string
	GroupRegex   string
}

var FlagState flagState = flagState{}
