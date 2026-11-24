package core

import "time"

// TickInput is what the shell knows at each tick.
type TickInput struct {
	Now         time.Time
	IdleSeconds int  // seconds since the last key press or mouse move
	Returned    bool // the owner just came back, or the Mac just woke
}

// paceTenths is the exam pace in tenths of a minute per unit.
func paceTenths(kind string) int {
	switch kind {
	case Math:
		return 16
	case English:
		return 12
	case Test:
		return 1500
	}
	return 16
}

// RemainingMinutes is how long the rest of the day takes at exam pace,
// rounded up to the minute. Integer arithmetic on purpose, so the result
// never depends on floating point.
func RemainingMinutes(d *Day) int {
	tenths := 0
	for _, t := range d.Tasks {
		tenths += max(0, t.Target-t.Done) * paceTenths(t.Kind)
	}
	return (tenths + 9) / 10
}

// Pnr is the point of no return in minutes after midnight.
func (st *State) Pnr(d *Day) int { return st.Settings.RemindUntil - RemainingMinutes(d) }

// PnrTime is the point of no return as a time on the day's calendar date.
func (st *State) PnrTime(d *Day) time.Time { return AtMinute(d.Date, st.Pnr(d)) }

// Fit is how many of the active task's units still fit before the window end.
func (st *State) Fit(d *Day, m int) int {
	if len(d.Tasks) == 0 {
		return 0
	}
	t := d.Tasks[d.Active()]
	left := max(0, st.Settings.RemindUntil-m)
	if t.Kind == Test {
		if left >= 150 {
			return 1
		}
		return min(t.Target, t.Done)
	}
	return min(t.Target, t.Done+(left*10)/paceTenths(t.Kind))
}

// Level is the escalation level for the next reminder.
func Level(d *Day, red bool) int {
	switch {
	case red || d.Ignored >= 4:
		return 3
	case d.Ignored >= 2:
		return 2
	}
	return 1
}

// Situation picks the wording of an hourly reminder.
func Situation(d *Day, red bool) string {
	if red {
		return "red"
	}
	t := d.Tasks[d.Active()]
	switch {
	case t.Done == 0:
		return "zero"
	case t.Target-t.Done <= ceilDiv(t.Target, 4):
		return "lastQuarter"
	}
	return "progress"
}

// closeDays closes every past day that is still open.
func (st *State) closeDays(today string) {
	for date, d := range st.Days {
		if date >= today || d.Closed != "" {
			continue
		}
		switch {
		case d.Off != "":
			d.Closed = "off"
		case d.IsDone():
			d.Closed = "done"
		case len(d.Tasks) == 0:
			d.Closed = "off"
		default:
			d.Closed = "missed"
		}
	}
}

