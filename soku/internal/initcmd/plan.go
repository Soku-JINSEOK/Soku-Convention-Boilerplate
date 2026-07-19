package initcmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

func preflight(root string, changes []Change) ([]Change, error) {
	if err := ensureNoState(root); err != nil {
		return nil, err
	}
	seenDisk := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(root, path)
		if relative == "." {
			return nil
		}
		relative = filepath.ToSlash(relative)
		if relative == ".git" || strings.HasPrefix(relative, ".git/") || relative == ".soku" || strings.HasPrefix(relative, ".soku/") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if err := manifest.ValidatePath(relative); err != nil {
			return fail(4, "path.invalid", "existing repository path: %v", err)
		}
		lower := strings.ToLower(relative)
		if previous, exists := seenDisk[lower]; exists && previous != relative {
			return fail(4, "path.conflict", "repository paths %q and %q collide by case", previous, relative)
		}
		seenDisk[lower] = relative
		return nil
	})
	if err != nil {
		if failure, ok := err.(*Failure); ok {
			return nil, failure
		}
		return nil, fail(4, "path.invalid", "inspect target repository: %v", err)
	}
	planned := append([]Change(nil), changes...)
	for index := range planned {
		change := &planned[index]
		if existingCase, exists := seenDisk[strings.ToLower(change.Path)]; exists && existingCase != change.Path {
			return nil, fail(4, "path.conflict", "planned path %q collides with existing path %q", change.Path, existingCase)
		}
		if err := ensureNoSymlink(root, change.Path); err != nil {
			return nil, err
		}
		current, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(change.Path)))
		if errors.Is(err, fs.ErrNotExist) {
			change.Action = "create"
			continue
		}
		if err != nil {
			return nil, fail(4, "path.conflict", "existing path %q is not a readable regular file", change.Path)
		}
		switch change.Path {
		case ".gitignore":
			change.Content, err = mergeGitignore(current, change.Content)
			change.Action = "merge"
		case ".editorconfig":
			change.Content, err = mergeEditorconfig(current, change.Content)
			change.Action = "merge"
		default:
			return nil, fail(4, "ownership.conflict", "existing path %q is project-owned; use soku status or upgrade instead", change.Path)
		}
		if err != nil {
			return nil, err
		}
		change.BaselineSHA256, err = manifest.HashContent(change.Content, change.ContentMode)
		if err != nil {
			return nil, err
		}
		if bytesEqualCanonical(current, change.Content) {
			change.Action = "unchanged"
		}
	}
	sort.Slice(planned, func(i, j int) bool { return planned[i].Path < planned[j].Path })
	return planned, nil
}

func ensureNoSymlink(root, relative string) error {
	current := root
	for _, component := range strings.Split(relative, "/") {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return fail(4, "path.invalid", "inspect %q: %v", relative, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fail(4, "path.conflict", "path %q traverses a symbolic link", relative)
		}
		if current != filepath.Join(root, filepath.FromSlash(relative)) && !info.IsDir() {
			return fail(4, "path.conflict", "parent of %q is not a directory", relative)
		}
	}
	return nil
}
func bytesEqualCanonical(a, b []byte) bool {
	ha, ea := manifest.HashContent(a, "text")
	hb, eb := manifest.HashContent(b, "text")
	return ea == nil && eb == nil && ha == hb
}
