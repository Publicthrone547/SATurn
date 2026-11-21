package core

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

type TaskView struct {
	Kind     string  `json:"kind"`
	Label    string  `json:"label"`
	Target   int     `json:"target"`
	Done     int     `json:"done"`
	Fraction float64 `json:"fraction"`
	QuickAdd [4]int  `json:"quickAdd"`
	IsTest   bool    `json:"isTest"`
}

type EntryView struct {
	Time   string `json:"time"`
	Text   string `json:"text"`
	Source string `json:"source"`
}

type ChangeView struct {
	Time string `json:"time"`
	Text string `json:"text"`
}

type DayView struct {
	Date          string       `json:"date"`
	DateLabel     string       `json:"dateLabel"`
	Label         string       `json:"label"`
	Preset        string       `json:"preset"`
	Off           string       `json:"off"`
	Status        string       `json:"status"`
	Tasks         []TaskView   `json:"tasks"`
	Active        int          `json:"active"`
	Done          bool         `json:"done"`
	Remaining     int          `json:"remaining"`
	RemainingText string       `json:"remainingText"`
	Entries       []EntryView  `json:"entries"`
	Changes       []ChangeView `json:"changes"`
	Score         int          `json:"score"`
	Tomorrow      string       `json:"tomorrow"`
	CanUndo       bool         `json:"canUndo"`
}

type StripDay struct {
	Date      string       `json:"date"`
	DateLabel string       `json:"dateLabel"`
	Weekday   string       `json:"weekday"`
	Label     string       `json:"label"`
	Preset    string       `json:"preset"`
	Off       string       `json:"off"`
	Status    string       `json:"status"`
	Fraction  float64      `json:"fraction"`
	IsToday   bool         `json:"isToday"`
	IsExam    bool         `json:"isExam"`
	Editable  bool         `json:"editable"`
	Tasks     []TaskView   `json:"tasks"`
	Summary   string       `json:"summary"`
	DoneText  string       `json:"doneText"`
	Changes   []ChangeView `json:"changes"`
	Score     int          `json:"score"`
}

type TrayView struct {
	Text     string  `json:"text"`
	Glyph    string  `json:"glyph"` // ring | done | off | exam
	Progress float64 `json:"progress"`
	Alert    bool    `json:"alert"`
	Session  bool    `json:"session"`
	Status   string  `json:"status"` // the menu's status line
}

type PnrView struct {
	Clock       string  `json:"clock"`
	MinutesLeft int     `json:"minutesLeft"`
	Red         bool    `json:"red"`
	Fit         int     `json:"fit"`
	FitText     string  `json:"fitText"`
	NowFrac     float64 `json:"nowFrac"`
	PnrFrac     float64 `json:"pnrFrac"`
	CanPromise  bool    `json:"canPromise"`
	PromiseMax  int     `json:"promiseMax"`
}

type SessionView struct {
	EndsAt       time.Time `json:"endsAt"`
	RemainingSec int       `json:"remainingSec"`
	TotalSec     int       `json:"totalSec"`
}

type PromiseView struct {
	Clock string `json:"clock"`
	Used  int    `json:"used"`
}

// SystemView is filled in by the shell.
type SystemView struct {
	Version       string   `json:"version"`
	Installed     bool     `json:"installed"`
	Dry           bool     `json:"dry"`
	Autostart     bool     `json:"autostart"`
	AutostartNote string   `json:"autostartNote"`
	Notifications bool     `json:"notifications"`
	HotkeyErrors  []string `json:"hotkeyErrors"`
	DataDir       string   `json:"dataDir"`
	ReduceMotion  bool     `json:"reduceMotion"`
}

type Snapshot struct {
	Phase       string       `json:"phase"` // setup | study | exam | after
	Today       string       `json:"today"`
	Now         string       `json:"now"` // the app clock, RFC 3339
	Clock       string       `json:"clock"`
	DayNumber   int          `json:"dayNumber"`
	DayCount    int          `json:"dayCount"`
	DaysLeft    int          `json:"daysLeft"`
	Day         DayView      `json:"day"`
	Strip       []StripDay   `json:"strip"`
	Streak      int          `json:"streak"`
	Total       int          `json:"total"`
	TestsDone   int          `json:"testsDone"`
	Totals      string       `json:"totals"`
	PlanSummary string       `json:"planSummary"`
	Tray        TrayView     `json:"tray"`
	Pnr         PnrView      `json:"pnr"`
	Session     *SessionView `json:"session"`
	Promise     *PromiseView `json:"promise"`
	CanQuit     bool         `json:"canQuit"`
	Settings    Settings     `json:"settings"`
	System      SystemView   `json:"system"`
}

