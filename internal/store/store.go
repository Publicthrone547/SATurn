// Package store loads and saves SATurn's state as one JSON file, atomically.
package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Publicthrone547/SATurn/internal/core"
)

type Store struct {
	Dir  string
	Path string
	last []byte
}

// DefaultDir is $SATURN_DATA_DIR, else ~/Library/Application Support/SATurn.
func DefaultDir() (string, error) {
	if d := os.Getenv("SATURN_DATA_DIR"); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Application Support", "SATurn"), nil
}

// Open prepares the folder (mode 0700).
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Store{Dir: dir, Path: filepath.Join(dir, "state.json")}, nil
}

// Load reads the state. A missing file gives a new state; a file that does
// not parse is set aside as state.broken-<unix>.json and a new state returned.
func (s *Store) Load() (*core.State, error) {
	b, err := os.ReadFile(s.Path)
	if os.IsNotExist(err) {
		return core.NewState(), nil
	}
	if err != nil {
		return nil, err
	}
	st := core.NewState()
	if err := json.Unmarshal(b, st); err != nil {
		broken := filepath.Join(s.Dir, fmt.Sprintf("state.broken-%d.json", time.Now().Unix()))
		log.Printf("store unreadable state file=%s err=%q moved=%s", s.Path, err, broken)
		_ = os.Rename(s.Path, broken)
		return core.NewState(), nil
	}
	st.Fix()
	s.last = b
	return st, nil
}

// Save writes the state atomically (temporary file, fsync, rename). It does
// nothing when the content has not changed since the last save.
func (s *Store) Save(st *core.State) error {
	b, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	if bytes.Equal(b, s.last) {
		return nil
	}
	tmp, err := os.CreateTemp(s.Dir, ".state-*.json")
	if err != nil {
		return err
	}
	name := tmp.Name()
	cleanup := func() { _ = os.Remove(name) }
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(name, s.Path); err != nil {
		cleanup()
		return err
	}
	s.last = b
	return nil
}
