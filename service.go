package main

import (
	"maps"
	"time"

	"github.com/Publicthrone547/SATurn/internal/core"
)

// AppService is the API the frontend calls. Every method that changes state
// returns the new snapshot.
type AppService struct{ sh *Shell }

func (a *AppService) Snapshot() core.Snapshot { return a.sh.Snapshot() }

func (a *AppService) Add(task int, delta int, source string) core.Snapshot {
	return a.sh.add(task, delta, source)
}

func (a *AppService) SetExact(task int, value int) core.Snapshot {
	snap, _ := a.sh.do("exact", func(st *core.State, now time.Time) ([]core.Action, error) {
		return core.SetExact(st, now, task, value), nil
	})
	return snap
}

func (a *AppService) MarkDone(task int) core.Snapshot {
	snap, _ := a.sh.do("done", func(st *core.State, now time.Time) ([]core.Action, error) {
		return core.MarkDone(st, now, task, "popover"), nil
	})
	return snap
}

func (a *AppService) Undo() core.Snapshot {
	snap, _ := a.sh.do("undo", wrap(core.Undo))
	return snap
}

func (a *AppService) StartSession(minutes int) core.Snapshot {
	snap, _ := a.sh.do("session", func(st *core.State, now time.Time) ([]core.Action, error) {
		return core.StartSession(st, now, minutes), nil
	})
	return snap
}

func (a *AppService) EndSession() core.Snapshot {
	snap, _ := a.sh.do("end session", wrap(core.EndSession))
	return snap
}

func (a *AppService) MakePromise(minutes int) (core.Snapshot, error) {
	return a.sh.do("promise", func(st *core.State, now time.Time) ([]core.Action, error) {
		return core.MakePromise(st, now, minutes)
	})
}

func (a *AppService) SetOff(date string, off string) (core.Snapshot, error) {
	return a.sh.do("off="+off, func(st *core.State, now time.Time) ([]core.Action, error) {
		return core.SetOff(st, now, date, off)
	})
}

func (a *AppService) SetDay(date string, preset string, targets map[string]int) (core.Snapshot, error) {
	return a.sh.do("day "+date+" "+preset, func(st *core.State, now time.Time) ([]core.Action, error) {
		return core.SetDay(st, now, date, preset, targets)
	})
}

func (a *AppService) SuggestPlan(overwrite bool) core.Snapshot {
	snap, _ := a.sh.do("suggest", func(st *core.State, now time.Time) ([]core.Action, error) {
		return core.SuggestPlan(st, now, overwrite), nil
	})
	return snap
}

func (a *AppService) UpdateSettings(s core.Settings) core.Snapshot {
	before := a.sh.Snapshot().Settings
	snap, _ := a.sh.do("settings", func(st *core.State, now time.Time) ([]core.Action, error) {
		return core.UpdateSettings(st, now, s), nil
	})
	if snap.Settings.Hotkeys != before.Hotkeys || !maps.Equal(snap.Settings.Keys, before.Keys) {
		a.sh.registerHotkeys()
		snap = a.sh.Snapshot()
		a.sh.push(snap)
	}
	return snap
}

// PauseHotkeys releases the global hotkeys while a new one is being recorded.
func (a *AppService) PauseHotkeys(paused bool) { a.sh.pauseHotkeys(paused) }

func (a *AppService) CompleteOnboarding(targetScore int) core.Snapshot {
	snap, _ := a.sh.do("onboarding", func(st *core.State, now time.Time) ([]core.Action, error) {
		return core.CompleteOnboarding(st, now, targetScore), nil
	})
	return snap
}

func (a *AppService) SetScore(date string, score int) (core.Snapshot, error) {
	return a.sh.do("score", func(st *core.State, now time.Time) ([]core.Action, error) {
		return nil, core.SetScore(st, now, date, score)
	})
}

// IslandDone ends an island presentation: "button", "timeout" or "dismiss".
func (a *AppService) IslandDone(seq int, reason string) { a.sh.islandDone(seq, reason) }

// IslandCurrent is the payload on screen, or nil.
func (a *AppService) IslandCurrent() *IslandPayload { return a.sh.islandCurrent() }

// OpenMain shows the main window: "today", "plan", "history", "settings" or "".
func (a *AppService) OpenMain(tab string) { a.sh.openMain(tab) }

// ShowPopover opens the popover; focus "exact" focuses the exact-value field.
func (a *AppService) ShowPopover(focus string) { a.sh.showPopover(focus) }

func (a *AppService) HidePopover() {
	if a.sh.popover != nil {
		a.sh.popover.Hide()
	}
}

func (a *AppService) RequestNotifications() bool { return a.sh.requestNotifications() }

func (a *AppService) OpenNotificationSettings() {
	_ = a.sh.app.Browser.OpenURL("x-apple.systempreferences:com.apple.Notifications-Settings.extension")
}

func (a *AppService) SetLaunchAtLogin(on bool) core.Snapshot {
	a.sh.setLaunchAtLogin(on)
	snap := a.sh.Snapshot()
	a.sh.push(snap)
	return snap
}

func (a *AppService) TestReminder(level int) { a.sh.testReminder(level) }

// Bailout quits despite strict mode; phrase must equal "I am done for today".
func (a *AppService) Bailout(phrase string) error { return a.sh.bailout(phrase) }

func (a *AppService) RetireAndQuit() { a.sh.retireAndQuit() }
