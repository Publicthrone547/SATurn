package main

import (
	"testing"

	"github.com/Publicthrone547/SATurn/internal/core"
)

func TestParseLink(t *testing.T) {
	for raw, want := range map[string][2]string{
		"saturn://add":       {"add", ""},
		"saturn://add/":      {"add", ""},
		"saturn://add?n=5":   {"add", "5"},
		"saturn://add/10":    {"add", "10"},
		"saturn:add/3":       {"add", "3"},
		"saturn://UNDO":      {"undo", ""},
		"saturn://popover":   {"popover", ""},
		"saturn:///add?n=25": {"add", "25"},
	} {
		cmd, arg, ok := parseLink(raw)
		if !ok || cmd != want[0] || arg != want[1] {
			t.Errorf("%s: got %q %q %v, want %q %q", raw, cmd, arg, ok, want[0], want[1])
		}
	}
	for _, raw := range []string{"https://add", "saturn://", "::"} {
		if _, _, ok := parseLink(raw); ok {
			t.Errorf("%s: accepted", raw)
		}
	}
}

func TestQuestionTask(t *testing.T) {
	day := func(tasks ...core.TaskView) core.Snapshot {
		var s core.Snapshot
		s.Day.Tasks = tasks
		return s
	}
	test := core.TaskView{Kind: "test", IsTest: true, Target: 1}
	math := core.TaskView{Kind: "math", Target: 100}
	if got := questionTask(day(math)); got != -1 {
		t.Errorf("math day: %d", got)
	}
	if got := questionTask(day(test, math)); got != 1 {
		t.Errorf("test + math, test active: %d", got)
	}
	if got := questionTask(day(test)); got != -2 {
		t.Errorf("test only: %d", got)
	}
}
