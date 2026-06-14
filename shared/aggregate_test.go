package shared

import "testing"

func statVal(stats []Stat, typ string) (float64, bool) {
	for _, s := range stats {
		if s.Type == typ {
			return s.Value, true
		}
	}
	return 0, false
}

func TestAggregateDataPoints_SumsDuplicateKeys(t *testing.T) {
	in := []DataPoint{
		{Name: "sales", XAxis: "2024-01-01", YAxis: "East", Stats: []Stat{{Type: "amount", Value: 100}}},
		{Name: "sales", XAxis: "2024-01-01", YAxis: "East", Stats: []Stat{{Type: "amount", Value: 250}}},
		{Name: "sales", XAxis: "2024-01-01", YAxis: "West", Stats: []Stat{{Type: "amount", Value: 40}}},
	}

	out := AggregateDataPoints(in)

	if len(out) != 2 {
		t.Fatalf("expected 2 grouped points, got %d", len(out))
	}

	v, ok := statVal(out[0].Stats, "amount")
	if !ok || v != 350 {
		t.Fatalf("East amount: want 350, got %v (ok=%v)", v, ok)
	}
	if out[0].YAxis != "East" {
		t.Fatalf("first group should preserve first-seen order (East), got %q", out[0].YAxis)
	}
}

func TestAggregateDataPoints_PreservesUniqueKeys(t *testing.T) {
	in := []DataPoint{
		{XAxis: "A", Stats: []Stat{{Type: "v", Value: 1}}},
		{XAxis: "B", Stats: []Stat{{Type: "v", Value: 2}}},
	}

	out := AggregateDataPoints(in)

	if len(out) != 2 {
		t.Fatalf("unique keys must stay separate, got %d", len(out))
	}
}

func TestAggregateDataPoints_SumsPerStatType(t *testing.T) {
	in := []DataPoint{
		{XAxis: "A", Stats: []Stat{{Type: "amount", Value: 10}, {Type: "tax", Value: 1}}},
		{XAxis: "A", Stats: []Stat{{Type: "amount", Value: 20}, {Type: "tax", Value: 3}}},
	}

	out := AggregateDataPoints(in)

	if len(out) != 1 {
		t.Fatalf("expected 1 group, got %d", len(out))
	}
	if v, _ := statVal(out[0].Stats, "amount"); v != 30 {
		t.Fatalf("amount: want 30, got %v", v)
	}
	if v, _ := statVal(out[0].Stats, "tax"); v != 4 {
		t.Fatalf("tax: want 4, got %v", v)
	}
}

func TestAggregateDataPoints_DoesNotMutateInput(t *testing.T) {
	in := []DataPoint{
		{XAxis: "A", Stats: []Stat{{Type: "v", Value: 5}}},
		{XAxis: "A", Stats: []Stat{{Type: "v", Value: 7}}},
	}

	AggregateDataPoints(in)

	if in[0].Stats[0].Value != 5 {
		t.Fatalf("input mutated: first point value changed to %v", in[0].Stats[0].Value)
	}
}
