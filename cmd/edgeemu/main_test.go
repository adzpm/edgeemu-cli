package main

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func buildInfo(moduleVersion string, ok bool) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		if !ok {
			return nil, false
		}

		return &debug.BuildInfo{Main: debug.Module{Version: moduleVersion}}, true
	}
}

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name    string
		stamped string
		module  string
		ok      bool
		want    string
	}{
		{"stamped release wins over build info", "0.4.0", "v9.9.9", true, "0.4.0"},
		{"go install reports the module version", devVersion, "v0.4.0", true, "v0.4.0"},
		{"plain go build without VCS stays dev", devVersion, "(devel)", true, devVersion},
		{"empty module version stays dev", devVersion, "", true, devVersion},
		{"no build info stays dev", devVersion, "", false, devVersion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveVersion(tt.stamped, buildInfo(tt.module, tt.ok)))
		})
	}
}
