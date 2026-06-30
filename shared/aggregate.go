package shared

import "fmt"

// AggregateDataPoints groups DataPoints by (Name, XAxis, YAxis, ZAxis) and sums
// Stat.Value for matching stat types within each group. Order of first occurrence
// is preserved. Use when CSV/JSON input contains multiple rows for the same logical
// data point (e.g. multiple sales on the same date/region) and the goal is a single
// summed value per combination.
func AggregateDataPoints(points []DataPoint) []DataPoint {
	type key struct{ name, x, y, z string }

	order := make([]key, 0, len(points))
	groups := make(map[key]*DataPoint, len(points))

	for i := range points {
		dp := &points[i]
		k := key{dp.Name, dp.XAxis, dp.YAxis, dp.ZAxis}

		existing, found := groups[k]
		if !found {
			clone := DataPoint{
				Name:   dp.Name,
				XAxis:  dp.XAxis,
				YAxis:  dp.YAxis,
				ZAxis:  dp.ZAxis,
				Metric: dp.Metric,
				Stats:  make([]Stat, len(dp.Stats)),
			}
			copy(clone.Stats, dp.Stats)
			groups[k] = &clone
			order = append(order, k)
			continue
		}

		statIdx := make(map[string]int, len(existing.Stats))
		for i, s := range existing.Stats {
			statIdx[fmt.Sprintf("%s|%s", s.Type, s.Symbol)] = i
		}

		for _, s := range dp.Stats {
			sk := fmt.Sprintf("%s|%s", s.Type, s.Symbol)
			if idx, ok := statIdx[sk]; ok {
				if s.Value != nil {
					var base float64
					if existing.Stats[idx].Value != nil {
						base = *existing.Stats[idx].Value
					}
					v := base + *s.Value
					existing.Stats[idx].Value = &v
				}
			} else {
				existing.Stats = append(existing.Stats, s)
				statIdx[sk] = len(existing.Stats) - 1
			}
		}
	}

	result := make([]DataPoint, 0, len(order))
	for _, k := range order {
		result = append(result, *groups[k])
	}
	return result
}

// CollapseDataPointsByKey merges DataPoints that share the same
// (Name, XAxis, YAxis, ZAxis) by appending stats without summing. Duplicate
// stat types are preserved so tabular rows overlay at one category in the UI.
func CollapseDataPointsByKey(points []DataPoint) []DataPoint {
	type key struct{ name, x, y, z string }

	order := make([]key, 0, len(points))
	groups := make(map[key]*DataPoint, len(points))

	for i := range points {
		dp := &points[i]
		k := key{dp.Name, dp.XAxis, dp.YAxis, dp.ZAxis}

		existing, found := groups[k]
		if !found {
			clone := DataPoint{
				Name:   dp.Name,
				XAxis:  dp.XAxis,
				YAxis:  dp.YAxis,
				ZAxis:  dp.ZAxis,
				Metric: dp.Metric,
				Stats:  make([]Stat, len(dp.Stats)),
			}
			copy(clone.Stats, dp.Stats)
			groups[k] = &clone
			order = append(order, k)
			continue
		}

		existing.Stats = append(existing.Stats, dp.Stats...)
	}

	result := make([]DataPoint, 0, len(order))
	for _, k := range order {
		result = append(result, *groups[k])
	}
	return result
}
