// Package sound plays SATurn's five short sounds with /usr/bin/afplay.
// Playback happens in Go, not in the webview, because WebKit may block audio
// that a click did not start (completion by hotkey or notification).
package sound

import (
	"embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

//go:embed wav/*.wav
var files embed.FS

var (
	mu  sync.Mutex
	dir string
)

// Names of the sounds.
const (
	Tick    = "tick"
	Quarter = "quarter"
	Done    = "done"
	Remind  = "remind"
	Alarm   = "alarm"
)

// Init writes the sounds to the user cache folder once.
func Init() error {
	base, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	d := filepath.Join(base, "SATurn", "sounds")
	if err := os.MkdirAll(d, 0o755); err != nil {
		return err
	}
	entries, err := files.ReadDir("wav")
	if err != nil {
		return err
	}
	for _, e := range entries {
		b, err := files.ReadFile("wav/" + e.Name())
		if err != nil {
			return err
		}
		p := filepath.Join(d, e.Name())
		if st, err := os.Stat(p); err == nil && st.Size() == int64(len(b)) {
			continue
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			return err
		}
	}
	mu.Lock()
	dir = d
	mu.Unlock()
	return nil
}

// Play starts a sound and returns at once.
func Play(name string) {
	mu.Lock()
	d := dir
	mu.Unlock()
	if d == "" {
		return
	}
	cmd := exec.Command("/usr/bin/afplay", filepath.Join(d, name+".wav"))
	if err := cmd.Start(); err != nil {
		log.Printf("sound error name=%s err=%q", name, err)
		return
	}
	go func() { _ = cmd.Wait() }()
}
