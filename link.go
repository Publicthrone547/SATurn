package main

import (
	"log"
	"net/url"
	"strconv"
	"strings"
)

// openLink runs a saturn:// link, so the Shortcuts app, Raycast, a Stream
// Deck or `open -g saturn://add` can log progress without a hotkey:
//
//	saturn://add            +1
//	saturn://add?n=5        +5 (1 to 200)
//	saturn://add/5          the same
//	saturn://undo           undo the last entry
//	saturn://popover        open or close the popover
//	saturn://open           the main window
func (s *Shell) openLink(raw string) {
	cmd, arg, ok := parseLink(raw)
	if !ok {
		log.Printf("link ignored %q", raw)
		return
	}
	log.Printf("link %s arg=%q", cmd, arg)
	switch cmd {
	case "add":
		n := 1
		if arg != "" {
			v, err := strconv.Atoi(arg)
			if err != nil || v < 1 || v > 200 {
				log.Printf("link add: bad count %q", arg)
				return
			}
			n = v
		}
		s.quickAdd(n, "link")
	case "undo", "popover":
		s.hotkeyAction(cmd, "link")
	case "open":
		s.openMain("")
	default:
		log.Printf("link: unknown command %q", cmd)
	}
}

// parseLink splits a saturn:// link into its command and argument.
func parseLink(raw string) (cmd, arg string, ok bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "saturn" {
		return "", "", false
	}
	// saturn://add/5 puts "add" in the host; saturn:add/5 puts it in Opaque.
	parts := strings.Split(strings.Trim(u.Host+"/"+u.Opaque+u.Path, "/"), "/")
	cmd = strings.ToLower(parts[0])
	if len(parts) > 1 {
		arg = parts[len(parts)-1]
	}
	if q := u.Query().Get("n"); q != "" {
		arg = q
	}
	return cmd, arg, cmd != ""
}
