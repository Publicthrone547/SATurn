package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Publicthrone547/SATurn/internal/core"
)

const launchAgentPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
  <key>Label</key><string>io.github.publicthrone547.saturn</string>
  <key>ProgramArguments</key><array><string>/Applications/SATurn.app/Contents/MacOS/SATurn</string></array>
  <key>RunAtLoad</key><true/>
  <key>KeepAlive</key><false/>
  <key>LimitLoadToSessionType</key><string>Aqua</string>
  <key>ProcessType</key><string>Interactive</string>
</dict></plist>
`

// syncAutostart reads the login registration into the system view.
func (s *Shell) syncAutostart() {
	if s.env.quiet() {
		note := "Only the copy in Applications starts at login"
		if s.env.dry {
			note = "Dry run: launch at login is not changed"
		}
		s.setSys(func(v *core.SystemView) { v.Autostart, v.AutostartNote = false, note })
		log.Printf("autostart skipped installed=%v dry=%v", s.env.installed, s.env.dry)
		return
	}
	st, err := s.app.Autostart.Status()
	note := ""
	if err != nil {
		note = "Could not read the login item: " + err.Error()
	} else if st.Enabled {
		note = "Starts at login (" + string(st.Strategy) + ")"
	} else {
		note = "Does not start at login"
	}
	log.Printf("autostart enabled=%v strategy=%s path=%s err=%v", st.Enabled, st.Strategy, st.Path, err)
	s.setSys(func(v *core.SystemView) { v.Autostart, v.AutostartNote = st.Enabled, note })
}

// setLaunchAtLogin registers or removes the login item.
func (s *Shell) setLaunchAtLogin(on bool) {
	if s.env.quiet() {
		log.Printf("autostart set=%v skipped (not installed or dry)", on)
		s.syncAutostart()
		return
	}
	if on {
		if err := s.app.Autostart.Enable(); err != nil {
			log.Printf("autostart enable err=%q; writing a LaunchAgent", err)
			if err := writeLaunchAgent(); err != nil {
				log.Printf("autostart launchagent err=%q", err)
			}
		}
	} else if err := s.app.Autostart.Disable(); err != nil {
		log.Printf("autostart disable err=%q", err)
	}
	s.syncAutostart()
}

func writeLaunchAgent() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, "Library", "LaunchAgents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	p := filepath.Join(dir, bundleID+".plist")
	log.Printf("autostart launchagent path=%s", p)
	return os.WriteFile(p, []byte(launchAgentPlist), 0o644)
}

// homeRel shortens paths under $HOME to ~/… for display.
func homeRel(p string) string {
	home, err := os.UserHomeDir()
	if err == nil && home != "" && strings.HasPrefix(p, home) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}
