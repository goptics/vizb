package shared

type actionState struct {
	SHA          string
	Tag          string
	Branch       string
	Append       string
	Output       string
	Keep         int
	GroupPattern string
	GroupRegex   string
}

var ActionState actionState
