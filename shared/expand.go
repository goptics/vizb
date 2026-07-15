package shared

// ExpandStatsOntoAxis turns multi-stat points into long form: each stat becomes
// its own DataPoint with the stat's type (column label) written onto dim, and a
// single empty-type Stat carrying only the value. Empty Type is omitted from JSON
// (Stat.Type has omitempty) so the UI sees an undefined type and can fall back
// to the dataset name for the chart title.
//
// Points with no stats are passed through unchanged. dim is one of n/x/y/z.
func ExpandStatsOntoAxis(points []DataPoint, dim Dimension) []DataPoint {
	if len(points) == 0 {
		return points
	}

	out := make([]DataPoint, 0, len(points))
	for i := range points {
		p := &points[i]
		if len(p.Stats) == 0 {
			out = append(out, deepCloneData(*p))
			continue
		}
		for _, st := range p.Stats {
			np := DataPoint{
				Name:   p.Name,
				XAxis:  p.XAxis,
				YAxis:  p.YAxis,
				ZAxis:  p.ZAxis,
				Metric: p.Metric,
			}
			// Column / chart label was stored in Type before expand; write it onto dim.
			switch dim {
			case DimensionXAxis:
				np.XAxis = st.Type
			case DimensionYAxis:
				np.YAxis = st.Type
			case DimensionZAxis:
				np.ZAxis = st.Type
			default: // DimensionName
				np.Name = st.Type
			}
			var val *float64
			if st.Value != nil {
				v := *st.Value
				val = &v
			}
			np.Stats = []Stat{{Value: val}}
			out = append(out, np)
		}
	}
	return out
}

// EnsureAxis adds dim to axes when missing so injected column names stay visible
// to the UI identity pipeline. Same ordering rules as merge tag-axis injection.
func EnsureAxis(axes []Axis, dim Dimension) []Axis {
	return ensureInjectAxis(axes, dim)
}
