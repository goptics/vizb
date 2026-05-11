package shared

import "sort"

func MergeRunIntoHistory(history *History, run Run, pruneCount int) {
	if history.Runs == nil {
		history.Runs = []Run{}
	}
	for _, existing := range history.Runs {
		if existing.Version == run.Version {
			return
		}
	}
	history.Runs = append([]Run{run}, history.Runs...)
	sort.Slice(history.Runs, func(i, j int) bool {
		return history.Runs[i].Date.After(history.Runs[j].Date)
	})
	if pruneCount > 0 && len(history.Runs) > pruneCount {
		history.Runs = history.Runs[:pruneCount]
	}
}
