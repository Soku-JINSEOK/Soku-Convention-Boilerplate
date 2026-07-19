package initcmd

import (
	"encoding/json"
	"sort"
	"strings"
)

func DecodeProfileIndex(data []byte) (ProfileIndex, error) {
	var index ProfileIndex
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&index); err != nil {
		return ProfileIndex{}, fail(5, "profile.incompatible", "decode profile index: %v", err)
	}
	if index.SchemaVersion != 2 || index.DefaultProfile != ProfileStandard || len(index.Profiles) != 3 || len(index.Layers) != 3 {
		return ProfileIndex{}, fail(5, "profile.incompatible", "profile index must define catalog v2 bootstrap, standard, and scaled composition")
	}
	wantIDs := []string{ProfileBootstrap, ProfileStandard, ProfileScaled}
	for position, id := range wantIDs {
		profile := index.Profiles[position]
		layer := index.Layers[position]
		if profile.ID != id || layer.ID != id || strings.Join(profile.Layers, "\x00") != strings.Join(wantIDs[:position+1], "\x00") {
			return ProfileIndex{}, fail(5, "profile.incompatible", "profiles must compose bootstrap, standard, and scaled in fixed order")
		}
		if layer.StackFileLimit < -1 {
			return ProfileIndex{}, fail(5, "profile.incompatible", "profile layer %q has an invalid stack file limit", id)
		}
		for _, file := range layer.Files {
			if file.Owner != "core" || file.Class != "core-managed" || file.Strategy != "render" || (file.ContentMode != "text" && file.ContentMode != "binary") {
				return ProfileIndex{}, fail(5, "profile.incompatible", "profile layer %q has an invalid file declaration", id)
			}
			if err := validateOutputPath(file.Source, false); err != nil {
				return ProfileIndex{}, fail(5, "profile.incompatible", "profile source %q is invalid", file.Source)
			}
			if err := validateOutputPath(file.Output, false); err != nil {
				return ProfileIndex{}, fail(5, "profile.incompatible", "profile output %q is invalid", file.Output)
			}
		}
	}
	return index, nil
}

func renderProfileCatalog(snapshot SourceSnapshot, catalog Catalog, config Config) ([]Change, error) {
	data, hasIndex := snapshot.Files[ProfileIndexPath]
	if !hasIndex {
		if config.Profile != ProfileStandard {
			return nil, fail(5, "profile.incompatible", "legacy core-v1 sources support only the standard profile")
		}
		return renderCatalog(snapshot, catalog, config)
	}
	index, err := DecodeProfileIndex(data)
	if err != nil {
		return nil, err
	}
	var selected *Profile
	for position := range index.Profiles {
		if index.Profiles[position].ID == config.Profile {
			selected = &index.Profiles[position]
			break
		}
	}
	if selected == nil {
		return nil, fail(5, "profile.incompatible", "profile %q is not declared by the source", config.Profile)
	}
	shared := map[string]bool{}
	stackLimit := 0
	extra := []CatalogFile{}
	for _, layerID := range selected.Layers {
		for _, layer := range index.Layers {
			if layer.ID != layerID {
				continue
			}
			for _, output := range layer.SharedOutputs {
				shared[output] = true
			}
			if layer.StackFileLimit == -1 || layer.StackFileLimit > stackLimit {
				stackLimit = layer.StackFileLimit
			}
			extra = append(extra, layer.Files...)
		}
	}
	composed := catalog
	composed.Files = nil
	seen := map[string]bool{}
	for _, declaration := range catalog.Files {
		if shared[declaration.Output] {
			composed.Files = append(composed.Files, declaration)
			seen[strings.ToLower(declaration.Output)] = true
		}
	}
	for _, declaration := range extra {
		folded := strings.ToLower(declaration.Output)
		if seen[folded] {
			return nil, fail(4, "ownership.conflict", "profile output %q collides with another core output", declaration.Output)
		}
		seen[folded] = true
		composed.Files = append(composed.Files, declaration)
	}
	for position := range composed.Stacks {
		files := catalog.Stacks[position].Files
		if stackLimit >= 0 && stackLimit < len(files) {
			files = files[:stackLimit]
		}
		composed.Stacks[position].Files = append([]CatalogFile(nil), files...)
	}
	sort.Slice(composed.Files, func(i, j int) bool { return composed.Files[i].Output < composed.Files[j].Output })
	return renderCatalog(snapshot, composed, config)
}
