package main

import (
	"log"

	"github.com/Publicthrone547/reminder_sat/internal/core"
)

// registerHotkeys registers the global shortcuts when the setting is on.
func (s *Shell) registerHotkeys() {
	if s.app == nil {
		return
	}
	_ = s.app.GlobalShortcut.UnregisterAll()
	s.mu.Lock()
	on := s.st.Settings.Hotkeys
	s.mu.Unlock()
	var errs []string
	if on {
		addKey := func(n int) func() {
			return func() {
				if activeIsTest(s.current()) {
					return
				}
				s.add(-1, n, "hotkey")
			}
		}
		keys := []struct {
			acc string
			fn  func()
		}{
			{"Ctrl+Option+1", addKey(1)},
			{"Ctrl+Option+5", addKey(5)},
			{"Ctrl+Option+0", addKey(10)},
			{"Ctrl+Option+Z", func() { _, _ = s.do("undo source=hotkey", wrap(core.Undo)) }},
			{"Ctrl+Option+S", func() { s.togglePopover() }},
		}
		for _, k := range keys {
			if err := s.app.GlobalShortcut.Register(k.acc, k.fn); err != nil {
				log.Printf("hotkey %s err=%q", k.acc, err)
				errs = append(errs, k.acc+": "+err.Error())
			} else {
				log.Printf("hotkey %s registered", k.acc)
			}
		}
	} else {
		log.Printf("hotkeys off")
	}
	s.setSys(func(v *core.SystemView) { v.HotkeyErrors = append([]string{}, errs...) })
}