func taskViews(st *State, d *Day) []TaskView {
	out := []TaskView{}
	for _, t := range d.Tasks {
		f := 0.0
		if t.Target > 0 {
			f = float64(t.Done) / float64(t.Target)
		}
		out = append(out, TaskView{
			Kind: t.Kind, Label: KindLabel(t.Kind), Target: t.Target, Done: t.Done,
			Fraction: f, QuickAdd: st.Settings.QuickAdd[t.Kind], IsTest: t.Kind == Test,
		})
	}
	return out
}

func changeViews(cs []Change) []ChangeView {
	out := []ChangeView{}
	for _, c := range cs {
		out = append(out, ChangeView{Time: ClockOf(c.At), Text: c.Text})
	}
	return out
}

func remainingText(d *Day) string {
	if len(d.Tasks) == 0 {
		return ""
	}
	t := d.Tasks[d.Active()]
	if d.IsDone() {
		return "Done for today"
	}
	if t.Kind == Test {
		return "One full practice test, about " + Duration(RemainingMinutes(d))
	}
	return fmt.Sprintf("%d left, about %s", t.Target-t.Done, Duration(RemainingMinutes(d)))
}

func doneText(d *Day) string {
	if d.Off != "" || len(d.Tasks) == 0 {
		return ""
	}
	var parts []string
	for _, t := range d.Tasks {
		if t.Kind == Test {
			if t.Done >= t.Target {
				parts = append(parts, "test done")
			} else {
				parts = append(parts, "test not done")
			}
		} else {
			parts = append(parts, fmt.Sprintf("%d of %d", t.Done, t.Target))
		}
	}
	return strings.Join(parts, ", ")
}

func frac(m int, s Settings) float64 {
	f := float64(m-s.RemindFrom) / float64(s.RemindUntil-s.RemindFrom)
	return min(1, max(0, f))
}

// View builds everything the interface shows. It may create missing plan
// days but never changes reminder bookkeeping.
func View(st *State, now time.Time) Snapshot {
	s := st.Settings
	today := st.Today(now)
	st.ensurePlan(today)
	snap := Snapshot{Today: today, Now: now.Format(time.RFC3339), Clock: ClockOf(now), Settings: s}
	switch {
	case !s.Onboarded:
		snap.Phase = "setup"
	case st.Retired || today > s.ExamDate:
		snap.Phase = "after"
	case today == s.ExamDate:
		snap.Phase = "exam"
	default:
		snap.Phase = "study"
	}
	snap.DayNumber, snap.DayCount = st.DayNumber(today)
	snap.DaysLeft = max(0, DaysBetween(today, s.ExamDate))
	snap.Streak = st.Streak(today)
	snap.Total = st.TotalQuestions()
	snap.TestsDone = st.TestsDone()
	snap.Totals = st.Totals()
	snap.PlanSummary = st.PlanSummary()

	d := st.day(today)
	dv := DayView{
		Date: today, DateLabel: DateLabel(today), Preset: PresetOf(d), Off: d.Off,
		Status: st.Status(d, today), Tasks: taskViews(st, d), Done: d.IsDone(),
		Changes: changeViews(d.Changes), Score: d.Score, CanUndo: len(d.Entries) > 0,
		Entries: []EntryView{},
	}
	if today == s.ExamDate {
		dv.Label = PresetLabel("exam")
	} else {
		dv.Label = PresetLabel(dv.Preset)
	}
	if len(d.Tasks) > 0 {
		dv.Active = d.Active()
		t := d.Tasks[dv.Active]
		dv.Remaining = max(0, t.Target-t.Done)
		dv.RemainingText = remainingText(d)
	}
	for i := len(d.Entries) - 1; i >= 0 && len(dv.Entries) < 50; i-- {
		e := d.Entries[i]
		kind := ""
		if e.Task >= 0 && e.Task < len(d.Tasks) {
			kind = " " + lower(KindLabel(d.Tasks[e.Task].Kind))
		}
		dv.Entries = append(dv.Entries, EntryView{Time: ClockOf(e.At), Text: fmt.Sprintf("%+d%s", e.Delta, kind), Source: e.Source})
	}
	if next := AddDays(today, 1); next < s.ExamDate {
		nd := st.day(next)
		if nd.Off != "" {
			dv.Tomorrow = "Tomorrow: " + lower(PresetLabel(nd.Off))
		} else {
			dv.Tomorrow = "Tomorrow: " + Summary(nd)
		}
	} else if next == s.ExamDate {
		dv.Tomorrow = "Tomorrow: the SAT"
	}
	snap.Day = dv

	for _, date := range st.PlanDates() {
		pd := st.day(date)
		snap.Strip = append(snap.Strip, StripDay{
			Date: date, DateLabel: DateLabel(date), Weekday: Weekday(date),
			Label: PresetLabel(PresetOf(pd)), Preset: PresetOf(pd), Off: pd.Off,
			Status: st.Status(pd, today), Fraction: pd.Fraction(), IsToday: date == today,
			Editable: date >= today, Tasks: taskViews(st, pd), Summary: capFirst(Summary(pd)),
			DoneText: doneText(pd), Changes: changeViews(pd.Changes), Score: pd.Score,
		})
	}
	if s.ExamDate != "" {
		snap.Strip = append(snap.Strip, StripDay{
			Date: s.ExamDate, DateLabel: DateLabel(s.ExamDate), Weekday: Weekday(s.ExamDate),
			Label: "Exam day", Preset: "exam", Status: StatusExam, IsExam: true,
			IsToday: s.ExamDate == today, Tasks: []TaskView{}, Summary: "The SAT", Changes: []ChangeView{},
		})
	}

	m := MinuteOfDay(now)
	pending := snap.Phase == "study" && d.Off == "" && len(d.Tasks) > 0 && !d.IsDone()
	if len(d.Tasks) > 0 {
		pnr := st.Pnr(d)
		sameDay := now.Format(dateLayout) == today
		mm := m
		if !sameDay {
			mm = m + 1440
		}
		snap.Pnr = PnrView{
			Clock: Clock(pnr), MinutesLeft: pnr - mm, Red: pending && mm >= pnr,
			Fit: st.Fit(d, mm), NowFrac: frac(mm, s), PnrFrac: frac(pnr, s),
		}
		if snap.Pnr.Red {
			t := d.Tasks[d.Active()]
			if t.Kind == Test {
				snap.Pnr.FitText = "A full test no longer fits before " + Clock(s.RemindUntil)
			} else {
				snap.Pnr.FitText = fmt.Sprintf("%d of %d still fit before %s", snap.Pnr.Fit, t.Target, Clock(s.RemindUntil))
			}
		}
		snap.Pnr.PromiseMax = max(0, pnr-mm)
		snap.Pnr.CanPromise = pending && d.Session == nil && d.PromisesUsed < 3 && snap.Pnr.PromiseMax >= 15
	}
	if d.Session != nil && now.Before(d.Session.EndsAt) {
		snap.Session = &SessionView{
			EndsAt:       d.Session.EndsAt,
			RemainingSec: int(d.Session.EndsAt.Sub(now).Seconds()),
			TotalSec:     int(d.Session.EndsAt.Sub(d.Session.StartedAt).Seconds()),
		}
	}
	if d.Promise != nil {
		snap.Promise = &PromiseView{Clock: ClockOf(d.Promise.At), Used: d.PromisesUsed}
	}
	snap.CanQuit = !s.Strict || !pending
	snap.Tray = trayView(st, d, snap, now)
	return snap
}

