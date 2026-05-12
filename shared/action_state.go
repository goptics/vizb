package shared

type actionState struct {
	SHA          string
	Tag          string
	Branch       string
	Merge        string
	Output       string
	HTML         bool
	Keep         int
	GroupPattern string
	GroupRegex   string
}

var ActionState actionState
