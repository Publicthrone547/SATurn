package core

type Settings struct {
	Onboarded      bool              `json:"onboarded"`
	ExamDate       string            `json:"examDate"`       // "2026-12-05"
	TargetScore    int               `json:"targetScore"`    // 0 = not set
	BriefAt        int               `json:"briefAt"`        // minutes after midnight
	RemindFrom     int               `json:"remindFrom"`     // minutes after midnight
	RemindUntil    int               `json:"remindUntil"`    // minutes after midnight
	IntervalMin    int               `json:"intervalMin"`    // minutes between reminders
	RedIntervalMin int               `json:"redIntervalMin"` // minutes between reminders in the red zone
	RolloverHour   int               `json:"rolloverHour"`   // the day ends at this hour
	Targets        map[string]int    `json:"targets"`        // per kind
	QuickAdd       map[string][4]int `json:"quickAdd"`       // per question kind
	SessionMin     int               `json:"sessionMin"`
	TestSessionMin int               `json:"testSessionMin"`
	Hotkeys        bool              `json:"hotkeys"`
	Sound          bool              `json:"sound"`
	Strict         bool              `json:"strict"`
	TrayText       string            `json:"trayText"` // "progress" | "countdown" | "none"
}

// DefaultExamDate is the owner's SAT.
const DefaultExamDate = "2026-12-05"

func DefaultSettings() Settings {
	return Settings{
		ExamDate:       DefaultExamDate,
		BriefAt:        540,
		RemindFrom:     540,
		RemindUntil:    1380,
		IntervalMin:    60,
		RedIntervalMin: 30,
		RolloverHour:   4,
		Targets:        map[string]int{Math: 100, English: 100, Test: 1},
		QuickAdd:       map[string][4]int{Math: {1, 5, 10, 25}, English: {1, 5, 10, 25}},
		SessionMin:     50,
		TestSessionMin: 150,
		Hotkeys:        true,
		Sound:          true,
		Strict:         true,
		TrayText:       "progress",
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Normalize fills missing values from the defaults and clamps the rest.
// A fully zeroed Settings comes back as the defaults.
func Normalize(s Settings) Settings {
	d := DefaultSettings()
	if _, err := ParseDate(s.ExamDate); err != nil {
		s.ExamDate = d.ExamDate
	}
	if s.TargetScore != 0 {
		s.TargetScore = clamp(s.TargetScore, 400, 1600)
	}
	zero := s.RemindFrom == 0 && s.RemindUntil == 0 && s.BriefAt == 0 && s.IntervalMin == 0
	if zero {
		s.BriefAt, s.RemindFrom, s.RemindUntil = d.BriefAt, d.RemindFrom, d.RemindUntil
		s.IntervalMin, s.RedIntervalMin, s.RolloverHour = d.IntervalMin, d.RedIntervalMin, d.RolloverHour
		s.SessionMin, s.TestSessionMin = d.SessionMin, d.TestSessionMin
		s.Hotkeys, s.Sound, s.Strict = d.Hotkeys, d.Sound, d.Strict
	}
	if s.RemindFrom == 0 {
		s.RemindFrom = d.RemindFrom
	}
	s.RemindFrom = clamp(s.RemindFrom, 300, 1200)
	if s.RemindUntil == 0 {
		s.RemindUntil = d.RemindUntil
	}
	s.RemindUntil = clamp(s.RemindUntil, s.RemindFrom+120, 1439)
	if s.BriefAt == 0 {
		s.BriefAt = s.RemindFrom
	}
	s.BriefAt = clamp(s.BriefAt, 300, s.RemindFrom)
	if s.IntervalMin == 0 {
		s.IntervalMin = d.IntervalMin
	}
	s.IntervalMin = clamp(s.IntervalMin, 15, 180)
	if s.RedIntervalMin == 0 {
		s.RedIntervalMin = d.RedIntervalMin
	}
	s.RedIntervalMin = clamp(s.RedIntervalMin, 10, s.IntervalMin)
	s.RolloverHour = clamp(s.RolloverHour, 0, 5)
	if s.Targets == nil {
		s.Targets = map[string]int{}
	}
	for k, v := range d.Targets {
		if s.Targets[k] == 0 {
			s.Targets[k] = v
		}
		s.Targets[k] = clamp(s.Targets[k], 1, 500)
	}
	if s.QuickAdd == nil {
		s.QuickAdd = map[string][4]int{}
	}
	for k, v := range d.QuickAdd {
		q := s.QuickAdd[k]
		for i := range q {
			if q[i] == 0 {
				q[i] = v[i]
			}
			q[i] = clamp(q[i], 1, 200)
		}
		s.QuickAdd[k] = q
	}
	if s.SessionMin == 0 {
		s.SessionMin = d.SessionMin
	}
	s.SessionMin = clamp(s.SessionMin, 10, 180)
	if s.TestSessionMin == 0 {
		s.TestSessionMin = d.TestSessionMin
	}
	s.TestSessionMin = clamp(s.TestSessionMin, 60, 240)
	switch s.TrayText {
	case "progress", "countdown", "none":
	default:
		s.TrayText = d.TrayText
	}
	return s
}
