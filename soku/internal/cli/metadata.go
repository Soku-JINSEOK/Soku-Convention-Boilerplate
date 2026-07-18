package cli

import (
	"runtime/debug"
	"strings"
)

// These values are populated by release builds through controlled ldflags.
var (
	Version = ""
	Commit  = ""
	BuiltAt = ""
)

// BuildMetadata is the public version identity of this executable.
type BuildMetadata struct {
	Version string
	Commit  string
	BuiltAt string
}

type buildInfoReader func() (*debug.BuildInfo, bool)

func resolveBuildMetadata() BuildMetadata {
	return resolveBuildMetadataWith(debug.ReadBuildInfo)
}

func resolveBuildMetadataWith(read buildInfoReader) BuildMetadata {
	metadata := BuildMetadata{
		Version: strings.TrimSpace(Version),
		Commit:  strings.TrimSpace(Commit),
		BuiltAt: strings.TrimSpace(BuiltAt),
	}

	if info, ok := read(); ok && info != nil {
		if metadata.Version == "" && info.Main.Version != "" && info.Main.Version != "(devel)" {
			metadata.Version = info.Main.Version
		}
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if metadata.Commit == "" {
					metadata.Commit = setting.Value
				}
			case "vcs.time":
				if metadata.BuiltAt == "" {
					metadata.BuiltAt = setting.Value
				}
			}
		}
	}

	if metadata.Version == "" || metadata.Version == "(devel)" {
		metadata.Version = "dev"
	} else if metadata.Version != "dev" && !strings.HasPrefix(metadata.Version, "v") {
		metadata.Version = "v" + metadata.Version
	}
	if metadata.Commit == "" {
		metadata.Commit = "unknown"
	}
	if metadata.BuiltAt == "" {
		metadata.BuiltAt = "unknown"
	}
	return metadata
}
