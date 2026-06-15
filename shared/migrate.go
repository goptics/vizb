package shared

import "encoding/json"

// MigrateDataset populates ds.Meta (and each HistoryEntry.Meta) from the
// legacy flat top-level fields (cpu/os/arch/pkg) that existed before the Meta
// struct was introduced. rawJSON must be the original bytes from which ds was
// unmarshalled; pass nil to skip migration.
func MigrateDataset(ds *Dataset, rawJSON []byte) {
	if len(rawJSON) == 0 {
		return
	}

	var legacy struct {
		CPU     *CPUInfo `json:"cpu"`
		OS      string   `json:"os"`
		Arch    string   `json:"arch"`
		Pkg     string   `json:"pkg"`
		History []struct {
			CPU *CPUInfo `json:"cpu"`
			OS  string   `json:"os"`
		} `json:"history"`
	}
	if err := json.Unmarshal(rawJSON, &legacy); err != nil {
		return
	}

	if ds.Meta == nil {
		m := &Meta{CPU: legacy.CPU, OS: legacy.OS, Arch: legacy.Arch, Pkg: legacy.Pkg}
		if m.CPU != nil || m.OS != "" || m.Arch != "" || m.Pkg != "" {
			ds.Meta = m
		}
	}

	for i := range ds.History {
		if i >= len(legacy.History) {
			break
		}
		if ds.History[i].Meta == nil {
			leg := legacy.History[i]
			m := &Meta{CPU: leg.CPU, OS: leg.OS}
			if m.CPU != nil || m.OS != "" {
				ds.History[i].Meta = m
			}
		}
	}
}
