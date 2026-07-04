package shared

type Dimension string

const (
	DimensionName  Dimension = "n"
	DimensionXAxis Dimension = "x"
	DimensionYAxis Dimension = "y"
	DimensionZAxis Dimension = "z"
)

// AxisKey returns the dataset.axes key for this inject dimension.
func (d Dimension) AxisKey() string {
	switch d {
	case DimensionXAxis:
		return "x"
	case DimensionYAxis:
		return "y"
	case DimensionZAxis:
		return "z"
	default:
		return "name"
	}
}
