package shared

type BenchEvent struct {
	Action string `json:"Action"`
	Test   string `json:"Test,omitempty"`
	Output string `json:"Output,omitempty"`
}
