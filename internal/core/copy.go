package core

import (
	"fmt"
	"time"
)

// Every user-facing reminder text is built here (IMPLEMENTATION.md 7.6).

var zeroLines = []string{
	"The first quarter is the hardest one to start.",
	"Open the first question. That is the whole job right now.",
	"One session now and the evening gets lighter.",
}

var progressLines = []string{
	"Keep the run going.",
	"%d more closes this quarter.",
	"Every one you log moves the deadline later.",
}

// DayNumber is today's position in the plan and the plan's length.
func (st *State) DayNumber(today string) (n, count int) {
	if st.FirstDay == "" {
		return 1, max(1, DaysBetween(today, st.Settings.ExamDate))
	}
	count = max(1, DaysBetween(st.FirstDay, st.Settings.ExamDate))
	n = clamp(DaysBetween(st.FirstDay, today)+1, 1, count)
	return n, count
}

func (st *State) sub(d *Day) string {
	n, c := st.DayNumber(d.Date)
	return fmt.Sprintf("Day %d of %d", n, c)
}

func baseData(st *State, d *Day) map[string]any {
	data := map[string]any{"date": d.Date}
	if len(d.Tasks) > 0 {
		t := d.Tasks[d.Active()]
		data["kind"] = t.Kind
		data["isTest"] = t.Kind == Test
		data["done"] = t.Done
		data["target"] = t.Target
	}
	return data
}

func rotate(lines []string, now time.Time, s Settings) (string, int) {
	slot := max(0, (MinuteOfDay(now)-s.RemindFrom)/s.IntervalMin)
	i := (now.YearDay() + slot) % len(lines)
	return lines[i], i
}

// toNextQuarter is how many units close the current quarter.
func toNextQuarter(t Task) int {
	for q := 1; q <= 4; q++ {
		if th := ceilDiv(t.Target*q, 4); th > t.Done {
			return th - t.Done
		}
	}
	return 0
}

func remindAction(st *State, d *Day, now time.Time, situation string, level int) Action {
	s := st.Settings
	t := d.Tasks[d.Active()]
	m := MinuteOfDay(now)
	pnr := Clock(st.Pnr(d))
	rem := max(0, t.Target-t.Done)
	dur := Duration(RemainingMinutes(d))
	a := Action{Kind: "remind", Level: level, Situation: situation, Sub: st.sub(d), Data: baseData(st, d)}

	if t.Kind == Test {
		switch situation {
		case "brief":
			a.Title, a.Body = "Practice test day.", "Start by "+pnr+" to finish today."
			a = briefVariants(st, d, a, pnr)
		case "lastStart":
			a.Title, a.Body = "Last comfortable start: "+pnr+".", "After that, today does not fit."
		case "red":
			a.Title, a.Body = "A full test no longer fits before "+Clock(s.RemindUntil)+".", "Start anyway, or mark the day."
		case "promise":
			a.Title, a.Body = ClockOf(now)+". You said you would start now.", "The practice test, about "+dur+"."
		case "broken":
			a.Body = "The test is still open. Start by " + pnr + "."
		default:
			a.Title, a.Body = "The practice test is still open. Start by "+pnr+".", "One full test, about "+dur+"."
		}
		return a
	}

	kind := KindLabel(t.Kind)
	switch situation {
	case "brief":
		a.Title = fmt.Sprintf("%s day. %d questions.", kind, t.Target)
		a.Body = "Start by " + pnr + " to finish today."
		a = briefVariants(st, d, a, pnr)
	case "zero":
		a.Title = fmt.Sprintf("0 of %d. Start by %s.", t.Target, pnr)
		a.Body, _ = rotate(zeroLines, now, s)
	case "progress":
		a.Title = fmt.Sprintf("%d of %d. %d left, about %s.", t.Done, t.Target, rem, dur)
		line, i := rotate(progressLines, now, s)
		if i == 1 {
			line = fmt.Sprintf(line, toNextQuarter(t))
		}
		a.Body = line
	case "lastQuarter":
		a.Title = fmt.Sprintf("%d of %d. %d left.", t.Done, t.Target, rem)
		a.Body = "About " + dur + "."
	case "lastStart":
		a.Title, a.Body = "Last comfortable start: "+pnr+".", "After that, today does not fit."
	case "red":
		fit := st.Fit(d, m)
		a.Title = fmt.Sprintf("%s. %d of %d still fit.", ClockOf(now), fit, t.Target)
		a.Body = fmt.Sprintf("Do the %d.", fit)
		a.Data["fit"] = fit
	case "promise":
		a.Title = ClockOf(now) + ". You said you would start now."
		a.Body = fmt.Sprintf("%d left, about %s.", rem, dur)
	case "broken":
		a.Body = fmt.Sprintf("%d left. Start by %s.", rem, pnr)
	}
	return a
}

func briefVariants(st *State, d *Day, a Action, pnr string) Action {
	y := st.Days[AddDays(d.Date, -1)]
	if y == nil || y.Date < st.FirstDay {
		return a
	}
	switch {
	case y.Off == Sick:
		a.Body = "Still sick? Use Day off. Otherwise start by " + pnr + "."
		a.Data["afterSick"] = true
	case y.Off == "" && len(y.Tasks) > 0 && !y.IsDone():
		yd, yt := y.doneTotals()
		a.Body = fmt.Sprintf("Yesterday closed at %d of %d. New day.", yd, yt)
		a.Data["afterMissed"] = true
	}
	return a
}

