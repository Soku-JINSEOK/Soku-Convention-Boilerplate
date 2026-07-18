package manifest

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var (
	// ErrNotInitialized means neither a manifest nor pending write exists.
	ErrNotInitialized = errors.New("soku manifest is not initialized")
	// ErrRecoveryRequired means a valid pending write requires an explicit decision.
	ErrRecoveryRequired = errors.New("manifest recovery is required")
)

// Store owns durable manifest I/O within one repository root.
type Store struct {
	root string
}

// NewStore binds manifest storage to root.
func NewStore(root string) *Store {
	return &Store{root: root}
}

// Load validates durable state and reports pending writes without changing them.
func (s *Store) Load() (Document, error) {
	manifestData, manifestErr := os.ReadFile(s.manifestPath())
	pendingData, pendingErr := os.ReadFile(s.pendingPath())
	manifestExists := manifestErr == nil
	pendingExists := pendingErr == nil
	if manifestErr != nil && !errors.Is(manifestErr, fs.ErrNotExist) {
		return Document{}, fmt.Errorf("read manifest: %w", manifestErr)
	}
	if pendingErr != nil && !errors.Is(pendingErr, fs.ErrNotExist) {
		return Document{}, fmt.Errorf("read pending manifest: %w", pendingErr)
	}
	if !manifestExists && !pendingExists {
		return Document{}, ErrNotInitialized
	}
	var manifestDocument Document
	var manifestDecodeErr error
	if manifestExists {
		manifestDocument, manifestDecodeErr = Decode(manifestData)
	}
	if pendingExists {
		if manifestDecodeErr != nil {
			return Document{}, fmt.Errorf("ambiguous manifest state: durable manifest is invalid: %v", manifestDecodeErr)
		}
		if _, err := Decode(pendingData); err != nil {
			return Document{}, fmt.Errorf("validate pending manifest: %v", err)
		}
		return Document{}, ErrRecoveryRequired
	}
	if manifestDecodeErr != nil {
		return Document{}, fmt.Errorf("validate manifest: %w", manifestDecodeErr)
	}
	return manifestDocument, nil
}

// Write durably replaces the manifest through a same-directory pending file.
func (s *Store) Write(document Document) error {
	data, err := MarshalCanonical(document)
	if err != nil {
		return err
	}
	stateDirectory := filepath.Dir(s.manifestPath())
	if err := os.MkdirAll(stateDirectory, 0o700); err != nil {
		return fmt.Errorf("create manifest directory: %w", err)
	}
	file, err := os.OpenFile(s.pendingPath(), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create pending manifest: %w", err)
	}
	closed := false
	defer func() {
		if !closed {
			_ = file.Close()
		}
	}()
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("write pending manifest: %w", err)
	}
	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync pending manifest: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close pending manifest: %w", err)
	}
	closed = true
	if err := replaceFile(s.pendingPath(), s.manifestPath()); err != nil {
		return fmt.Errorf("replace manifest: %w", err)
	}
	if err := syncDirectory(stateDirectory); err != nil {
		return fmt.Errorf("sync manifest directory: %w", err)
	}
	return nil
}

// Recover resolves only unambiguous, fully valid pending states.
func (s *Store) Recover() (Document, error) {
	manifestData, manifestErr := os.ReadFile(s.manifestPath())
	pendingData, pendingErr := os.ReadFile(s.pendingPath())
	manifestExists := manifestErr == nil
	pendingExists := pendingErr == nil
	if manifestErr != nil && !errors.Is(manifestErr, fs.ErrNotExist) {
		return Document{}, fmt.Errorf("read manifest: %w", manifestErr)
	}
	if pendingErr != nil && !errors.Is(pendingErr, fs.ErrNotExist) {
		return Document{}, fmt.Errorf("read pending manifest: %w", pendingErr)
	}
	if !pendingExists {
		return s.Load()
	}
	pending, err := Decode(pendingData)
	if err != nil {
		return Document{}, fmt.Errorf("validate pending manifest: %v", err)
	}
	if manifestExists {
		manifest, err := Decode(manifestData)
		if err != nil {
			return Document{}, fmt.Errorf("ambiguous manifest state: durable manifest is invalid: %v", err)
		}
		if err := os.Remove(s.pendingPath()); err != nil {
			return Document{}, fmt.Errorf("discard pending manifest: %w", err)
		}
		if err := syncDirectory(filepath.Dir(s.manifestPath())); err != nil {
			return Document{}, fmt.Errorf("sync manifest directory: %w", err)
		}
		return manifest, nil
	}
	if err := replaceFile(s.pendingPath(), s.manifestPath()); err != nil {
		return Document{}, fmt.Errorf("promote pending manifest: %w", err)
	}
	if err := syncDirectory(filepath.Dir(s.manifestPath())); err != nil {
		return Document{}, fmt.Errorf("sync manifest directory: %w", err)
	}
	return pending, nil
}

func (s *Store) manifestPath() string {
	return filepath.Join(s.root, filepath.FromSlash(ManifestPath))
}

func (s *Store) pendingPath() string {
	return filepath.Join(s.root, filepath.FromSlash(PendingPath))
}
