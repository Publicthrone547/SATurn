package main

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/Publicthrone547/reminder_sat/internal/core"
)

type IslandButton struct {
	ID      string `json:"id"`
	Label   string `json:"label"`
	Primary bool   `json:"primary"`
}

type IslandPayload struct {
	Seq         int            `json:"seq"`
	Mode        string         `json:"mode"` // remind | ask | milestone | celebrate | welcome
	Level       int            `json:"level"`
	Title       string         `json:"title"`
	Body        string         `json:"body"`
	Sticky      bool           `json:"sticky"`
	HideAfterMs int            `json:"hideAfterMs"`
	Buttons     []IslandButton `json:"buttons"`
	CanLater    bool           `json:"canLater"`
	LaterMax    int            `json:"laterMax"`
	Fraction    float64        `json:"fraction"`
	Done        int            `json:"done"`
	Target      int            `json:"target"`
	IsTest      bool           `json:"isTest"`
	QuickAdd    [4]int         `json:"quickAdd"`
	Data        map[string]any `json:"data,omitempty"`
}

type islandState struct {
	mu      sync.Mutex
	seq     int
	sticky  bool
	visible bool
	current *IslandPayload
	timer   *time.Timer
}

// islandFor builds the island payload of an action (IMPLEMENTATION.md 7.17).
func islandFor(a core.Action, snap core.Snapshot) IslandPayload {
	p := IslandPayload{Mode: a.Kind, Level: a.Level, Title: a.Title, Body: a.Body, Data: a.Data,
		CanLater: snap.Pnr.CanPromise, LaterMax: snap.Pnr.PromiseMax, Buttons: []IslandButton{}}
	if len(snap.Day.Tasks) > 0 {
		t := snap.Day.Tasks[snap.Day.Active]
		p.Fraction, p.Done, p.Target, p.IsTest = t.Fraction, t.Done, t.Target, t.IsTest
	}
	q := activeQuick(snap)
	p.QuickAdd = q
	add := func(i int) IslandButton {
		return IslandButton{ID: "add:" + strconv.Itoa(i), Label: "+" + strconv.Itoa(q[i])}
	}
	switch a.Kind {
	case "remind":
		if p.IsTest {
			p.Buttons = []IslandButton{{ID: "done", Label: "Test done"}, {ID: "studying", Label: "Taking it now", Primary: true}}
		} else {
			p.Buttons = []IslandButton{add(2), add(3), {ID: "studying", Label: "I'm studying", Primary: true}}
		}
		if a.Level >= 2 {
			p.Sticky = true
		} else {
			p.HideAfterMs = 8000
		}
	case "ask":
		p.Sticky = true
		p.CanLater = false
		if p.IsTest {
			p.Buttons = []IslandButton{{ID: "done", Label: "Test done", Primary: true}, {ID: "more", Label: "Still going"}, {ID: "none", Label: "Not yet"}}
		} else {
			p.Buttons = []IslandButton{add(1), add(2), add(3), {ID: "other", Label: "Other"}, {ID: "none", Label: "Nothing yet"}}
		}
	case "welcome":
		p.HideAfterMs = 8000
		p.CanLater = false
		if p.IsTest {
			p.Buttons = []IslandButton{{ID: "done", Label: "Test done"}, {ID: "studying", Label: "Taking it now", Primary: true}}
		} else {
			p.Buttons = []IslandButton{add(2), add(3), {ID: "studying", Label: "I'm studying", Primary: true}}
		}
	case "milestone":
		p.HideAfterMs = 2200
		p.CanLater = false
	case "celebrate":
		p.HideAfterMs = 5000
		p.CanLater = false
		p.Fraction = 1
	}
	return p
}

// showIsland drops the island with a payload.
func (s *Shell) showIsland(p IslandPayload) {
	if s.island == nil || s.app == nil {
		return
	}
	s.isl.mu.Lock()
	s.isl.seq++
	p.Seq = s.isl.seq
	s.isl.sticky = p.Sticky
	s.isl.visible = true
	cp := p
	s.isl.current = &cp
	if s.isl.timer != nil {
		s.isl.timer.Stop()
		s.isl.timer = nil
	}
	if !p.Sticky && p.HideAfterMs > 0 {
		seq := p.Seq
		s.isl.timer = time.AfterFunc(time.Duration(p.HideAfterMs+2000)*time.Millisecond, func() { s.islandDone(seq, "fallback") })
	}
	s.isl.mu.Unlock()

	s.mu.Lock()
	now := s.env.now()
	if d := s.st.Days[s.st.Today(now)]; d != nil {
		d.LastIslandAt = now
	}
	s.mu.Unlock()

	log.Printf("island show seq=%d mode=%s sticky=%v", p.Seq, p.Mode, p.Sticky)
	s.app.Event.Emit("island", p)
	s.island.Show()
}

// islandDone hides the island if seq is still the one on screen.
func (s *Shell) islandDone(seq int, reason string) {
	s.isl.mu.Lock()
	if seq != s.isl.seq || !s.isl.visible {
		s.isl.mu.Unlock()
		return
	}
	s.isl.visible = false
	s.isl.current = nil
	if s.isl.timer != nil {
		s.isl.timer.Stop()
		s.isl.timer = nil
	}
	s.isl.mu.Unlock()
	log.Printf("island hide seq=%d reason=%s", seq, reason)
	s.island.Hide()
}

// hideIslandIfPassing hides a non-sticky island (the popover was opened).
func (s *Shell) hideIslandIfPassing() {
	s.isl.mu.Lock()
	seq, sticky, vis := s.isl.seq, s.isl.sticky, s.isl.visible
	s.isl.mu.Unlock()
	if vis && !sticky {
		s.islandDone(seq, "popover")
	}
}

// islandCurrent is the payload on screen, for an island page that loaded late.
func (s *Shell) islandCurrent() *IslandPayload {
	s.isl.mu.Lock()
	defer s.isl.mu.Unlock()
	if s.isl.current == nil {
		return nil
	}
	cp := *s.isl.current
	return &cp
}
