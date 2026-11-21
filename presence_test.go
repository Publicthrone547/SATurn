package main

import "testing"

func TestParseIdle(t *testing.T) {
	out := []byte(`+-o IOHIDSystem
    | |     "HIDIdleTime" = 449758085333
`)
	if got := parseIdle(out); got != 449 {
		t.Fatalf("%d", got)
	}
	if got := parseIdle([]byte("nothing")); got != 0 {
		t.Fatalf("%d", got)
	}
}
