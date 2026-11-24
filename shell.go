package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"github.com/Publicthrone547/SATurn/internal/core"
	"github.com/Publicthrone547/SATurn/internal/sound"
	"github.com/Publicthrone547/SATurn/internal/store"
)

// BailoutPhrase is what strict mode asks the owner to type to quit anyway.
const BailoutPhrase = "I am done for today"

// Shell is the only place where the core, the clock, storage and Wails meet.
// Rule: never call a Wails API while holding mu.
type Shell struct {
	mu    sync.Mutex
	st    *core.State
	store *store.Store
	env   env
	sys   core.SystemView // guarded by mu

	app     *application.App
	tray    *application.SystemTray
	menu    *trayMenu
	popover *application.WebviewWindow
	main    *application.WebviewWindow
	island  *application.NotchWindow
	notes   *notifications.NotificationService

	viewMu   sync.Mutex
	lastSnap core.Snapshot
	trayKey  string
	trayText string
	catKey   string

	isl islandState

	notifMu   sync.Mutex
	lastNotif []string

	keysPaused bool // guarded by mu; the settings page is recording a hotkey

	allowQuit atomic.Bool
	woke      atomic.Bool
	started   atomic.Bool
	prevIdle  int
	kick      chan struct{}
}

type cmdFn func(st *core.State, now time.Time) ([]core.Action, error)

func newShell(notes *notifications.NotificationService, e env) (*Shell, error) {
	dir, err := store.DefaultDir()
	if err != nil {
		return nil, err
	}
	stor, err := store.Open(dir)
	if err != nil {
		return nil, err
	}
	st, err := stor.Load()
	if err != nil {
		return nil, err
	}
	s := &Shell{st: st, store: stor, env: e, notes: notes, kick: make(chan struct{}, 1)}
	s.sys = core.SystemView{Version: version, Installed: e.installed, Dry: e.dry, DataDir: homeRel(dir), HotkeyErrors: []string{}}
	now := e.now()
	core.SuggestPlan(s.st, now, false) // creates missing plan days; sets FirstDay when empty
	if err := s.store.Save(s.st); err != nil {
		log.Printf("store save err=%q", err)
	}
	return s, nil
}

// do runs a command under the lock, saves, then performs its actions and
// pushes a snapshot.
func (s *Shell) do(name string, fn cmdFn) (core.Snapshot, error) {
	s.mu.Lock()
	now := s.env.now()
	acts, err := fn(s.st, now)
	snap := s.snapshotLocked(now)
	saveErr := s.store.Save(s.st)
	s.mu.Unlock()
	if saveErr != nil {
		log.Printf("store save err=%q", saveErr)
	}
	if err != nil {
		log.Printf("cmd %s err=%q", name, err)
	} else if name != "" {
		log.Printf("cmd %s %s", name, dayLine(snap))
	}
	s.perform(acts, snap)
	s.push(snap)
	return snap, err
}

func dayLine(snap core.Snapshot) string {
	var parts []string
	for _, t := range snap.Day.Tasks {
		parts = append(parts, t.Kind+"="+strconv.Itoa(t.Done)+"/"+strconv.Itoa(t.Target))
	}
	if snap.Day.Off != "" {
		parts = append(parts, "off="+snap.Day.Off)
	}
	return "date=" + snap.Today + " " + strings.Join(parts, " ")
}

func (s *Shell) snapshotLocked(now time.Time) core.Snapshot {
	snap := core.View(s.st, now)
	snap.System = s.sys
	snap.System.HotkeyErrors = append([]string{}, s.sys.HotkeyErrors...)
	return snap
}

// Snapshot returns the current snapshot without changing anything.
func (s *Shell) Snapshot() core.Snapshot {
	s.mu.Lock()
	snap := s.snapshotLocked(s.env.now())
	s.mu.Unlock()
	return snap
}

func (s *Shell) setSys(fn func(*core.SystemView)) {
	s.mu.Lock()
	fn(&s.sys)
	s.mu.Unlock()
}

