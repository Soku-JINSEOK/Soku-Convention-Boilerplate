package cli

import (
	"runtime/debug"
	"testing"
)

func TestResolveBuildMetadata(t *testing.T) {
	previousVersion, previousCommit, previousBuiltAt := Version, Commit, BuiltAt
	t.Cleanup(func() {
		Version, Commit, BuiltAt = previousVersion, previousCommit, previousBuiltAt
	})

	tests := []struct {
		name    string
		version string
		commit  string
		builtAt string
		info    *debug.BuildInfo
		ok      bool
		want    BuildMetadata
	}{
		{
			name:    "ldflags take precedence",
			version: "v1.2.3",
			commit:  "ldflags-commit",
			builtAt: "ldflags-time",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}, Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "build-commit"},
				{Key: "vcs.time", Value: "build-time"},
			}},
			ok:   true,
			want: BuildMetadata{Version: "v1.2.3", Commit: "ldflags-commit", BuiltAt: "ldflags-time"},
		},
		{
			name: "build info fallback",
			info: &debug.BuildInfo{Main: debug.Module{Version: "1.4.0"}, Settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "build-commit"},
				{Key: "vcs.time", Value: "build-time"},
			}},
			ok:   true,
			want: BuildMetadata{Version: "v1.4.0", Commit: "build-commit", BuiltAt: "build-time"},
		},
		{
			name: "development fallback",
			info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:   true,
			want: BuildMetadata{Version: "dev", Commit: "unknown", BuiltAt: "unknown"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Version, Commit, BuiltAt = test.version, test.commit, test.builtAt
			got := resolveBuildMetadataWith(func() (*debug.BuildInfo, bool) { return test.info, test.ok })
			if got != test.want {
				t.Fatalf("metadata=%#v want=%#v", got, test.want)
			}
		})
	}
}