// Tick is the reminder engine. The shell calls it every 20 seconds; it
// updates the bookkeeping and returns at most one action. The order of the
// steps is the specification.
func Tick(st *State, in TickInput) []Action {
	s := st.Settings
	now := in.Now
	// 1, 2
	today := LogicalDate(now, s.RolloverHour)
	st.closeDays(today)
	// 3
	if !s.Onboarded || st.Retired {
		return nil
	}
	// 4
	if today > s.ExamDate {
		if !st.RetireShown {
			st.RetireShown = true
			return []Action{retireAction(st, now)}
		}
		return nil
	}
	if today == s.ExamDate {
		if !st.ExamGreeted && now.Format(dateLayout) == today && MinuteOfDay(now) >= 6*60 {
			st.ExamGreeted = true
			return []Action{examAction(st)}
		}
		return nil
	}
	// 5
	st.ensurePlan(today)
	d := st.day(today)
	m := MinuteOfDay(now)
	sameDay := now.Format(dateLayout) == today
	if today == st.LastPlanDay() && sameDay && m >= 21*60 && !d.EveSent {
		d.EveSent = true
		return []Action{notifyAction(st, d, now, "eve")}
	}
	// 6
	if d.Off != "" || d.IsDone() || len(d.Tasks) == 0 {
		return nil
	}
	// 7
	if d.Session != nil {
		if now.Before(d.Session.EndsAt) {
			return nil
		}
		d.Session = nil
		d.AskUntil = now.Add(10 * time.Minute)
		return []Action{askAction(st, d, now)}
	}
	// 8
	if now.Before(d.AskUntil) {
		return nil
	}
	// 9
	if !sameDay {
		return nil
	}
	if m >= s.RemindUntil {
		if !d.ClosingSent {
			d.ClosingSent = true
			return []Action{notifyAction(st, d, now, "closing")}
		}
		return nil
	}
	if m < min(s.BriefAt, s.RemindFrom) {
		return nil
	}
	// 10
	if in.IdleSeconds > 300 {
		return nil
	}
	// 11
	if !d.BriefSent && m >= s.BriefAt {
		d.BriefSent = true
		d.LastRemindAt = now
		if m >= s.RemindFrom {
			slot := s.RemindFrom + ((m-s.RemindFrom)/s.IntervalMin)*s.IntervalMin
			d.LastSlot = AtMinute(today, slot)
		}
		return []Action{remindAction(st, d, now, "brief", 1)}
	}
	// 12
	if p := d.Promise; p != nil {
		if now.Before(p.At) {
			return nil
		}
		if !p.Fired {
			p.Fired = true
			d.LastRemindAt = now
			return []Action{remindAction(st, d, now, "promise", 2)}
		}
		if now.Before(p.At.Add(10 * time.Minute)) {
			return nil
		}
		pat := p.At
		d.Promise = nil
		d.Ignored = max(d.Ignored, 4)
		d.LastRemindAt = now
		a := remindAction(st, d, now, "broken", 3)
		a.Title = brokenTitle(pat, now)
		return []Action{a}
	}
	// 13
	pnr := st.Pnr(d)
	red := m >= pnr
	if !d.PnrWarned && m >= s.RemindFrom && m >= pnr-30 && m < pnr {
		d.PnrWarned = true
		d.LastRemindAt = now
		return []Action{remindAction(st, d, now, "lastStart", max(2, Level(d, red)))}
	}
	// 14
	if m < s.RemindFrom {
		return welcome(st, d, in)
	}
	// 15
	interval := s.IntervalMin
	if red {
		interval = s.RedIntervalMin
	}
	slot := s.RemindFrom + ((m-s.RemindFrom)/interval)*interval
	slotTime := AtMinute(today, slot)
	if slotTime.After(d.LastSlot) {
		d.LastSlot = slotTime
		if now.Sub(d.LastRemindAt) < 20*time.Minute {
			return nil
		}
		if now.Sub(d.LastProgressAt) < 20*time.Minute {
			return nil
		}
		if !d.LastRemindAt.IsZero() && d.LastResponseAt.Before(d.LastRemindAt) &&
			now.Sub(d.LastRemindAt) <= time.Duration(interval+30)*time.Minute {
			d.Ignored++
		}
		d.LastRemindAt = now
		return []Action{remindAction(st, d, now, Situation(d, red), Level(d, red))}
	}
	// 16
	return welcome(st, d, in)
}

func welcome(st *State, d *Day, in TickInput) []Action {
	if in.Returned && in.Now.Sub(d.LastIslandAt) >= 20*time.Minute && in.Now.Sub(d.LastRemindAt) >= 20*time.Minute {
		d.LastIslandAt = in.Now
		return []Action{welcomeAction(st, d, in.Now)}
	}
	return nil
}

// CurrentLevel is the level a reminder would have right now (for the tray's alert style).
func (st *State) CurrentLevel(d *Day, now time.Time) int {
	if len(d.Tasks) == 0 || d.Off != "" || d.IsDone() {
		return 0
	}
	red := now.Format(dateLayout) == d.Date && MinuteOfDay(now) >= st.Pnr(d)
	return Level(d, red)
}

// PreviewRemind builds a reminder of the given level from the current state
// without changing it (Settings → Send a test reminder).
func PreviewRemind(st *State, now time.Time, level int) Action {
	cp := *st
	d := st.Days[st.Today(now)]
	if d == nil || len(d.Tasks) == 0 {
		d = newDay(st.Today(now), Math, st.Settings)
	}
	dd := *d
	situation := "zero"
	if len(dd.Tasks) > 0 {
		red := MinuteOfDay(now) >= cp.Pnr(&dd)
		situation = Situation(&dd, red)
	}
	a := remindAction(&cp, &dd, now, situation, clamp(level, 1, 3))
	a.Data["test"] = true
	return a
}
