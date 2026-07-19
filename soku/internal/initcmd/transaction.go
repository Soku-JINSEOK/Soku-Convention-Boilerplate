package initcmd

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Soku-JINSEOK/Soku-Convention-Boilerplate/soku/internal/manifest"
)

type transactionRecord struct {
	ID              string            `json:"id"`
	State           string            `json:"state"`
	ManifestExisted bool              `json:"manifest_existed"`
	Paths           []transactionPath `json:"paths"`
}
type transactionPath struct {
	Path    string `json:"path"`
	Existed bool   `json:"existed"`
	Mode    uint32 `json:"mode,omitempty"`
}
type ApplyHook func(stage, path string) error

func applyTransaction(root string, changes []Change, document manifest.Document, hook ApplyHook) (string, error) {
	id, err := transactionID()
	if err != nil {
		return "", err
	}
	directory := filepath.Join(root, ".soku", "transactions", id)
	backupRoot := filepath.Join(directory, "backup")
	if err := os.MkdirAll(backupRoot, 0o700); err != nil {
		return id, fail(7, "apply.rolled_back", "create transaction: %v", err)
	}
	record := transactionRecord{ID: id, State: "prepared"}
	manifestPath := filepath.Join(root, filepath.FromSlash(manifest.ManifestPath))
	if data, readErr := os.ReadFile(manifestPath); readErr == nil {
		backup := filepath.Join(backupRoot, filepath.FromSlash(manifest.ManifestPath))
		if err := os.MkdirAll(filepath.Dir(backup), 0o700); err != nil {
			_ = os.RemoveAll(directory)
			return id, fail(7, "apply.rolled_back", "backup previous manifest: %v", err)
		}
		if err := os.WriteFile(backup, data, 0o600); err != nil {
			_ = os.RemoveAll(directory)
			return id, fail(7, "apply.rolled_back", "backup previous manifest: %v", err)
		}
		record.ManifestExisted = true
	} else if !errors.Is(readErr, fs.ErrNotExist) {
		return rollbackResult(root, directory, record, readErr, hook)
	}
	active := make([]Change, 0, len(changes))
	for _, change := range changes {
		if change.Action != "unchanged" {
			active = append(active, change)
		}
	}
	for _, change := range active {
		target := filepath.Join(root, filepath.FromSlash(change.Path))
		info, statErr := os.Stat(target)
		item := transactionPath{Path: change.Path}
		if statErr == nil {
			item.Existed = true
			item.Mode = uint32(info.Mode().Perm())
			data, readErr := os.ReadFile(target)
			if readErr != nil {
				return rollbackResult(root, directory, record, fmt.Errorf("backup %s: %w", change.Path, readErr), hook)
			}
			backup := filepath.Join(backupRoot, filepath.FromSlash(change.Path))
			if err := os.MkdirAll(filepath.Dir(backup), 0o700); err != nil {
				return rollbackResult(root, directory, record, err, hook)
			}
			if err := os.WriteFile(backup, data, 0o600); err != nil {
				return rollbackResult(root, directory, record, err, hook)
			}
		} else if !errors.Is(statErr, fs.ErrNotExist) {
			return rollbackResult(root, directory, record, statErr, hook)
		}
		record.Paths = append(record.Paths, item)
	}
	sort.Slice(record.Paths, func(i, j int) bool { return record.Paths[i].Path < record.Paths[j].Path })
	if err := writeJournal(directory, record); err != nil {
		return rollbackResult(root, directory, record, err, hook)
	}
	for _, change := range active {
		if hook != nil {
			if err := hook("before-write", change.Path); err != nil {
				return rollbackResult(root, directory, record, err, hook)
			}
		}
		target := filepath.Join(root, filepath.FromSlash(change.Path))
		if change.Action == "removed" || change.Action == "remove" {
			if err := os.Remove(target); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return rollbackResult(root, directory, record, err, hook)
			}
			removeEmptyParents(filepath.Dir(target), root)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return rollbackResult(root, directory, record, err, hook)
		}
		temporary := target + ".soku-tmp-" + id
		if err := os.WriteFile(temporary, change.Content, 0o644); err != nil {
			return rollbackResult(root, directory, record, err, hook)
		}
		if err := os.Rename(temporary, target); err != nil {
			_ = os.Remove(temporary)
			return rollbackResult(root, directory, record, err, hook)
		}
		data, err := os.ReadFile(target)
		if err != nil {
			return rollbackResult(root, directory, record, err, hook)
		}
		hash, err := manifest.HashContent(data, change.ContentMode)
		if err != nil || hash != change.BaselineSHA256 {
			return rollbackResult(root, directory, record, errors.New("applied content hash mismatch"), hook)
		}
	}
	if hook != nil {
		if err := hook("before-manifest", manifest.ManifestPath); err != nil {
			return rollbackResult(root, directory, record, err, hook)
		}
	}
	if err := manifest.NewStore(root).Write(document); err != nil {
		return rollbackResult(root, directory, record, err, hook)
	}
	record.State = "committed"
	_ = writeJournal(directory, record)
	if err := os.RemoveAll(directory); err != nil {
		return id, fail(8, "recovery.required", "manifest committed but transaction cleanup failed; preserve %s", filepath.ToSlash(directory))
	}
	_ = os.Remove(filepath.Join(root, ".soku", "transactions"))
	return id, nil
}

