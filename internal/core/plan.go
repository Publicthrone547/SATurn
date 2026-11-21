package core

import (
	"fmt"
	"sort"
	"strings"
)

// Presets are the day shapes the plan offers.
var Presets = []string{Math, English, Test, "test+math", "test+english", Rest, Sick}

// presetKinds lists the tasks of a study preset, in order.
func presetKinds(preset string) []string {
	switch preset {
	case Math, English, Test:
		return []string{preset}
	case "test+math":
		return []string{Test, Math}
	case "test+english":
		return []string{Test, English}
	}
	return nil
}

// IsPreset reports whether p is one of the presets.
func IsPreset(p string) bool {
	for _, q := range Presets {
		if q == p {
			return true
		}
	}
	return false
}

// proposal is the preset for k days before the exam (k = 1 is the day before).
var proposal = [14]string{"", Rest, Math, Test, English, Math, Rest, Test, English, Math, English, Math, Test, English}

// ProposedPreset is the suggested preset for a date, counted back from the exam.
// It returns "" for the exam day and later.
func ProposedPreset(date, examDate string) string {
	k := DaysBetween(date, examDate)
	if k <= 0 {
		return ""
	}
	for k > 13 {
		k -= 7
	}
	return proposal[k]
}

// newTasks builds the tasks of a preset with the given targets (a missing
// target falls back to the settings).
func newTasks(preset string, targets map[string]int, s Settings) []Task {
	var out []Task
	for _, k := range presetKinds(preset) {
		t := targets[k]
		if t <= 0 {
			t = s.Targets[k]
		}
		if t <= 0 {
			t = 1
		}
		out = append(out, Task{Kind: k, Target: clamp(t, 1, 500)})
	}
	return out
}

// newDay creates a day from a preset.
func newDay(date, preset string, s Settings) *Day {
	d := &Day{Date: date}
	switch preset {
	case Rest, Sick:
		d.Off = preset
	default:
		d.Tasks = newTasks(preset, nil, s)
	}
	return d
}

// PresetOf is the preset that describes a day.
func PresetOf(d *Day) string {
	if d == nil {
		return ""
	}
	if d.Off != "" {
		return d.Off
	}
	var kinds []string
	for _, t := range d.Tasks {
		kinds = append(kinds, t.Kind)
	}
	return strings.Join(kinds, "+")
}

// LastPlanDay is the day before the exam.
func (st *State) LastPlanDay() string { return AddDays(st.Settings.ExamDate, -1) }

// PlanDates lists every date of the plan, FirstDay through the day before the exam.
func (st *State) PlanDates() []string {
	if st.FirstDay == "" {
		return nil
	}
	var out []string
	last := st.LastPlanDay()
	for d := st.FirstDay; d <= last; d = AddDays(d, 1) {
		out = append(out, d)
		if len(out) > 400 {
			break
		}
	}
	return out
}

// ensurePlan sets FirstDay when it is missing and creates every missing day
// of the plan from the proposal.
func (st *State) ensurePlan(today string) {
	if st.Days == nil {
		st.Days = map[string]*Day{}
	}
	if st.FirstDay == "" {
		st.FirstDay = today
	}
	for _, date := range st.PlanDates() {
		if st.Days[date] == nil {
			st.Days[date] = newDay(date, ProposedPreset(date, st.Settings.ExamDate), st.Settings)
		}
	}
}

// day returns the day for date, creating it from the proposal when missing.
// Dates on or after the exam get an empty day that is never stored.
func (st *State) day(date string) *Day {
	if d := st.Days[date]; d != nil {
		return d
	}
	if date >= st.Settings.ExamDate {
		return &Day{Date: date}
	}
	d := newDay(date, ProposedPreset(date, st.Settings.ExamDate), st.Settings)
	st.Days[date] = d
	return d
}

// IsDone reports whether every task is at or over its target.
func (d *Day) IsDone() bool {
	if len(d.Tasks) == 0 {
		return false
	}
	for _, t := range d.Tasks {
		if t.Done < t.Target {
			return false
		}
	}
	return true
}

// Active is the index of the first task not at its target, or the last task.
func (d *Day) Active() int {
	for i, t := range d.Tasks {
		if t.Done < t.Target {
			return i
		}
	}
	return len(d.Tasks) - 1
}

func (d *Day) hasTest() bool {
	for _, t := range d.Tasks {
		if t.Kind == Test {
			return true
		}
	}
	return false
}

// doneTotals sums done (capped at target) and target over the tasks.
func (d *Day) doneTotals() (done, target int) {
	for _, t := range d.Tasks {
		done += min(t.Done, t.Target)
		target += t.Target
	}
	return
}

