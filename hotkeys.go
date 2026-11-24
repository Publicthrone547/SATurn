package main

import (
	"log"
	"strings"
	"time"

	"github.com/Publicthrone547/SATurn/internal/core"
)

// hotkeyAction runs one hotkey action ("add1", "undo", ...) from outside the
// windows: a global hotkey or a saturn:// link.
func (s *Shell) hotkeyAction(action, source string) {
	switch action {
	case "add1":
		s.quickAdd(1, source)
	case "add5":
		s.quickAdd(5, source)
	case "add10":
		s.quickAdd(10, source)
	case "undo":
		_, _ = s.do("undo source="+source, wrap(core.Undo))
	case "popover":
		s.togglePopover()
	}
}

// registerHotkeys registers the global shortcuts when the setting is on.
func (s *Shell) registerHotkeys() {
	if s.app == nil {
		return
	}
	_ = s.app.GlobalShortcut.UnregisterAll()
	s.mu.Lock()
	paused := s.keysPaused
	on := s.st.Settings.Hotkeys && !paused
	keys := s.st.Settings.Keys
	s.mu.Unlock()
	var errs []string
	if on {
		used := map[string]string{}
		for _, action := range core.HotkeyActions {
			acc := keys[action]
			if acc == "" {
				continue
			}
			norm := strings.ToLower(acc)
			if other, dup := used[norm]; dup {
				errs = append(errs, acc+": already used by "+other)
				continue
			}
			used[norm] = action
			if err := s.app.GlobalShortcut.Register(acc, func() { s.hotkeyAction(action, "hotkey") }); err != nil {
				log.Printf("hotkey %s=%s err=%q", action, acc, err)
				errs = append(errs, acc+": "+err.Error())
			} else {
				log.Printf("hotkey %s=%s registered", action, acc)
			}
		}
	} else {
		log.Printf("hotkeys off paused=%v", paused)
	}
	s.setSys(func(v *core.SystemView) { v.HotkeyErrors = append([]string{}, errs...) })
}

// pauseHotkeys releases the global shortcuts while the settings page records
// a new one, so the current combination reaches the recorder too.
func (s *Shell) pauseHotkeys(paused bool) {
	s.mu.Lock()
	same := s.keysPaused == paused
	s.keysPaused = paused
	s.mu.Unlock()
	if same {
		return
	}
	s.registerHotkeys()
	if paused {
		// A recorder that never reports back (its window closed) must not
		// leave the hotkeys off for good.
		time.AfterFunc(30*time.Second, func() { s.pauseHotkeys(false) })
	}
}
