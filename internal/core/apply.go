package core

import (
	"errors"
	"fmt"
	"time"
)

// Errors returned by commands.
var (
	ErrNoPromisesLeft = errors.New("three promises a day is the limit")
	ErrPastDeadline   = errors.New("that is later than the last start time")
	ErrNotToday       = errors.New("only today and later days can be changed")
	ErrNoSession      = errors.New("no session is running")
	ErrBadScore       = errors.New("a score is between 400 and 1600")
	ErrBadPreset      = errors.New("unknown day type")
)

// Today is the logical date of now.
func (st *State) Today(now time.Time) string {
	return LogicalDate(now, st.Settings.RolloverHour)
}

// TodayDay returns today's day, creating the plan as needed.
func (st *State) TodayDay(now time.Time) *Day {
	today := st.Today(now)
	st.ensurePlan(today)
	return st.day(today)
}

func (st *State) change(d *Day, now time.Time, text string) {
	d.Changes = append(d.Changes, Change{At: now, Text: text})
}

// Add logs delta units on task (task < 0 means the active task).
func Add(st *State, now time.Time, task, delta int, source string) []Action {
	d := st.TodayDay(now)
	if len(d.Tasks) == 0 || delta == 0 {
		return nil
	}
	if task < 0 || task >= len(d.Tasks) {
		task = d.Active()
	}
	t := &d.Tasks[task]
	before := t.Done
	t.Done = max(0, t.Done+delta)
	if t.Kind == Test {
		t.Done = min(t.Done, t.Target) // a test is done or not; it never goes over
	}
	applied := t.Done - before
	if applied == 0 {
		return nil
	}
	d.Entries = append(d.Entries, Entry{At: now, Task: task, Delta: applied, Source: source})
	if applied > 0 {
		d.LastProgressAt, d.LastResponseAt = now, now
		d.Ignored = 0
		d.AskUntil = time.Time{}
		d.Promise = nil
	}
	return progressActions(st, d, now, task, before, t.Done)
}

// progressActions returns the milestone or celebration a change earned.
func progressActions(st *State, d *Day, now time.Time, task, before, after int) []Action {
	if d.IsDone() {
		if d.CelebratedAt.IsZero() {
			d.CelebratedAt = now
			d.Session = nil
			return []Action{celebrateAction(st, d, now)}
		}
		return nil
	}
	d.CelebratedAt = time.Time{}
	t := d.Tasks[task]
	if t.Kind == Test || t.Target < 8 || after <= before {
		return nil
	}
	for q := 3; q >= 1; q-- {
		th := ceilDiv(t.Target*q, 4)
		if before < th && th <= after {
			return []Action{milestoneAction(st, d, now, task, q)}
		}
	}
	return nil
}

// SetExact sets a task's done count.
func SetExact(st *State, now time.Time, task, value int) []Action {
	d := st.TodayDay(now)
	if len(d.Tasks) == 0 {
		return nil
	}
	if task < 0 || task >= len(d.Tasks) {
		task = d.Active()
	}
	return Add(st, now, task, max(0, value)-d.Tasks[task].Done, "main")
}

// MarkDone brings a task to its target.
func MarkDone(st *State, now time.Time, task int, source string) []Action {
	d := st.TodayDay(now)
	if len(d.Tasks) == 0 {
		return nil
	}
	if task < 0 || task >= len(d.Tasks) {
		task = d.Active()
	}
	t := d.Tasks[task]
	if t.Target-t.Done <= 0 {
		return nil
	}
	return Add(st, now, task, t.Target-t.Done, source)
}

// Undo removes today's last entry.
func Undo(st *State, now time.Time) []Action {
	d := st.TodayDay(now)
	if len(d.Entries) == 0 {
		return nil
	}
	e := d.Entries[len(d.Entries)-1]
	d.Entries = d.Entries[:len(d.Entries)-1]
	if e.Task >= 0 && e.Task < len(d.Tasks) {
		d.Tasks[e.Task].Done = max(0, d.Tasks[e.Task].Done-e.Delta)
	}
	if !d.IsDone() {
		d.CelebratedAt = time.Time{}
	}
	return nil
}

// SessionLength is the default session for the active task.
func (st *State) SessionLength(d *Day) int {
	if len(d.Tasks) > 0 && d.Tasks[d.Active()].Kind == Test {
		return st.Settings.TestSessionMin
	}
	return st.Settings.SessionMin
}

// StartSession starts a quiet study session.
func StartSession(st *State, now time.Time, minutes int) []Action {
	d := st.TodayDay(now)
	if minutes <= 0 {
		minutes = st.SessionLength(d)
	}
	d.Session = &Session{StartedAt: now, EndsAt: now.Add(time.Duration(minutes) * time.Minute)}
	d.LastResponseAt = now
	d.Ignored = 0
	d.Promise = nil
	d.AskUntil = time.Time{}
	return nil
}

// EndSession ends the running session and asks how it went.
func EndSession(st *State, now time.Time) []Action {
	d := st.TodayDay(now)
	if d.Session == nil {
		return nil
	}
	d.Session = nil
	d.AskUntil = now.Add(10 * time.Minute)
	return []Action{askAction(st, d, now)}
}

// MakePromise records "I will start in N minutes".
func MakePromise(st *State, now time.Time, minutes int) ([]Action, error) {
	d := st.TodayDay(now)
	if d.PromisesUsed >= 3 {
		return nil, ErrNoPromisesLeft
	}
	at := now.Add(time.Duration(minutes) * time.Minute)
	if at.After(st.PnrTime(d)) {
		return nil, ErrPastDeadline
	}
	d.Promise = &Promise{At: at}
	d.PromisesUsed++
	d.LastResponseAt = now
	return nil, nil
}

