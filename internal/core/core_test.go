package core

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

const day = "2026-11-24" // a Tuesday, a math day in the proposal

func at(t *testing.T, hhmm string) time.Time {
	t.Helper()
	return atOn(t, day, hhmm)
}

func atOn(t *testing.T, date, hhmm string) time.Time {
	t.Helper()
	var h, m int
	if _, err := fmt.Sscanf(hhmm, "%d:%d", &h, &m); err != nil {
		t.Fatalf("bad time %q", hhmm)
	}
	return AtMinute(date, h*60+m)
}

// fresh is an onboarded state whose plan starts on day with a math day.
func fresh(t *testing.T, preset string) *State {
	t.Helper()
	st := NewState()
	st.Settings.Onboarded = true
	st.FirstDay = day
	if _, err := SetDay(st, at(t, "08:00"), day, preset, nil); err != nil {
		t.Fatal(err)
	}
	return st
}

type hit struct {
	At string
	Action
}

func (h hit) String() string {
	if h.Kind == "remind" {
		return fmt.Sprintf("%s remind %s L%d", h.At, h.Situation, h.Level)
	}
	return fmt.Sprintf("%s %s %s", h.At, h.Kind, h.Situation)
}

// step runs the engine one minute at a time from..to (inclusive) on date,
// calling before(m) first each minute. skip reports minutes with no tick.
type sim struct {
	idle   func(m int) int
	skip   func(m int) bool
	before func(now time.Time, m int) []Action
}

func runDay(t *testing.T, st *State, date, from, to string, s sim) []hit {
	t.Helper()
	var out []hit
	f, e := MinuteOfDay(atOn(t, date, from)), MinuteOfDay(atOn(t, date, to))
	for m := f; m <= e; m++ {
		now := AtMinute(date, m)
		if s.before != nil {
			for _, a := range s.before(now, m) {
				out = append(out, hit{ClockOf(now), a})
			}
		}
		if s.skip != nil && s.skip(m) {
			continue
		}
		idle := 0
		if s.idle != nil {
			idle = s.idle(m)
		}
		for _, a := range Tick(st, TickInput{Now: now, IdleSeconds: idle}) {
			out = append(out, hit{ClockOf(now), a})
		}
	}
	return out
}

func strs(hs []hit) []string {
	var out []string
	for _, h := range hs {
		out = append(out, h.String())
	}
	return out
}

func expect(t *testing.T, got []hit, want ...string) {
	t.Helper()
	g := strs(got)
	if strings.Join(g, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got:\n  %s\nwant:\n  %s", strings.Join(g, "\n  "), strings.Join(want, "\n  "))
	}
}

func hm(h, m int) int { return h*60 + m }

// 1
func TestLogicalDate(t *testing.T) {
	if got := LogicalDate(atOn(t, "2026-11-25", "03:59"), 4); got != "2026-11-24" {
		t.Fatalf("03:59 → %s", got)
	}
	if got := LogicalDate(atOn(t, "2026-11-25", "04:00"), 4); got != "2026-11-25" {
		t.Fatalf("04:00 → %s", got)
	}
}

// 2
func TestProposedPlan(t *testing.T) {
	want := "english test math english math english test rest math english test math rest"
	var got []string
	for d := "2026-11-22"; d < "2026-12-05"; d = AddDays(d, 1) {
		got = append(got, ProposedPreset(d, "2026-12-05"))
	}
	if strings.Join(got, " ") != want {
		t.Fatalf("got %v", got)
	}
	if p := ProposedPreset("2026-12-05", "2026-12-05"); p != "" {
		t.Fatalf("exam day proposal %q", p)
	}
	if p := ProposedPreset("2026-11-15", "2026-12-05"); p != ProposedPreset("2026-11-22", "2026-12-05") {
		t.Fatalf("k > 13 does not repeat k-7: %q", p)
	}
}