// push sends the snapshot to every window and updates the menu bar.
func (s *Shell) push(snap core.Snapshot) {
	s.viewMu.Lock()
	s.lastSnap = snap
	s.viewMu.Unlock()
	if s.app == nil || !s.started.Load() {
		return
	}
	s.app.Event.Emit("state", snap)
	s.updateTray(snap)
	s.syncCategories(snap)
}

func (s *Shell) current() core.Snapshot {
	s.viewMu.Lock()
	defer s.viewMu.Unlock()
	return s.lastSnap
}

// start runs once the application is up.
func (s *Shell) start() {
	log.Printf("start version=%s installed=%v dry=%v data=%s", version, s.env.installed, s.env.dry, s.store.Dir)
	if err := sound.Init(); err != nil {
		log.Printf("sound init err=%q", err)
	}
	s.setSys(func(v *core.SystemView) { v.ReduceMotion = reduceMotion() })
	s.started.Store(true)
	s.setupNotifications()
	s.registerHotkeys()
	s.syncAutostart()
	snap := s.Snapshot()
	s.push(snap)
	if snap.Phase == "setup" {
		s.openMain("")
	}
	go s.loop()
	if s.env.open != "" {
		time.AfterFunc(2*time.Second, func() { s.openForTest(s.env.open) })
	}
	if s.env.script != "" {
		go s.runScript(s.env.script)
	}
}

// loop is the ticker: every tick it reads presence and runs the engine.
func (s *Shell) loop() {
	t := time.NewTicker(s.env.tick)
	defer t.Stop()
	s.tick()
	for {
		select {
		case <-t.C:
		case <-s.kick:
		}
		s.tick()
	}
}

func (s *Shell) tick() {
	idle := s.env.idle
	if idle < 0 {
		idle = idleSeconds()
	}
	returned := s.woke.Swap(false) || (s.prevIdle >= 600 && idle < 60)
	s.prevIdle = idle
	s.mu.Lock()
	now := s.env.now()
	acts := core.Tick(s.st, core.TickInput{Now: now, IdleSeconds: idle, Returned: returned})
	snap := s.snapshotLocked(now)
	err := s.store.Save(s.st)
	s.mu.Unlock()
	if err != nil {
		log.Printf("store save err=%q", err)
	}
	s.perform(acts, snap)
	s.push(snap)
}

func (s *Shell) woken() {
	log.Printf("wake")
	s.woke.Store(true)
	select {
	case s.kick <- struct{}{}:
	default:
	}
}

// perform carries out the actions a command or a tick returned.
func (s *Shell) perform(acts []core.Action, snap core.Snapshot) {
	for _, a := range acts {
		switch a.Kind {
		case "remind":
			log.Printf("remind level=%d situation=%s title=%q body=%q", a.Level, a.Situation, a.Title, a.Body)
			s.sendReminder(a, snap)
			s.showIsland(islandFor(a, snap))
			if a.Level >= 2 {
				if a.Situation == "red" {
					s.play(sound.Alarm)
				} else {
					s.play(sound.Remind)
				}
			}
		case "notify":
			log.Printf("notify situation=%s title=%q body=%q", a.Situation, a.Title, a.Body)
			s.sendPlain(a)
		case "ask":
			log.Printf("ask title=%q", a.Title)
			s.sendReminder(a, snap)
			s.showIsland(islandFor(a, snap))
			s.play(sound.Tick)
		case "milestone":
			log.Printf("milestone quarter=%v title=%q", a.Data["quarter"], a.Title)
			s.emitCelebrate("quarter", a)
			s.showIsland(islandFor(a, snap))
			s.play(sound.Quarter)
		case "celebrate":
			log.Printf("celebrate kind=%s title=%q body=%q", a.Situation, a.Title, a.Body)
			s.emitCelebrate(a.Situation, a)
			s.showIsland(islandFor(a, snap))
			s.play(sound.Done)
			s.clearReminders()
		case "welcome":
			log.Printf("welcome title=%q", a.Title)
			s.showIsland(islandFor(a, snap))
		case "exam", "retire":
			log.Printf("%s", a.Kind)
			s.openMain("")
		}
	}
}