// SetOff marks a day as rest or sick, or back to the plan with off == "".
func SetOff(st *State, now time.Time, date, off string) ([]Action, error) {
	today := st.Today(now)
	st.ensurePlan(today)
	if date == "" {
		date = today
	}
	if date < today {
		return nil, ErrNotToday
	}
	if off != "" && off != Rest && off != Sick {
		return nil, ErrBadPreset
	}
	if date >= st.Settings.ExamDate {
		return nil, ErrNotToday
	}
	d := st.day(date)
	if d.Off == off {
		return nil, nil
	}
	before := PresetOf(d)
	d.Off = off
	if off == "" && len(d.Tasks) == 0 {
		p := ProposedPreset(date, st.Settings.ExamDate)
		if presetKinds(p) == nil {
			p = Math
		}
		d.Tasks = newTasks(p, nil, st.Settings)
	}
	if date == today {
		d.Session, d.Promise, d.AskUntil = nil, nil, time.Time{}
		st.change(d, now, fmt.Sprintf("%s changed to %s", PresetLabel(before), lower(PresetLabel(PresetOf(d)))))
	}
	return nil, nil
}

// SetDay replaces a day's tasks with a preset (targets per kind are optional).
func SetDay(st *State, now time.Time, date, preset string, targets map[string]int) ([]Action, error) {
	today := st.Today(now)
	st.ensurePlan(today)
	if !IsPreset(preset) {
		return nil, ErrBadPreset
	}
	if date < today || date >= st.Settings.ExamDate {
		return nil, ErrNotToday
	}
	if preset == Rest || preset == Sick {
		return SetOff(st, now, date, preset)
	}
	d := st.day(date)
	before := PresetOf(d)
	old := d.Tasks
	d.Tasks = newTasks(preset, targets, st.Settings)
	d.Off = ""
	if date == today {
		// Keep what was already done on kinds that stay; remap the log.
		idx := map[string]int{}
		for i, t := range d.Tasks {
			idx[t.Kind] = i
		}
		for _, t := range old {
			if i, ok := idx[t.Kind]; ok {
				d.Tasks[i].Done = t.Done
			}
		}
		for i := range d.Entries {
			e := &d.Entries[i]
			if e.Task >= 0 && e.Task < len(old) {
				if j, ok := idx[old[e.Task].Kind]; ok {
					e.Task = j
				} else {
					e.Task = -1
				}
			}
		}
		if !d.IsDone() {
			d.CelebratedAt = time.Time{}
		}
		if after := PresetOf(d); after != before {
			st.change(d, now, fmt.Sprintf("%s changed to %s", PresetLabel(before), lower(PresetLabel(after))))
		}
		d.Session, d.Promise = nil, nil
	}
	return nil, nil
}

// SuggestPlan fills the plan from the proposal. Without overwrite only days
// that do not exist yet are created.
func SuggestPlan(st *State, now time.Time, overwrite bool) []Action {
	today := st.Today(now)
	st.ensurePlan(today)
	if !overwrite {
		return nil
	}
	for date := today; date <= st.LastPlanDay(); date = AddDays(date, 1) {
		p := ProposedPreset(date, st.Settings.ExamDate)
		if p == "" || PresetOf(st.day(date)) == p {
			continue
		}
		if _, err := SetDay(st, now, date, p, nil); err != nil {
			continue
		}
	}
	return nil
}

// UpdateSettings stores new settings. The onboarding flag is not editable here.
func UpdateSettings(st *State, now time.Time, s Settings) []Action {
	s.Onboarded = st.Settings.Onboarded
	s = Normalize(s)
	examChanged := s.ExamDate != st.Settings.ExamDate
	st.Settings = s
	if examChanged {
		for date, d := range st.Days {
			if date >= s.ExamDate && len(d.Entries) == 0 {
				delete(st.Days, date)
			}
		}
		st.ExamGreeted, st.RetireShown = false, false
	}
	st.ensurePlan(st.Today(now))
	return nil
}

// CompleteOnboarding ends first run.
func CompleteOnboarding(st *State, now time.Time, targetScore int) []Action {
	today := st.Today(now)
	st.Settings.Onboarded = true
	if targetScore != 0 {
		targetScore = clamp(targetScore, 400, 1600)
	}
	st.Settings.TargetScore = targetScore
	if st.FirstDay != today {
		for date, d := range st.Days {
			if date < today && len(d.Entries) == 0 {
				delete(st.Days, date)
			}
		}
		if st.FirstDay == "" || st.FirstDay < today {
			st.FirstDay = today
		}
	}
	st.ensurePlan(today)
	return nil
}

// SetScore records a practice test score (0 clears it).
func SetScore(st *State, now time.Time, date string, score int) error {
	if score != 0 && (score < 400 || score > 1600) {
		return ErrBadScore
	}
	if date == "" {
		date = st.Today(now)
	}
	d := st.Days[date]
	if d == nil {
		return ErrNotToday
	}
	d.Score = score
	return nil
}

// Bailout records that strict mode was overridden.
func Bailout(st *State, now time.Time) {
	st.change(st.TodayDay(now), now, "Quit with the bailout phrase")
}

// Retire marks the plan as over.
func Retire(st *State) { st.Retired = true }

func lower(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	if b[0] >= 'A' && b[0] <= 'Z' {
		b[0] += 'a' - 'A'
	}
	return string(b)
}