// 3, 4
func TestMilestones(t *testing.T) {
	st := fresh(t, Math)
	acts := Add(st, at(t, "10:00"), -1, 30, "test")
	if len(acts) != 1 || acts[0].Kind != "milestone" || acts[0].Data["quarter"] != 1 {
		t.Fatalf("add 30: %+v", acts)
	}
	st = fresh(t, Math)
	Add(st, at(t, "10:00"), -1, 20, "test")
	acts = Add(st, at(t, "10:05"), -1, 60, "test")
	if len(acts) != 1 || acts[0].Data["quarter"] != 3 {
		t.Fatalf("add 60 from 20: %+v", acts)
	}
}

// 5, 6
func TestCelebrateOnce(t *testing.T) {
	st := fresh(t, Math)
	acts := Add(st, at(t, "10:00"), -1, 100, "test")
	if len(acts) != 1 || acts[0].Kind != "celebrate" || acts[0].Situation != "day" {
		t.Fatalf("to target: %+v", acts)
	}
	if acts := Add(st, at(t, "10:01"), -1, 5, "test"); len(acts) != 0 {
		t.Fatalf("second add celebrated again: %+v", acts)
	}
	Undo(st, at(t, "10:02")) // back to 100: still done
	Undo(st, at(t, "10:03")) // 0: not done
	d := st.Days[day]
	if !d.CelebratedAt.IsZero() || d.IsDone() {
		t.Fatalf("undo below target: done=%d celebrated=%v", d.Tasks[0].Done, d.CelebratedAt)
	}
	acts = Add(st, at(t, "10:04"), -1, 100, "test")
	if len(acts) != 1 || acts[0].Kind != "celebrate" {
		t.Fatalf("celebrate again: %+v", acts)
	}
}

// 7
func TestDoneNeverNegative(t *testing.T) {
	st := fresh(t, Math)
	Add(st, at(t, "10:00"), -1, 10, "test")
	Add(st, at(t, "10:01"), -1, -50, "test")
	if got := st.Days[day].Tasks[0].Done; got != 0 {
		t.Fatalf("done %d", got)
	}
}

// 8, 9, 10
func TestBrief(t *testing.T) {
	st := fresh(t, Math)
	if a := Tick(st, TickInput{Now: at(t, "08:59")}); len(a) != 0 {
		t.Fatalf("08:59: %+v", a)
	}
	a := Tick(st, TickInput{Now: at(t, "09:00")})
	if len(a) != 1 || a[0].Situation != "brief" || a[0].Level != 1 {
		t.Fatalf("09:00: %+v", a)
	}
	if a := Tick(st, TickInput{Now: at(t, "09:00").Add(20 * time.Second)}); len(a) != 0 {
		t.Fatalf("09:00:20: %+v", a)
	}
}

// 11: timeline A
func TestTimelineA(t *testing.T) {
	st := fresh(t, Math)
	got := runDay(t, st, day, "08:00", "23:30", sim{})
	want := []string{"09:00 remind brief L1", "10:00 remind zero L1", "11:00 remind zero L2", "12:00 remind zero L2"}
	for h := 13; h <= 19; h++ {
		want = append(want, fmt.Sprintf("%02d:00 remind zero L3", h))
	}
	want = append(want, "19:50 remind lastStart L3",
		"20:30 remind red L3", "21:00 remind red L3", "21:30 remind red L3", "22:00 remind red L3", "22:30 remind red L3",
		"23:00 notify closing")
	expect(t, got, want...)
	fits := []string{}
	for _, h := range got {
		if h.Situation == "red" {
			fits = append(fits, fmt.Sprint(h.Data["fit"]))
		}
	}
	if strings.Join(fits, " ") != "93 75 56 37 18" {
		t.Fatalf("fit values %v", fits)
	}
	if got[12].Title != "20:30. 93 of 100 still fit." {
		t.Fatalf("red title %q", got[12].Title)
	}
}