func (s *Shell) emitCelebrate(kind string, a core.Action) {
	if s.app == nil {
		return
	}
	data := map[string]any{}
	for k, v := range a.Data {
		data[k] = v
	}
	data["task"] = a.Data["kind"]
	data["kind"], data["title"], data["body"] = kind, a.Title, a.Body
	s.app.Event.Emit("celebrate", data)
}

func (s *Shell) play(name string) {
	s.mu.Lock()
	on := s.st.Settings.Sound
	s.mu.Unlock()
	if !on {
		return
	}
	if s.env.dry {
		log.Printf("sound dry name=%s", name)
		return
	}
	sound.Play(name)
}

// add logs progress from any surface and plays the tick.
func (s *Shell) add(task, delta int, source string) core.Snapshot {
	snap, _ := s.addActs(task, delta, source)
	return snap
}

// addActs is add that also reports how many actions (milestones,
// celebrations) the entry caused.
func (s *Shell) addActs(task, delta int, source string) (core.Snapshot, int) {
	var acts int
	snap, _ := s.do("add delta="+strconv.Itoa(delta)+" source="+source, func(st *core.State, now time.Time) ([]core.Action, error) {
		a := core.Add(st, now, task, delta, source)
		acts = len(a)
		return a, nil
	})
	if delta > 0 && acts == 0 {
		s.play(sound.Tick)
	}
	return snap, acts
}

// quickAdd logs from a hotkey or a link, when no SATurn window may be in
// view, and confirms the new count on the island for a moment.
func (s *Shell) quickAdd(n int, source string) {
	task := questionTask(s.current())
	if task == -2 {
		log.Printf("add ignored source=%s: today has only a practice test", source)
		return
	}
	snap, acts := s.addActs(task, n, source)
	if acts == 0 {
		s.showTally(n, task, snap)
	}
}

// questionTask is the task a hotkey or a link adds to: the active one, or on
// a day with a test and questions, the question task while the test is
// active. -2 when today has nothing to count.
func questionTask(snap core.Snapshot) int {
	if !activeIsTest(snap) {
		return -1
	}
	for i, t := range snap.Day.Tasks {
		if !t.IsTest {
			return i
		}
	}
	return -2
}

// activeIsTest reports whether today's active task is a practice test.
func activeIsTest(snap core.Snapshot) bool {
	if len(snap.Day.Tasks) == 0 {
		return false
	}
	return snap.Day.Tasks[snap.Day.Active].IsTest
}

func activeQuick(snap core.Snapshot) [4]int {
	if len(snap.Day.Tasks) == 0 {
		return [4]int{1, 5, 10, 25}
	}
	q := snap.Day.Tasks[snap.Day.Active].QuickAdd
	if q == [4]int{} {
		return [4]int{1, 5, 10, 25}
	}
	return q
}

// openMain shows the main window, optionally on a tab.
func (s *Shell) openMain(tab string) {
	if s.main == nil {
		return
	}
	if tab != "" {
		s.app.Event.Emit("navigate", tab)
	}
	s.main.Show().Focus()
	activateApp()
}

func (s *Shell) showPopover(focus string) {
	if s.tray == nil {
		return
	}
	if focus != "" {
		s.app.Event.Emit("popover:focus", focus)
	}
	s.tray.ShowWindow()
	activateApp()
}

func (s *Shell) togglePopover() {
	if s.popover != nil && s.popover.IsVisible() {
		s.popover.Hide()
		return
	}
	s.showPopover("")
}

// popoverShown refreshes the popover and retires a passing island.
func (s *Shell) popoverShown() {
	s.push(s.Snapshot())
	s.hideIslandIfPassing()
}

// shouldQuit is consulted for every quit request. Strict mode blocks only a
// quit from inside the app; logout, restart and scripts always go through.
func (s *Shell) shouldQuit() bool {
	if s.allowQuit.Load() || quitFromOutside() {
		return true
	}
	if snap := s.Snapshot(); !snap.CanQuit {
		log.Printf("quit blocked strict=true allow=%v", s.allowQuit.Load())
		go s.openMain("settings")
		return false
	}
	return true
}

