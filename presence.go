package main

import (
	"bufio"
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

// idleSeconds is the time since the last key press or mouse move, read from
// IOHIDSystem's HIDIdleTime (nanoseconds). It needs no permission. Any error
// counts as "present".
func idleSeconds() int {
	out, err := exec.Command("/usr/sbin/ioreg", "-c", "IOHIDSystem", "-d", "4").Output()
	if err != nil {
		return 0
	}
	return parseIdle(out)
}

func parseIdle(out []byte) int {
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		i := strings.Index(line, `"HIDIdleTime" = `)
		if i < 0 {
			continue
		}
		v := strings.TrimSpace(line[i+len(`"HIDIdleTime" = `):])
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return int(n / 1e9)
	}
	return 0
}