// Fraction is the share of the day's work done, 0..1 (tasks weighted equally).
func (d *Day) Fraction() float64 {
	if len(d.Tasks) == 0 {
		return 0
	}
	var f float64
	for _, t := range d.Tasks {
		if t.Target > 0 {
			f += min(1, float64(t.Done)/float64(t.Target))
		}
	}
	return f / float64(len(d.Tasks))
}

// Status is the day's derived status relative to today.
func (st *State) Status(d *Day, today string) string {
	switch {
	case d.Date == st.Settings.ExamDate:
		return StatusExam
	case d.Off == Rest:
		return StatusRest
	case d.Off == Sick:
		return StatusSick
	case d.IsDone():
		return StatusDone
	case d.Date == today:
		return StatusPending
	case d.Date < today:
		return StatusMissed
	}
	return StatusUpcoming
}

// Streak counts consecutive done days back from today, skipping rest and
// sick days. Today counts when done and never breaks the streak when pending.
func (st *State) Streak(today string) int {
	n := 0
	for date := today; date >= st.FirstDay && st.FirstDay != ""; date = AddDays(date, -1) {
		d := st.Days[date]
		if d == nil {
			break
		}
		switch st.Status(d, today) {
		case StatusDone:
			n++
		case StatusRest, StatusSick:
		case StatusPending:
		default:
			return n
		}
	}
	return n
}

// TotalQuestions is every question logged over the plan.
func (st *State) TotalQuestions() int {
	n := 0
	for _, d := range st.Days {
		for _, t := range d.Tasks {
			if t.Kind != Test {
				n += t.Done
			}
		}
	}
	return n
}

// TestsDone counts finished practice tests.
func (st *State) TestsDone() int {
	n := 0
	for _, d := range st.Days {
		for _, t := range d.Tasks {
			if t.Kind == Test && t.Done >= t.Target {
				n++
			}
		}
	}
	return n
}

// KindLabel is the plain name of a task kind.
func KindLabel(kind string) string {
	switch kind {
	case Math:
		return "Math"
	case English:
		return "English"
	case Test:
		return "Practice test"
	}
	return kind
}

// PresetLabel names a day: "Math day", "Practice test and math", "Rest day".
func PresetLabel(preset string) string {
	switch preset {
	case Math:
		return "Math day"
	case English:
		return "English day"
	case Test:
		return "Practice test day"
	case "test+math":
		return "Practice test and math"
	case "test+english":
		return "Practice test and English"
	case Rest:
		return "Rest day"
	case Sick:
		return "Sick day"
	case "exam":
		return "Exam day"
	}
	return "Study day"
}

// Summary describes what a day asks for: "English, 100 questions" or "a practice test".
func Summary(d *Day) string {
	if d == nil {
		return ""
	}
	if d.Off != "" {
		return PresetLabel(d.Off)
	}
	var parts []string
	for _, t := range d.Tasks {
		if t.Kind == Test {
			parts = append(parts, "a practice test")
		} else {
			parts = append(parts, fmt.Sprintf("%s, %d questions", KindLabel(t.Kind), t.Target))
		}
	}
	if len(parts) == 0 {
		return "Nothing planned"
	}
	return strings.Join(parts, " and ")
}

// capFirst upper-cases the first letter.
func capFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// PlanSummary counts the plan: "3 tests, 4 math, 4 English, 2 rest: 800 questions".
func (st *State) PlanSummary() string {
	var tests, maths, engl, rest, questions int
	for _, date := range st.PlanDates() {
		d := st.Days[date]
		if d == nil {
			continue
		}
		if d.Off != "" {
			rest++
			continue
		}
		for _, t := range d.Tasks {
			switch t.Kind {
			case Test:
				tests++
			case Math:
				maths++
				questions += t.Target
			case English:
				engl++
				questions += t.Target
			}
		}
	}
	var parts []string
	add := func(n int, one, many string) {
		if n == 1 {
			parts = append(parts, "1 "+one)
		} else if n > 1 {
			parts = append(parts, fmt.Sprintf("%d %s", n, many))
		}
	}
	add(tests, "test", "tests")
	add(maths, "math", "math")
	add(engl, "English", "English")
	add(rest, "rest", "rest")
	return fmt.Sprintf("%s: %d questions", strings.Join(parts, ", "), questions)
}

// sortedDates returns the stored dates in order.
func (st *State) sortedDates() []string {
	out := make([]string, 0, len(st.Days))
	for k := range st.Days {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
