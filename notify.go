package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/services/notifications"

	"github.com/Publicthrone547/reminder_sat/internal/core"
)

// setupNotifications checks the permission (installed, not dry) and
// registers the categories.
func (s *Shell) setupNotifications() {
	if s.env.dry {
		log.Printf("notifications dry")
		return
	}
	if s.env.installed {
		ok, err := s.notes.CheckNotificationAuthorization()
		log.Printf("notifications authorized=%v err=%v", ok, err)
		s.setSys(func(v *core.SystemView) { v.Notifications = ok })
	}
	s.syncCategories(s.Snapshot())
}

// syncCategories registers the categories again when the quick-add titles change.
func (s *Shell) syncCategories(snap core.Snapshot) {
	if s.env.dry || s.notes == nil {
		return
	}
	q := activeQuick(snap)
	key := fmt.Sprint(q)
	s.viewMu.Lock()
	same := key == s.catKey
	s.catKey = key
	s.viewMu.Unlock()
	if same {
		return
	}
	reply := func(c notifications.NotificationCategory) notifications.NotificationCategory {
		c.HasReplyField, c.ReplyPlaceholder, c.ReplyButtonTitle = true, "How many solved?", "Log"
		return c
	}
	label := func(i int) string { return "+" + strconv.Itoa(q[i]) }
	cats := []notifications.NotificationCategory{
		reply(notifications.NotificationCategory{ID: "remind-q", Actions: []notifications.NotificationAction{
			{ID: "add:2", Title: label(2)}, {ID: "add:3", Title: label(3)}, {ID: "studying", Title: "I'm studying"},
		}}),
		{ID: "remind-test", Actions: []notifications.NotificationAction{
			{ID: "done", Title: "Test done"}, {ID: "studying", Title: "Taking it now"},
		}},
		reply(notifications.NotificationCategory{ID: "ask", Actions: []notifications.NotificationAction{
			{ID: "add:1", Title: label(1)}, {ID: "add:2", Title: label(2)}, {ID: "add:3", Title: label(3)},
		}}),
	}
	go func() {
		for _, c := range cats {
			if err := s.notes.RegisterNotificationCategory(c); err != nil {
				log.Printf("notifications category=%s err=%q", c.ID, err)
			}
		}
	}()
}

// sendReminder posts a reminder (or the "how many?" question) with buttons,
// replacing the previous one.
func (s *Shell) sendReminder(a core.Action, snap core.Snapshot) {
	cat := "remind-q"
	if activeIsTest(snap) {
		cat = "remind-test"
	} else if a.Kind == "ask" {
		cat = "ask"
	}
	id := fmt.Sprintf("remind-%s-%s", snap.Today, strings.ReplaceAll(snap.Clock, ":", ""))
	if _, test := a.Data["test"]; test {
		id += fmt.Sprintf("-test%d", time.Now().UnixNano()%1000)
	}
	opts := notifications.NotificationOptions{
		ID: id, Title: a.Title, Subtitle: a.Sub, Body: a.Body, CategoryID: cat, ThreadID: "remind",
	}
	if a.Level >= 3 {
		opts.InterruptionLevel = notifications.InterruptionLevelTimeSensitive
	}
	if a.Level >= 2 {
		opts.Sound = &notifications.NotificationSound{Silent: true}
	}
	if s.env.dry {
		log.Printf("notification dry id=%s category=%s level=%d", id, cat, a.Level)
		return
	}
	go func() {
		s.clearReminders()
		err := s.notes.SendNotificationWithActions(opts)
		log.Printf("notification sent id=%s category=%s err=%v", id, cat, err)
		if err == nil {
			s.notifMu.Lock()
			s.lastNotif = append(s.lastNotif, id)
			s.notifMu.Unlock()
		}
	}()
}

// sendPlain posts a notification without buttons (closing message, exam eve).
func (s *Shell) sendPlain(a core.Action) {
	id := fmt.Sprintf("%s-%s", a.Situation, a.Data["date"])
	if s.env.dry {
		log.Printf("notification dry id=%s", id)
		return
	}
	go func() {
		err := s.notes.SendNotification(notifications.NotificationOptions{ID: id, Title: a.Title, Subtitle: a.Sub, Body: a.Body, ThreadID: "remind"})
		log.Printf("notification sent id=%s err=%v", id, err)
	}()
}

// clearReminders removes delivered reminders, so they do not pile up.
func (s *Shell) clearReminders() {
	if s.env.dry || s.notes == nil {
		return
	}
	s.notifMu.Lock()
	ids := s.lastNotif
	s.lastNotif = nil
	s.notifMu.Unlock()
	for _, id := range ids {
		_ = s.notes.RemoveDeliveredNotification(id)
	}
}

// onNotificationResponse handles a button, a typed reply or a click.
func (s *Shell) onNotificationResponse(r notifications.NotificationResult) {
	if r.Error != nil {
		log.Printf("notification response err=%q", r.Error)
		return
	}
	id := r.Response.ActionIdentifier
	log.Printf("notification response action=%s", id)
	snap := s.Snapshot()
	switch {
	case strings.HasPrefix(id, "add:"):
		i, err := strconv.Atoi(strings.TrimPrefix(id, "add:"))
		if err != nil || i < 0 || i > 3 || activeIsTest(snap) {
			return
		}
		s.add(-1, activeQuick(snap)[i], "notification")
	case id == "studying":
		_, _ = s.do("session source=notification", func(st *core.State, now time.Time) ([]core.Action, error) {
			return core.StartSession(st, now, 0), nil
		})
	case id == "done":
		_, _ = s.do("done source=notification", func(st *core.State, now time.Time) ([]core.Action, error) {
			return core.MarkDone(st, now, -1, "notification"), nil
		})
	case id == "TEXT_REPLY":
		text := strings.TrimPrefix(strings.TrimSpace(r.Response.UserText), "+")
		n, err := strconv.Atoi(text)
		if err != nil || n < 1 || n > 500 || activeIsTest(snap) {
			log.Printf("notification reply ignored text=%q", r.Response.UserText)
			return
		}
		s.add(-1, n, "notification")
	case id == notifications.DefaultActionIdentifier:
		s.showPopover("")
	}
}

// requestNotifications asks macOS for permission (blocks until answered).
func (s *Shell) requestNotifications() bool {
	if s.env.quiet() {
		log.Printf("notifications request skipped (not installed or dry)")
		return false
	}
	ok, err := s.notes.RequestNotificationAuthorization()
	log.Printf("notifications request authorized=%v err=%v", ok, err)
	s.setSys(func(v *core.SystemView) { v.Notifications = ok })
	s.push(s.Snapshot())
	return ok
}
