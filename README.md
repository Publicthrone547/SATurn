# SATurn

SATurn is a menu bar app for macOS that keeps one person studying for one exam: the SAT on 5 December 2026. Every day is a ring of 100 math questions, 100 English questions or one practice test, closed by logging what you solved. Until the ring is closed, SATurn reminds you every hour, and each reminder says how much is left and the latest time you can still start.

SATurn is a reminder, not a study platform. It never shows questions; it only counts what you tell it you did.

## Install

Open `SATurn-1.1.0-darwin-arm64.dmg` and drag SATurn to Applications. The app is not notarised, so the first launch needs **Open Anyway** in System Settings → Privacy & Security.

From the source: `make install` builds the app, replaces `/Applications/SATurn.app` and starts it.

Requirements: macOS 14 or later, Apple silicon.

## First run

The first launch opens a window with four steps: the exam date, the hours SATurn may remind you, the plan (a suggestion counted back from the exam; click any day to change it), and the last things: allow notifications, launch at login, your target score.

In System Settings → Notifications → SATurn, choose the **Alerts** style, so reminders stay on screen until you answer them.

## Every day

- **The menu bar** shows today's count, `37/100`. Click it for the popover; right-click for a menu.
- **Logging**: the `+1 +5 +10 +25` buttons in the popover (`+1` is the big one), the right-click menu, the buttons on a reminder, a number typed into a reminder's reply field, a hotkey or a `saturn://` link from any app. Undo removes the last entry.
- **The popover**: today's count, a tally of 100 cells (one per question; each row is a quarter of the day), and a line of the reminder hours showing the latest time you can start and still finish.
- **Sessions**: "Start a 50 min session" silences reminders while you work, then asks how many you solved.
- **Later**: on a reminder, choose when you will start (15 min, 30 min, 1 hour). SATurn holds you to it.
- **Day off**: mark today (or any later day) as a rest day or a sick day. Reminders stop and the streak is kept.
- **Practice test days**: hold the button when the test is finished; the score is optional.

### Log from any app

Hotkeys, from any app (change them in Settings → Log from any app):

| Keys | Action |
| --- | --- |
| `⌥⌘1` | +1 |
| `⌥⌘5` | +5 |
| `⌥⌘0` | +10 |
| `⌥⌘Z` | undo |
| `⌥⌘S` | open or close the popover |

An entry from a hotkey or a link shows the new count under the notch for a moment. On a day with a practice test and questions, it counts towards the questions.

Links, for the Shortcuts app (action **Open URLs**), Raycast, Alfred, a Stream Deck or Terminal (`open -g saturn://add`):

| Link | Action |
| --- | --- |
| `saturn://add` | +1 |
| `saturn://add?n=5` | +5 (1 to 200) |
| `saturn://undo` | undo |
| `saturn://popover` | open or close the popover |
| `saturn://open` | the main window |

A Shortcuts shortcut with `saturn://add` can sit in the menu bar, run from Spotlight or Siri, or get its own key.

## Reminders

Between the reminder hours (09:00 to 23:00 by default) SATurn reminds you every hour until the day is done. It stays quiet during a session, right after you logged something, and while you are away from the Mac. Ignored reminders escalate: the island stays on screen, then reminders get a sound and come every 30 minutes once the day no longer fits. The day ends at 04:00, so late work still counts.

**Strict mode** (on by default) disables Quit while today is not finished. Settings → Strict mode → Quit anyway asks you to type `I am done for today`. Logging out or shutting down is never blocked.

## Settings

Exam date, target score, reminder hours, interval, the hour the day ends, daily targets, quick-add values, session length, menu bar text, sound, hotkeys (each one can be re-recorded) and links, launch at login, notifications, a test reminder, strict mode.

## Data

- State: `~/Library/Application Support/SATurn/state.json`
- Log: `~/Library/Logs/SATurn/saturn.log`

## Uninstall

Turn off Settings → Launch at login, quit SATurn, delete `/Applications/SATurn.app` and `~/Library/Application Support/SATurn`. After the exam, SATurn offers a button that does the first two steps for you.

## Build

Needs Go 1.27, Node 26 and the Wails CLI v3.0.0-beta.30 (`~/go/bin/wails3`).

```sh
make app        # bin/SATurn.app
make test       # Go and frontend tests
make simulate   # print a simulated day of reminders
make release    # checks, then bin/release-<version>/ with the dmg and SHA256SUMS
```

`npm run dev:mock` in `frontend/` serves the interface with a fake backend at `http://127.0.0.1:9246/#/popover`, `#/island` and `#/main` (flags in `frontend/src/mock.ts`, for example `?red=1` or `?island=tally`).

Development switches for the bundled binary: `SATURN_DATA_DIR` (a separate data folder), `SATURN_DRY=1` (no notifications, sounds or login changes), `SATURN_NOW` and `SATURN_SPEED` (a fast clock), `SATURN_TICK`, `SATURN_IDLE`, `SATURN_OPEN` and `SATURN_SCRIPT`. See `IMPLEMENTATION.md` section 7.11.