func (s *Shell) quit() {
	s.allowQuit.Store(true)
	s.app.Quit()
}

func (s *Shell) shutdown() {
	s.mu.Lock()
	if err := s.store.Save(s.st); err != nil {
		log.Printf("store save err=%q", err)
	}
	s.mu.Unlock()
	log.Printf("shutdown")
}

// bailout quits despite strict mode once the phrase is typed exactly.
func (s *Shell) bailout(phrase string) error {
	if strings.TrimSpace(phrase) != BailoutPhrase {
		return errors.New(`type "I am done for today" exactly`)
	}
	_, _ = s.do("bailout", func(st *core.State, now time.Time) ([]core.Action, error) {
		core.Bailout(st, now)
		return nil, nil
	})
	s.quit()
	return nil
}

func (s *Shell) retireAndQuit() {
	s.setLaunchAtLogin(false)
	_, _ = s.do("retire", func(st *core.State, now time.Time) ([]core.Action, error) {
		core.Retire(st)
		return nil, nil
	})
	s.quit()
}

// testReminder shows a reminder of the given level without changing state.
func (s *Shell) testReminder(level int) {
	s.mu.Lock()
	now := s.env.now()
	a := core.PreviewRemind(s.st, now, level)
	snap := s.snapshotLocked(now)
	s.mu.Unlock()
	log.Printf("test reminder level=%d", level)
	s.perform([]core.Action{a}, snap)
}

func (s *Shell) openForTest(what string) {
	log.Printf("open %s", what)
	switch what {
	case "popover":
		s.showPopover("")
	case "main":
		s.openMain("")
	case "island":
		s.testReminder(2)
	default:
		if strings.HasPrefix(what, "main:") {
			s.openMain(strings.TrimPrefix(what, "main:"))
		}
	}
}

// runScript drives commands one per second (SATURN_SCRIPT), through the
// same path a click uses.
func (s *Shell) runScript(script string) {
	time.Sleep(2 * time.Second)
	for _, raw := range strings.Split(script, ";") {
		f := strings.Fields(raw)
		if len(f) == 0 {
			continue
		}
		n := 0
		if len(f) > 1 {
			n, _ = strconv.Atoi(f[1])
		}
		log.Printf("script %s", strings.Join(f, " "))
		var err error
		switch f[0] {
		case "add":
			s.add(-1, n, "script")
		case "undo":
			_, err = s.do("undo", wrap(core.Undo))
		case "exact":
			_, err = s.do("exact", func(st *core.State, now time.Time) ([]core.Action, error) { return core.SetExact(st, now, -1, n), nil })
		case "done":
			_, err = s.do("done", func(st *core.State, now time.Time) ([]core.Action, error) {
				return core.MarkDone(st, now, -1, "script"), nil
			})
		case "session":
			_, err = s.do("session", func(st *core.State, now time.Time) ([]core.Action, error) { return core.StartSession(st, now, n), nil })
		case "end":
			_, err = s.do("end session", wrap(core.EndSession))
		case "promise":
			_, err = s.do("promise", func(st *core.State, now time.Time) ([]core.Action, error) { return core.MakePromise(st, now, n) })
		case "off":
			off := ""
			if len(f) > 1 && f[1] != "none" {
				off = f[1]
			}
			_, err = s.do("off", func(st *core.State, now time.Time) ([]core.Action, error) { return core.SetOff(st, now, "", off) })
		case "test":
			s.testReminder(n)
		case "quit":
			// The same path as Cmd+Q or the menu item.
			s.app.Quit()
		case "bailout":
			err = s.bailout(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(raw), "bailout")))
		default:
			err = errors.New("unknown verb")
		}
		if err != nil {
			log.Printf("script %s err=%q", f[0], err)
		}
		time.Sleep(time.Second)
	}
	log.Printf("script done")
}

func wrap(fn func(*core.State, time.Time) []core.Action) cmdFn {
	return func(st *core.State, now time.Time) ([]core.Action, error) { return fn(st, now), nil }
}
