// Command simulate replays one study day against the reminder engine with a
// fake clock and prints every action as a timeline. No window opens.
//
//	go run ./cmd/simulate -date 2026-11-24 -preset math \
//	    -events "10:10 add 25, 14:00 session 50, 16:30 add 25"
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Publicthrone547/reminder_sat/internal/core"
)

type event struct {
	at   int
	verb string
	arg  string
}

func parseClock(s string) (int, error) {
	t, err := time.Parse("15:04", strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("bad time %q", s)
	}
	return t.Hour()*60 + t.Minute(), nil
}

func parseEvents(s string) ([]event, error) {
	var out []event
	for _, part := range strings.Split(s, ",") {
		f := strings.Fields(part)
		if len(f) == 0 {
			continue
		}
		if len(f) < 2 {
			return nil, fmt.Errorf("bad event %q", part)
		}
		at, err := parseClock(f[0])
		if err != nil {
			return nil, err
		}
		e := event{at: at, verb: f[1]}
		if len(f) > 2 {
			e.arg = f[2]
		}
		out = append(out, e)
	}
	return out, nil
}

func line(at time.Time, a core.Action) string {
	head := a.Kind
	if a.Kind == "remind" {
		head = fmt.Sprintf("remind L%d", a.Level)
	}
	return fmt.Sprintf("%s  %-10s %-12s %s | %s", at.Format("15:04"), head, a.Situation, a.Title, a.Body)
}

func main() {
	date := flag.String("date", "2026-11-24", "the logical date to simulate")
	exam := flag.String("exam", core.DefaultExamDate, "the exam date")
	preset := flag.String("preset", "math", "the day's preset")
	from := flag.String("from", "08:00", "start time")
	to := flag.String("to", "23:30", "end time")
	evs := flag.String("events", "", `scripted input, e.g. "10:10 add 25, 14:00 session 50"`)
	snapshot := flag.Bool("snapshot", false, "print the snapshot as JSON at the end instead of the timeline")
	flag.Parse()

	events, err := parseEvents(*evs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	start, err1 := parseClock(*from)
	end, err2 := parseClock(*to)
	if err1 != nil || err2 != nil {
		fmt.Fprintln(os.Stderr, "bad -from or -to")
		os.Exit(2)
	}

	st := core.NewState()
	st.Settings.ExamDate = *exam
	st.Settings.Onboarded = true
	st.FirstDay = *date
	at := func(m int) time.Time { return core.AtMinute(*date, m) }
	if _, err := core.SetDay(st, at(start), *date, *preset, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	quiet := *snapshot
	print := func(now time.Time, acts []core.Action) {
		if quiet {
			return
		}
		for _, a := range acts {
			fmt.Println(line(now, a))
		}
	}
	idle, sleepUntil, returned := 0, -1, false
	for m := start; m <= end; m++ {
		now := at(m)
		for _, e := range events {
			if e.at != m {
				continue
			}
			n, _ := strconv.Atoi(e.arg)
			switch e.verb {
			case "add":
				print(now, core.Add(st, now, -1, n, "sim"))
			case "session":
				print(now, core.StartSession(st, now, n))
			case "end":
				print(now, core.EndSession(st, now))
			case "promise":
				if _, err := core.MakePromise(st, now, n); err != nil && !quiet {
					fmt.Printf("%s  promise refused: %v\n", now.Format("15:04"), err)
				}
			case "off":
				if e.arg == "none" {
					e.arg = ""
				}
				_, _ = core.SetOff(st, now, "", e.arg)
			case "idle":
				idle = n
			case "sleep":
				if s, err := parseClock(e.arg); err == nil {
					sleepUntil = s
				}
			}
		}
		if sleepUntil >= 0 {
			if m < sleepUntil {
				continue
			}
			sleepUntil, returned = -1, true
		}
		print(now, core.Tick(st, core.TickInput{Now: now, IdleSeconds: idle, Returned: returned}))
		returned = false
	}
	if *snapshot {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(core.View(st, at(end)))
	}
}