func rollbackResult(root, directory string, record transactionRecord, applyErr error, hook ApplyHook) (string, error) {
	id := record.ID
	if hook != nil {
		if err := hook("before-rollback", ""); err != nil {
			return id, fail(8, "rollback.failed", "apply failed (%v) and rollback injection failed (%v); preserve .soku/transactions/%s", applyErr, err, id)
		}
	}
	for index := len(record.Paths) - 1; index >= 0; index-- {
		item := record.Paths[index]
		target := filepath.Join(root, filepath.FromSlash(item.Path))
		if item.Existed {
			data, err := os.ReadFile(filepath.Join(directory, "backup", filepath.FromSlash(item.Path)))
			if err != nil {
				return id, fail(8, "rollback.failed", "apply failed and backup %q cannot be read; preserve .soku/transactions/%s", item.Path, id)
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return id, fail(8, "rollback.failed", "apply failed and rollback cannot create parent; preserve .soku/transactions/%s", id)
			}
			if err := os.WriteFile(target, data, fs.FileMode(item.Mode)); err != nil {
				return id, fail(8, "rollback.failed", "apply failed and rollback cannot restore %q; preserve .soku/transactions/%s", item.Path, id)
			}
		} else if err := os.Remove(target); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return id, fail(8, "rollback.failed", "apply failed and rollback cannot remove %q; preserve .soku/transactions/%s", item.Path, id)
		} else if err == nil {
			removeEmptyParents(filepath.Dir(target), root)
		}
	}
	_ = os.Remove(filepath.Join(root, filepath.FromSlash(manifest.PendingPath)))
	manifestPath := filepath.Join(root, filepath.FromSlash(manifest.ManifestPath))
	if record.ManifestExisted {
		data, err := os.ReadFile(filepath.Join(directory, "backup", filepath.FromSlash(manifest.ManifestPath)))
		if err != nil {
			return id, fail(8, "rollback.failed", "apply failed and previous manifest cannot be read; preserve .soku/transactions/%s", id)
		}
		if err := os.MkdirAll(filepath.Dir(manifestPath), 0o700); err != nil {
			return id, fail(8, "rollback.failed", "apply failed and previous manifest directory cannot be restored; preserve .soku/transactions/%s", id)
		}
		if err := os.WriteFile(manifestPath, data, 0o600); err != nil {
			return id, fail(8, "rollback.failed", "apply failed and previous manifest cannot be restored; preserve .soku/transactions/%s", id)
		}
	} else {
		_ = os.Remove(manifestPath)
	}
	if err := os.RemoveAll(directory); err != nil {
		return id, fail(8, "rollback.failed", "apply failed and rollback cleanup failed; preserve .soku/transactions/%s", id)
	}
	return id, fail(7, "apply.rolled_back", "apply failed and rollback restored the previous state: %v", applyErr)
}

func removeEmptyParents(directory, root string) {
	for directory != root && directory != filepath.Join(root, ".soku") {
		if err := os.Remove(directory); err != nil {
			return
		}
		directory = filepath.Dir(directory)
	}
}
func writeJournal(directory string, record transactionRecord) error {
	recordData, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	recordData = append(recordData, '\n')
	path := filepath.Join(directory, "journal.json")
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err = file.Write(recordData); err != nil {
		_ = file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}
func transactionID() (string, error) {
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return time.Now().UTC().Format("20060102T150405Z") + "-" + hex.EncodeToString(random), nil
}
