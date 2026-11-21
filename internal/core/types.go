// Package core is SATurn's domain logic: the day model, the plan, the
// commands that change state and the reminder engine. It is pure: every
// function that needs the time takes it as an argument, and nothing here
// does I/O or knows about the UI.
package core

import "time"

// Task kinds.
const (
	Math    = "math"
	English = "english"
	Test    = "test"
)

// Off states.
const (
	Rest = "rest"
	Sick = "sick"
)

// Day statuses (derived, never stored).
const (
	StatusExam     = "exam"
	StatusRest     = "rest"
	StatusSick     = "sick"
	StatusDone     = "done"
	StatusPending  = "pending"
	StatusMissed   = "missed"
	StatusUpcoming = "upcoming"
)

type Task struct {
	Kind   string `json:"kind"` // "math" | "english" | "test"
	Target int    `json:"target"`
	Done   int    `json:"done"`
}

type Entry struct {
	At     time.Time `json:"at"`
	Task   int       `json:"task"` // index into Day.Tasks; -1 when the task was removed
	Delta  int       `json:"delta"`
	Source string    `json:"source"` // popover | menu | hotkey | notification | island | main
}

type Change struct {
	At   time.Time `json:"at"`
	Text string    `json:"text"`
}

type Session struct {
	StartedAt time.Time `json:"startedAt"`
	EndsAt    time.Time `json:"endsAt"`
}

type Promise struct {
	At    time.Time `json:"at"`
	Fired bool      `json:"fired"`
}

type Day struct {
	Date    string   `json:"date"` // logical date, "2026-11-23"
	Off     string   `json:"off"`  // "" | "rest" | "sick"
	Tasks   []Task   `json:"tasks"`
	Entries []Entry  `json:"entries"`
	Changes []Change `json:"changes"`
	Closed  string   `json:"closed"` // "" | "done" | "missed" | "off", set at rollover
	Score   int      `json:"score"`  // optional practice test score, 0 = none

	// Reminder bookkeeping. Persisted, so a restart never repeats a reminder.
	BriefSent      bool      `json:"briefSent"`
	ClosingSent    bool      `json:"closingSent"`
	PnrWarned      bool      `json:"pnrWarned"`
	EveSent        bool      `json:"eveSent"`
	Ignored        int       `json:"ignored"`
	PromisesUsed   int       `json:"promisesUsed"`
	LastSlot       time.Time `json:"lastSlot"`
	LastRemindAt   time.Time `json:"lastRemindAt"`
	LastResponseAt time.Time `json:"lastResponseAt"`
	LastProgressAt time.Time `json:"lastProgressAt"`
	LastIslandAt   time.Time `json:"lastIslandAt"`
	AskUntil       time.Time `json:"askUntil"`
	CelebratedAt   time.Time `json:"celebratedAt"`
	Session        *Session  `json:"session,omitempty"`
	Promise        *Promise  `json:"promise,omitempty"`
}

type State struct {
	Version     int             `json:"version"` // 1
	Settings    Settings        `json:"settings"`
	Days        map[string]*Day `json:"days"`
	FirstDay    string          `json:"firstDay"`
	ExamGreeted bool            `json:"examGreeted"`
	RetireShown bool            `json:"retireShown"`
	Retired     bool            `json:"retired"`
}

// Action is something the shell must do: show a reminder, celebrate, open a
// screen. Commands and Tick return them; the core never performs them.
type Action struct {
	Kind      string         `json:"kind"`  // remind | notify | ask | milestone | celebrate | welcome | exam | retire
	Level     int            `json:"level"` // 1..3 for remind
	Situation string         `json:"situation"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Sub       string         `json:"sub"`
	Data      map[string]any `json:"data,omitempty"`
}

// NewState returns an empty state with default settings.
func NewState() *State {
	return &State{Version: 1, Settings: DefaultSettings(), Days: map[string]*Day{}}
}

// Fix fills in what an older or hand-written state file may lack.
func (st *State) Fix() {
	if st.Version == 0 {
		st.Version = 1
	}
	if st.Days == nil {
		st.Days = map[string]*Day{}
	}
	for k, d := range st.Days {
		if d == nil {
			delete(st.Days, k)
			continue
		}
		d.Date = k
	}
	st.Settings = Normalize(st.Settings)
}
