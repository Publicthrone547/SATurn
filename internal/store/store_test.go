package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Publicthrone547/reminder_sat/internal/core"
)

func TestRoundTrip(t *testing.T) {
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	st, err := s.Load()
	if err != nil || st.Settings.ExamDate != core.DefaultExamDate {
		t.Fatalf("new state: %v %+v", err, st.Settings)
	}
	st.FirstDay = "2026-11-24"
	if err := s.Save(st); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(s.Path)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", info.Mode())
	}
	s2, _ := Open(s.Dir)
	back, err := s2.Load()
	if err != nil || back.FirstDay != "2026-11-24" {
		t.Fatalf("%v %+v", err, back)
	}
}

func TestMinimalFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "state.json"), []byte(`{ "version": 1, "settings": { "onboarded": true, "examDate": "2026-12-05" } }`), 0o600)
	s, _ := Open(dir)
	st, err := s.Load()
	if err != nil || !st.Settings.Onboarded || st.Settings.RemindUntil != 1380 || st.Days == nil {
		t.Fatalf("%v %+v", err, st.Settings)
	}
}

func TestBrokenFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "state.json"), []byte("{nope"), 0o600)
	s, _ := Open(dir)
	st, err := s.Load()
	if err != nil || st == nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "state.broken-") {
			found = true
		}
	}
	if !found {
		t.Fatal("broken file not set aside")
	}
}