// 12: timeline B
func TestTimelineB(t *testing.T) {
	st := fresh(t, Math)
	events := map[int]func(now time.Time) []Action{
		hm(10, 10): func(now time.Time) []Action { return Add(st, now, -1, 25, "test") },
		hm(11, 50): func(now time.Time) []Action { return Add(st, now, -1, 25, "test") },
		hm(14, 0):  func(now time.Time) []Action { return StartSession(st, now, 50) },
		hm(14, 52): func(now time.Time) []Action { return Add(st, now, -1, 25, "test") },
		hm(16, 30): func(now time.Time) []Action { return Add(st, now, -1, 25, "test") },
	}
	got := runDay(t, st, day, "08:00", "23:30", sim{before: func(now time.Time, m int) []Action {
		if f := events[m]; f != nil {
			return f(now)
		}
		return nil
	}})
	expect(t, got,
		"09:00 remind brief L1", "10:00 remind zero L1", "10:10 milestone q1",
		"11:00 remind progress L1", "11:50 milestone q2", "13:00 remind progress L1",
		"14:50 ask ask", "14:52 milestone q3", "16:00 remind lastQuarter L1", "16:30 celebrate day")
	if got[3].Title != "25 of 100. 75 left, about 2 h." || got[5].Title != "50 of 100. 50 left, about 1 h 20 min." || got[8].Title != "75 of 100. 25 left." {
		t.Fatalf("titles: %q / %q / %q", got[3].Title, got[5].Title, got[8].Title)
	}
}

// 13
func TestIdleHoldsReminder(t *testing.T) {
	st := fresh(t, Math)
	got := runDay(t, st, day, "08:58", "10:30", sim{idle: func(m int) int {
		if m >= hm(9, 55) && m < hm(10, 7) {
			return 600
		}
		return 2
	}})
	expect(t, got, "09:00 remind brief L1", "10:07 remind zero L1")
}

// 14
func TestCatchUpAfterSleep(t *testing.T) {
	st := fresh(t, Math)
	got := runDay(t, st, day, "08:58", "14:30", sim{skip: func(m int) bool { return m > hm(10, 30) && m < hm(14, 20) }})
	expect(t, got, "09:00 remind brief L1", "10:00 remind zero L1", "14:20 remind zero L1")
	if st.Days[day].Ignored != 1 {
		t.Fatalf("ignored %d", st.Days[day].Ignored)
	}
}

// 15
func TestSession(t *testing.T) {
	st := fresh(t, Math)
	got := runDay(t, st, day, "13:59", "15:00", sim{before: func(now time.Time, m int) []Action {
		if m == hm(13, 59) {
			// The morning already happened.
			d := st.TodayDay(now)
			d.BriefSent, d.LastSlot, d.LastRemindAt = true, at(t, "13:00"), at(t, "13:00")
		}
		if m == hm(14, 0) {
			return StartSession(st, now, 50)
		}
		return nil
	}})
	expect(t, got, "14:50 ask ask", "15:00 remind zero L1")
}

// 16
func TestPromiseBroken(t *testing.T) {
	st := fresh(t, Math)
	got := runDay(t, st, day, "08:59", "12:00", sim{before: func(now time.Time, m int) []Action {
		if m == hm(10, 1) {
			if _, err := MakePromise(st, now, 30); err != nil {
				t.Fatal(err)
			}
		}
		return nil
	}})
	expect(t, got, "09:00 remind brief L1", "10:00 remind zero L1", "10:31 remind promise L2", "10:41 remind broken L3", "12:00 remind zero L3")
	if st.Days[day].Ignored < 4 {
		t.Fatalf("ignored %d", st.Days[day].Ignored)
	}
	if got[3].Title != "You said 10:31. It is 10:41." {
		t.Fatalf("broken title %q", got[3].Title)
	}
}

