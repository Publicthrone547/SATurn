package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/Publicthrone547/SATurn/internal/core"
	"github.com/Publicthrone547/SATurn/internal/trayicon"
)

type trayMenu struct {
	menu     *application.Menu
	status   *application.MenuItem
	adds     [4]*application.MenuItem
	testDone *application.MenuItem
	undo     *application.MenuItem
	session  *application.MenuItem
	rest     *application.MenuItem
	sick     *application.MenuItem
	quit     *application.MenuItem
	key      string
}

// buildTray creates the menu bar item and its menu once.
func (s *Shell) buildTray() {
	s.tray.SetTemplateIcon(trayicon.Render("ring", 0, false))
	s.tray.SetTooltip("SATurn")
	m := &trayMenu{menu: application.NewMenu()}
	m.status = m.menu.Add("SATurn").SetEnabled(false)
	m.menu.AddSeparator()
	for i := range m.adds {
		m.adds[i] = m.menu.Add("+").OnClick(func(*application.Context) {
			snap := s.current()
			if activeIsTest(snap) {
				return
			}
			s.add(-1, activeQuick(snap)[i], "menu")
		})
	}
	m.testDone = m.menu.Add("Test done").OnClick(func(*application.Context) {
		_, _ = s.do("done source=menu", func(st *core.State, now time.Time) ([]core.Action, error) {
			return core.MarkDone(st, now, -1, "menu"), nil
		})
	})
	m.undo = m.menu.Add("Undo last entry").OnClick(func(*application.Context) {
		_, _ = s.do("undo source=menu", wrap(core.Undo))
	})
	m.menu.AddSeparator()
	m.session = m.menu.Add("Start a session").OnClick(func(*application.Context) {
		if s.current().Session != nil {
			_, _ = s.do("end session source=menu", wrap(core.EndSession))
			return
		}
		_, _ = s.do("session source=menu", func(st *core.State, now time.Time) ([]core.Action, error) {
			return core.StartSession(st, now, 0), nil
		})
	})
	m.rest = m.menu.Add("Rest day").OnClick(func(*application.Context) {
		off := core.Rest
		if s.current().Day.Off != "" {
			off = ""
		}
		_, _ = s.do("off="+off+" source=menu", func(st *core.State, now time.Time) ([]core.Action, error) {
			return core.SetOff(st, now, "", off)
		})
	})
	m.sick = m.menu.Add("Sick day").OnClick(func(*application.Context) {
		_, _ = s.do("off=sick source=menu", func(st *core.State, now time.Time) ([]core.Action, error) {
			return core.SetOff(st, now, "", core.Sick)
		})
	})
	m.menu.AddSeparator()
	m.menu.Add("Open SATurn").OnClick(func(*application.Context) { s.openMain("") })
	m.quit = m.menu.Add("Quit SATurn").OnClick(func(*application.Context) { s.app.Quit() })
	s.menu = m
	s.tray.SetMenu(m.menu)
	s.tray.AttachWindow(s.popover).WindowOffset(6)
	s.tray.OnRightClick(func() { s.tray.OpenMenu() })
	log.Printf("tray created")
}

// updateTray changes the icon, the text and the menu only when they changed.
func (s *Shell) updateTray(snap core.Snapshot) {
	if s.tray == nil {
		return
	}
	tv := snap.Tray
	key := trayicon.Key(tv.Glyph, tv.Progress, tv.Session)
	s.viewMu.Lock()
	iconChanged, textChanged := key != s.trayKey, tv.Text != s.trayText
	s.trayKey, s.trayText = key, tv.Text
	s.viewMu.Unlock()
	if iconChanged {
		s.tray.SetTemplateIcon(trayicon.Render(tv.Glyph, tv.Progress, tv.Session))
	}
	if textChanged {
		s.tray.SetLabel(tv.Text)
	}
	if iconChanged || textChanged {
		log.Printf("tray text=%q glyph=%s progress=%.2f alert=%v", tv.Text, tv.Glyph, tv.Progress, tv.Alert)
	}
	s.updateMenu(snap)
}

func (s *Shell) updateMenu(snap core.Snapshot) {
	m := s.menu
	if m == nil {
		return
	}
	study := snap.Phase == "study"
	off := snap.Day.Off != ""
	isTest := activeIsTest(snap)
	q := activeQuick(snap)
	sessionLabel := "Start a " + strconv.Itoa(snap.Settings.SessionMin) + " minute session"
	if isTest {
		sessionLabel = "Start the test (" + strconv.Itoa(snap.Settings.TestSessionMin) + " minutes)"
	}
	if snap.Session != nil {
		sessionLabel = fmt.Sprintf("End session (%d min left)", (snap.Session.RemainingSec+59)/60)
	}
	restLabel := "Rest day"
	if off {
		restLabel = "Back to the plan"
	}
	quitLabel := "Quit SATurn"
	if !snap.CanQuit {
		quitLabel = "Quit SATurn (finish today first)"
	}
	logging := study && !off && len(snap.Day.Tasks) > 0
	key := fmt.Sprint(snap.Tray.Status, study, off, isTest, q, sessionLabel, quitLabel, snap.CanQuit, snap.Day.CanUndo, logging)
	if key == m.key {
		return
	}
	m.key = key
	m.status.SetLabel(snap.Tray.Status)
	for i, it := range m.adds {
		it.SetLabel("+" + strconv.Itoa(q[i])).SetHidden(!logging || isTest)
	}
	m.testDone.SetHidden(!logging || !isTest)
	m.undo.SetHidden(!study).SetEnabled(snap.Day.CanUndo)
	m.session.SetLabel(sessionLabel).SetHidden(!logging || snap.Day.Done)
	m.rest.SetLabel(restLabel).SetHidden(!study)
	m.sick.SetHidden(!study || off)
	m.quit.SetLabel(quitLabel).SetEnabled(snap.CanQuit)
	m.menu.Update()
}
