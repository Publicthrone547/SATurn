package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// env holds the development switches (SATURN_* environment variables).
type env struct {
	dataDir   string
	dry       bool
	installed bool
	tick      time.Duration
	idle      int // SATURN_IDLE: fixed idle seconds for test runs; -1 reads the real value
	open      string
	script    string
	base      time.Time // SATURN_NOW
	speed     float64   // SATURN_SPEED
	started   time.Time
}

func readEnv() env {
	e := env{
		dry:     os.Getenv("SATURN_DRY") == "1",
		tick:    20 * time.Second,
		open:    os.Getenv("SATURN_OPEN"),
		script:  os.Getenv("SATURN_SCRIPT"),
		speed:   1,
		started: time.Now(),
		idle:    -1,
	}
	if v := os.Getenv("SATURN_IDLE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			e.idle = n
		}
	}
	if exe, err := os.Executable(); err == nil {
		if r, err := filepath.EvalSymlinks(exe); err == nil {
			exe = r
		}
		e.installed = strings.HasPrefix(exe, "/Applications/")
	}
	if v := os.Getenv("SATURN_TICK"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			e.tick = d
		}
	}
	if v := os.Getenv("SATURN_NOW"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02T15:04:05", v, time.Local); err == nil {
			e.base = t
		} else {
			log.Printf("env bad SATURN_NOW=%q", v)
		}
	}
	if v := os.Getenv("SATURN_SPEED"); v != "" && !e.base.IsZero() {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			e.speed = f
		}
	}
	return e
}

// now is the app clock: real time, or SATURN_NOW running at SATURN_SPEED.
func (e env) now() time.Time {
	if e.base.IsZero() {
		return time.Now()
	}
	elapsed := time.Since(e.started)
	return e.base.Add(time.Duration(float64(elapsed) * e.speed))
}

// quiet reports whether permission requests and login registration must be
// skipped: dry runs and builds outside /Applications.
func (e env) quiet() bool { return e.dry || !e.installed }

// setupLog writes the log to standard error and to a file: the data folder
// when SATURN_DATA_DIR is set, else ~/Library/Logs/SATurn/saturn.log.
func setupLog() string {
	var path string
	if d := os.Getenv("SATURN_DATA_DIR"); d != "" {
		path = filepath.Join(d, "saturn.log")
	} else if home, err := os.UserHomeDir(); err == nil {
		path = filepath.Join(home, "Library", "Logs", "SATurn", "saturn.log")
	}
	log.SetFlags(log.Ldate | log.Ltime)
	if path == "" {
		return ""
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return ""
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return ""
	}
	log.SetOutput(io.MultiWriter(os.Stderr, f))
	return path
}