// 17
func TestPromiseKept(t *testing.T) {
	st := fresh(t, Math)
	got := runDay(t, st, day, "08:59", "10:59", sim{before: func(now time.Time, m int) []Action {
		if m == hm(10, 1) {
			_, _ = MakePromise(st, now, 30)
		}
		if m == hm(10, 35) {
			return Add(st, now, -1, 5, "test")
		}
		return nil
	}})
	expect(t, got, "09:00 remind brief L1", "10:00 remind zero L1", "10:31 remind promise L2")
	if st.Days[day].Promise != nil {
		t.Fatal("promise still set")
	}
}

// 18
func TestPromiseLimits(t *testing.T) {
	st := fresh(t, Math)
	for i := 0; i < 3; i++ {
		if _, err := MakePromise(st, at(t, "10:00"), 15); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := MakePromise(st, at(t, "10:00"), 15); err != ErrNoPromisesLeft {
		t.Fatalf("fourth promise: %v", err)
	}
	st = fresh(t, Math)
	if _, err := MakePromise(st, at(t, "20:00"), 30); err != ErrPastDeadline {
		t.Fatalf("past the PNR: %v", err)
	}
}

// 19
func TestRestSilences(t *testing.T) {
	st := fresh(t, Math)
	got := runDay(t, st, day, "08:59", "23:30", sim{before: func(now time.Time, m int) []Action {
		if m == hm(10, 30) {
			if _, err := SetOff(st, now, day, Rest); err != nil {
				t.Fatal(err)
			}
		}
		return nil
	}})
	expect(t, got, "09:00 remind brief L1", "10:00 remind zero L1")
	if len(st.Days[day].Changes) != 1 {
		t.Fatalf("changes %+v", st.Days[day].Changes)
	}
}

// 20
func TestTestDay(t *testing.T) {
	st := fresh(t, Test)
	if p := Clock(st.Pnr(st.Days[day])); p != "20:30" {
		t.Fatalf("pnr %s", p)
	}
	got := runDay(t, st, day, "19:30", "21:35", sim{before: func(now time.Time, m int) []Action {
		if m == hm(19, 30) {
			d := st.TodayDay(now)
			d.BriefSent, d.LastSlot, d.LastRemindAt = true, at(t, "19:00"), at(t, "19:00")
		}
		return nil
	}})
	expect(t, got, "20:00 remind lastStart L2", "20:30 remind red L3", "21:00 remind red L3", "21:30 remind red L3")
}

// 21
func TestAfterMidnight(t *testing.T) {
	st := fresh(t, Math)
	runDay(t, st, day, "08:59", "23:30", sim{})
	late := atOn(t, "2026-11-25", "00:30")
	if a := Tick(st, TickInput{Now: late}); len(a) != 0 {
		t.Fatalf("00:30: %+v", a)
	}
	Add(st, late, -1, 10, "test")
	if st.Days[day].Tasks[0].Done != 10 {
		t.Fatal("00:30 did not count for the 24th")
	}
}

// 22
func TestMissedDay(t *testing.T) {
	st := fresh(t, Math)
	Add(st, at(t, "10:00"), -1, 37, "test")
	a := Tick(st, TickInput{Now: atOn(t, "2026-11-25", "09:00")})
	if st.Days[day].Closed != "missed" {
		t.Fatalf("closed %q", st.Days[day].Closed)
	}
	if len(a) != 1 || a[0].Situation != "brief" || a[0].Body != "Yesterday closed at 37 of 100. New day." {
		t.Fatalf("brief %+v", a)
	}
}

// 23
func TestExamEve(t *testing.T) {
	st := fresh(t, Math)
	eve := "2026-12-04" // a rest day in the proposal
	got := runDay(t, st, eve, "20:58", "21:30", sim{})
	expect(t, got, "21:00 notify eve")
}

// 24
func TestExamDay(t *testing.T) {
	st := fresh(t, Math)
	got := runDay(t, st, "2026-12-05", "05:00", "23:30", sim{})
	expect(t, got, "06:00 exam exam")
}

// 25
func TestRetire(t *testing.T) {
	st := fresh(t, Math)
	got := runDay(t, st, "2026-12-06", "09:00", "10:00", sim{})
	expect(t, got, "09:00 retire retire")
}

// 26
func TestStreak(t *testing.T) {
	st := NewState()
	st.Settings.Onboarded = true
	st.FirstDay = "2026-11-18"
	now := atOn(t, "2026-11-24", "08:00")
	st.ensurePlan("2026-11-24")
	set := func(date, kind string) {
		d := st.Days[date]
		switch kind {
		case "done":
			d.Off, d.Tasks = "", []Task{{Kind: Math, Target: 100, Done: 100}}
		case "missed":
			d.Off, d.Tasks = "", []Task{{Kind: Math, Target: 100, Done: 10}}
		default:
			d.Off = kind
		}
	}
	set("2026-11-18", "missed")
	set("2026-11-19", "done")
	set("2026-11-20", Rest)
	set("2026-11-21", "done")
	set("2026-11-22", Sick)
	set("2026-11-23", "done")
	if n := st.Streak(st.Today(now)); n != 3 {
		t.Fatalf("streak %d", n)
	}
}

// 27
func TestTrayText(t *testing.T) {
	cases := []struct {
		name string
		prep func(st *State) time.Time
		want string
	}{
		{"setup", func(st *State) time.Time { st.Settings.Onboarded = false; return at(t, "10:00") }, "setup"},
		{"exam", func(st *State) time.Time { return atOn(t, "2026-12-05", "10:00") }, "SAT"},
		{"after", func(st *State) time.Time { return atOn(t, "2026-12-06", "10:00") }, ""},
		{"rest", func(st *State) time.Time { _, _ = SetOff(st, at(t, "10:00"), day, Rest); return at(t, "10:00") }, ""},
		{"done", func(st *State) time.Time { Add(st, at(t, "10:00"), -1, 100, "t"); return at(t, "10:01") }, "✓"},
		{"session", func(st *State) time.Time {
			Add(st, at(t, "10:00"), -1, 37, "t")
			StartSession(st, at(t, "10:00"), 50)
			return at(t, "10:26")
		}, "37 · 24m"},
		{"session3", func(st *State) time.Time {
			_, _ = SetDay(st, at(t, "10:00"), day, Math, map[string]int{Math: 150})
			Add(st, at(t, "10:00"), -1, 120, "t")
			StartSession(st, at(t, "10:00"), 90)
			return at(t, "10:01")
		}, "120·89m"},
		{"test session", func(st *State) time.Time {
			_, _ = SetDay(st, at(t, "10:00"), day, Test, nil)
			StartSession(st, at(t, "10:00"), 0)
			return at(t, "10:54")
		}, "test 96m"},
		{"pending", func(st *State) time.Time { Add(st, at(t, "10:00"), -1, 37, "t"); return at(t, "10:01") }, "37/100"},
		{"alert", func(st *State) time.Time { Add(st, at(t, "10:00"), -1, 37, "t"); return at(t, "22:00") }, "! 37/100"},
		{"alert wide", func(st *State) time.Time {
			_, _ = SetDay(st, at(t, "10:00"), day, Math, map[string]int{Math: 150})
			Add(st, at(t, "10:00"), -1, 120, "t")
			return at(t, "22:50")
		}, "!120/150"},
		{"test", func(st *State) time.Time { _, _ = SetDay(st, at(t, "10:00"), day, Test, nil); return at(t, "10:01") }, "test"},
		{"test alert", func(st *State) time.Time { _, _ = SetDay(st, at(t, "10:00"), day, Test, nil); return at(t, "21:00") }, "! test"},
		{"countdown", func(st *State) time.Time { st.Settings.TrayText = "countdown"; return at(t, "10:00") }, "T-11"},
		{"none", func(st *State) time.Time { st.Settings.TrayText = "none"; return at(t, "10:00") }, ""},
	}
	for _, c := range cases {
		st := fresh(t, Math)
		now := c.prep(st)
		got := View(st, now).Tray.Text
		if got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
		if utf8.RuneCountInString(got) > 8 {
			t.Errorf("%s: %q is longer than eight characters", c.name, got)
		}
	}
}

// 28
func TestRemainingMinutes(t *testing.T) {
	for _, c := range []struct {
		task Task
		want int
	}{
		{Task{Kind: Math, Target: 100, Done: 37}, 101},
		{Task{Kind: English, Target: 100, Done: 50}, 60},
		{Task{Kind: Test, Target: 1}, 150},
	} {
		if got := RemainingMinutes(&Day{Tasks: []Task{c.task}}); got != c.want {
			t.Errorf("%+v: %d, want %d", c.task, got, c.want)
		}
	}
}

// 29
func TestDuration(t *testing.T) {
	for in, want := range map[int]string{101: "1 h 41 min", 60: "1 h", 40: "40 min"} {
		if got := Duration(in); got != want {
			t.Errorf("%d: %q", in, got)
		}
	}
}

// 30
func TestNormalizeZero(t *testing.T) {
	got := Normalize(Settings{})
	want := DefaultSettings()
	if fmt.Sprintf("%+v", got) != fmt.Sprintf("%+v", want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
}

// 31
func TestTwoTaskDay(t *testing.T) {
	st := fresh(t, "test+math")
	d := st.Days[day]
	if d.Active() != 0 {
		t.Fatalf("active %d", d.Active())
	}
	acts := MarkDone(st, at(t, "12:00"), -1, "test")
	if len(acts) != 0 || d.Active() != 1 {
		t.Fatalf("after the test: active %d, %+v", d.Active(), acts)
	}
	acts = Add(st, at(t, "18:00"), -1, 100, "test")
	if len(acts) != 1 || acts[0].Kind != "celebrate" || acts[0].Situation != "test" {
		t.Fatalf("both done: %+v", acts)
	}
}

func TestFinalCelebration(t *testing.T) {
	st := fresh(t, Math)
	date := "2026-12-03" // the last study day (the 4th is rest)
	acts := Add(st, atOn(t, date, "12:00"), -1, 100, "test")
	if len(acts) != 1 || acts[0].Situation != "final" {
		t.Fatalf("%+v", acts)
	}
}

func TestSetDayKeepsProgress(t *testing.T) {
	st := fresh(t, Math)
	Add(st, at(t, "10:00"), -1, 30, "test")
	if _, err := SetDay(st, at(t, "10:05"), day, "test+math", nil); err != nil {
		t.Fatal(err)
	}
	d := st.Days[day]
	if d.Tasks[1].Kind != Math || d.Tasks[1].Done != 30 || d.Entries[0].Task != 1 {
		t.Fatalf("%+v %+v", d.Tasks, d.Entries)
	}
	if _, err := SetDay(st, at(t, "10:05"), AddDays(day, -1), Math, nil); err != ErrNotToday {
		t.Fatalf("past day: %v", err)
	}
}

func TestOnboardingBuildsPlan(t *testing.T) {
	st := NewState()
	now := atOn(t, "2026-11-22", "15:00")
	View(st, now)
	CompleteOnboarding(st, now, 1450)
	snap := View(st, now)
	if snap.Phase != "study" || snap.DayCount != 13 || len(snap.Strip) != 14 || st.Settings.TargetScore != 1450 {
		t.Fatalf("phase %s count %d strip %d", snap.Phase, snap.DayCount, len(snap.Strip))
	}
	if snap.PlanSummary != "3 tests, 4 math, 4 English, 2 rest: 800 questions" {
		t.Fatalf("summary %q", snap.PlanSummary)
	}
}