func brokenTitle(promised, now time.Time) string {
	return fmt.Sprintf("You said %s. It is %s.", ClockOf(promised), ClockOf(now))
}

func askAction(st *State, d *Day, now time.Time) Action {
	a := Action{Kind: "ask", Situation: "ask", Sub: st.sub(d), Data: baseData(st, d)}
	t := d.Tasks[d.Active()]
	if t.Kind == Test {
		a.Title, a.Body = "Is the test finished?", "Mark it when it is."
	} else {
		a.Title, a.Body = "How many did you solve?", fmt.Sprintf("%d of %d so far.", t.Done, t.Target)
	}
	return a
}

func notifyAction(st *State, d *Day, now time.Time, situation string) Action {
	a := Action{Kind: "notify", Situation: situation, Sub: st.sub(d), Data: map[string]any{"date": d.Date}}
	switch situation {
	case "eve":
		a.Title, a.Body = "Tomorrow is the SAT.", "Pack your ID, admission ticket, laptop and charger. Then sleep."
	case "closing":
		until := Clock(st.Settings.RemindUntil)
		roll := Clock(st.Settings.RolloverHour * 60)
		t := d.Tasks[d.Active()]
		if t.Kind == Test {
			a.Title = until + ". The practice test is still open."
		} else {
			a.Title = fmt.Sprintf("%s. Today ends with %d of %d.", until, t.Done, t.Target)
		}
		a.Body = "Log what you did before " + roll + "."
	}
	return a
}

func milestoneAction(st *State, d *Day, now time.Time, task, quarter int) Action {
	t := d.Tasks[task]
	bodies := map[int]string{1: "First quarter closed.", 2: "Halfway.", 3: "One quarter to go."}
	a := Action{Kind: "milestone", Situation: fmt.Sprintf("q%d", quarter), Sub: st.sub(d), Data: baseData(st, d)}
	a.Title = fmt.Sprintf("%d down.", t.Done)
	a.Body = bodies[quarter]
	a.Data["quarter"] = quarter
	a.Data["id"] = fmt.Sprintf("%s-q%d-%d", d.Date, quarter, now.Unix())
	return a
}

// isFinalDay reports whether no later plan day still has work.
func (st *State) isFinalDay(d *Day) bool {
	for date := AddDays(d.Date, 1); date <= st.LastPlanDay(); date = AddDays(date, 1) {
		if n := st.day(date); n.Off == "" && len(n.Tasks) > 0 {
			return false
		}
	}
	return true
}

func celebrateAction(st *State, d *Day, now time.Time) Action {
	kind := "day"
	switch {
	case st.isFinalDay(d):
		kind = "final"
	case d.hasTest():
		kind = "test"
	}
	n, _ := st.DayNumber(d.Date)
	streak := st.Streak(d.Date)
	tomorrow := ""
	if next := AddDays(d.Date, 1); next < st.Settings.ExamDate {
		tomorrow = "Tomorrow: " + Summary(st.day(next)) + "."
		if t := st.day(next); t.Off != "" {
			tomorrow = "Tomorrow: " + lower(PresetLabel(t.Off)) + "."
		}
	} else {
		tomorrow = "Tomorrow: the SAT."
	}
	a := Action{Kind: "celebrate", Situation: kind, Sub: st.sub(d), Data: baseData(st, d)}
	a.Title = fmt.Sprintf("Day %d complete", n)
	a.Body = fmt.Sprintf("Streak %d. %s", streak, tomorrow)
	a.Data["celebration"] = kind
	a.Data["streak"] = streak
	a.Data["dayNumber"] = n
	a.Data["tomorrow"] = tomorrow
	a.Data["id"] = fmt.Sprintf("%s-done-%d", d.Date, now.Unix())
	if kind == "final" {
		a.Data["totals"] = st.Totals()
	}
	return a
}

// Totals is the plan in one line: "13 days. 800 questions. 3 tests."
func (st *State) Totals() string {
	_, c := st.DayNumber(st.LastPlanDay())
	return fmt.Sprintf("%d days. %d questions. %d tests.", c, st.TotalQuestions(), st.TestsDone())
}

func welcomeAction(st *State, d *Day, now time.Time) Action {
	t := d.Tasks[d.Active()]
	a := Action{Kind: "welcome", Situation: "welcome", Sub: st.sub(d), Data: baseData(st, d)}
	pnr := Clock(st.Pnr(d))
	if t.Kind == Test {
		a.Title, a.Body = "Practice test day", "Not done yet · start by "+pnr
	} else {
		a.Title = fmt.Sprintf("%d of %d", t.Done, t.Target)
		a.Body = fmt.Sprintf("%d left · start by %s", max(0, t.Target-t.Done), pnr)
	}
	return a
}

func examAction(st *State) Action {
	return Action{Kind: "exam", Situation: "exam", Title: "Today.", Body: "You did the work. Go."}
}

func retireAction(st *State, now time.Time) Action {
	return Action{Kind: "retire", Situation: "retire", Title: "That was the plan.", Body: st.Totals()}
}
