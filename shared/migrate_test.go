package shared

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
)

const baseSettings = `"settings":{"charts":["bar"],"sort":{"enabled":false,"order":"asc"},"showLabels":false,"scale":"linear"}`

type MigrateSuite struct {
	suite.Suite
}

func (s *MigrateSuite) TestMigrateDataset() {
	s.Run("migrates legacy flat cpu/os/arch/pkg to Meta", func() {
		raw := []byte(`{"name":"test","cpu":{"name":"Intel i7","cores":8},"os":"linux","arch":"amd64","pkg":"github.com/foo/bar",` + baseSettings + `,"data":[]}`)
		var ds Dataset
		s.Require().NoError(json.Unmarshal(raw, &ds))
		s.Require().Nil(ds.Meta, "expected nil Meta before migration")
		MigrateDataset(&ds, raw)
		s.Require().NotNil(ds.Meta, "expected Meta to be populated after migration")
		s.Require().NotNil(ds.Meta.CPU)
		s.Equal("Intel i7", ds.Meta.CPU.Name)
		s.Equal(8, ds.Meta.CPU.Cores)
		s.Equal("linux", ds.Meta.OS)
		s.Equal("amd64", ds.Meta.Arch)
		s.Equal("github.com/foo/bar", ds.Meta.Pkg)
	})

	s.Run("migrates legacy cpu/os on history entries", func() {
		raw := []byte(`{"name":"test","history":[{"tag":"v1","timestamp":"2025-01-01T00:00:00Z","cpu":{"name":"AMD Ryzen","cores":4},"os":"darwin"}],` + baseSettings + `,"data":[]}`)
		var ds Dataset
		s.Require().NoError(json.Unmarshal(raw, &ds))
		MigrateDataset(&ds, raw)
		s.Require().Len(ds.History, 1)
		h := ds.History[0]
		s.Require().NotNil(h.Meta, "expected history entry Meta to be populated")
		s.Require().NotNil(h.Meta.CPU)
		s.Equal("AMD Ryzen", h.Meta.CPU.Name)
		s.Equal(4, h.Meta.CPU.Cores)
		s.Equal("darwin", h.Meta.OS)
	})

	s.Run("skips migration when Meta already present", func() {
		raw := []byte(`{"name":"test","cpu":{"name":"Old CPU","cores":2},"meta":{"cpu":{"name":"New CPU","cores":16},"os":"windows"},` + baseSettings + `,"data":[]}`)
		var ds Dataset
		s.Require().NoError(json.Unmarshal(raw, &ds))
		MigrateDataset(&ds, raw)
		s.Require().NotNil(ds.Meta)
		s.Require().NotNil(ds.Meta.CPU)
		s.Equal("New CPU", ds.Meta.CPU.Name)
	})

	s.Run("skips history entry migration when Meta already present", func() {
		raw := []byte(`{"name":"test","history":[{"tag":"v1","timestamp":"2025-01-01T00:00:00Z","cpu":{"name":"Old"},"meta":{"cpu":{"name":"Keep"},"os":"linux"}}],` + baseSettings + `,"data":[]}`)
		var ds Dataset
		s.Require().NoError(json.Unmarshal(raw, &ds))
		MigrateDataset(&ds, raw)
		s.Require().NotNil(ds.History[0].Meta)
		s.Require().NotNil(ds.History[0].Meta.CPU)
		s.Equal("Keep", ds.History[0].Meta.CPU.Name)
	})

	s.Run("no migration when no legacy fields and no meta", func() {
		raw := []byte(`{"name":"test",` + baseSettings + `,"data":[]}`)
		var ds Dataset
		s.Require().NoError(json.Unmarshal(raw, &ds))
		MigrateDataset(&ds, raw)
		s.Nil(ds.Meta)
	})

	s.Run("skips migration when rawJSON is nil", func() {
		ds := &Dataset{}
		MigrateDataset(ds, nil)
		s.Nil(ds.Meta)
	})

	s.Run("mixed history: only migrates entries with nil Meta", func() {
		raw := []byte(`{"name":"test","history":[` +
			`{"tag":"v1","timestamp":"2025-01-01T00:00:00Z","cpu":{"name":"Old","cores":2}},` +
			`{"tag":"v2","timestamp":"2025-06-01T00:00:00Z","meta":{"cpu":{"name":"New","cores":8},"os":"linux"}}` +
			`],` + baseSettings + `,"data":[]}`)
		var ds Dataset
		s.Require().NoError(json.Unmarshal(raw, &ds))
		MigrateDataset(&ds, raw)
		s.Require().NotNil(ds.History[0].Meta)
		s.Require().NotNil(ds.History[0].Meta.CPU)
		s.Equal("Old", ds.History[0].Meta.CPU.Name)
		s.Require().NotNil(ds.History[1].Meta)
		s.Require().NotNil(ds.History[1].Meta.CPU)
		s.Equal("New", ds.History[1].Meta.CPU.Name)
	})
}

func TestMigrateSuite(t *testing.T) {
	suite.Run(t, new(MigrateSuite))
}