func fitTray(s string) string {
	if utf8.RuneCountInString(s) > 8 {
		s = strings.ReplaceAll(s, " ", "")
	}
	return s
}

func trayView(st *State, d *Day, snap Snapshot, now time.Time) TrayView {
	tv := TrayView{Glyph: "ring"}
	switch snap.Phase {
	case "setup":
		tv.Text, tv.Status = "setup", "Finish setup to start"
		return tv
	case "exam":
		tv.Text, tv.Glyph, tv.Status = "SAT", "exam", "Exam day. Go."
		return tv
	case "after":
		tv.Glyph, tv.Status, tv.Progress = "done", "The plan is over", 1
		return tv
	}
	if d.Off != "" {
		tv.Glyph, tv.Status = "off", PresetLabel(d.Off)
		return tv
	}
	if len(d.Tasks) == 0 {
		tv.Glyph, tv.Status = "off", "Nothing planned"
		return tv
	}
	t := d.Tasks[d.Active()]
	tv.Progress = min(1, d.Fraction())
	if d.IsDone() {
		tv.Text, tv.Glyph, tv.Progress = "✓", "done", 1
		tv.Status = PresetLabel(PresetOf(d)) + ": done"
		return tv
	}
	if t.Kind == Test {
		tv.Status = PresetLabel(PresetOf(d)) + ": test not done"
	} else {
		tv.Status = fmt.Sprintf("%s: %d of %d", PresetLabel(PresetOf(d)), t.Done, t.Target)
	}
	if snap.Session != nil {
		tv.Session = true
		left := (snap.Session.RemainingSec + 59) / 60
		if t.Kind == Test {
			tv.Text = fitTray(fmt.Sprintf("test %dm", left))
		} else {
			tv.Text = fitTray(fmt.Sprintf("%d · %dm", t.Done, left))
		}
		return tv
	}
	tv.Alert = st.CurrentLevel(d, now) == 3
	switch st.Settings.TrayText {
	case "countdown":
		tv.Text = fmt.Sprintf("T-%d", snap.DaysLeft)
	case "none":
		tv.Text = ""
	default:
		if t.Kind == Test {
			tv.Text = "test"
		} else {
			tv.Text = fmt.Sprintf("%d/%d", t.Done, t.Target)
		}
		if tv.Alert {
			tv.Text = fitTray("! " + tv.Text)
		}
	}
	return tv
}
