// Command gensounds writes SATurn's five sounds as 16-bit mono WAV files.
//
//	go run ./tools/gensounds internal/sound/wav
package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

const rate = 44100

type note struct{ freq, start float64 }

type sound struct {
	name   string
	notes  []note
	tau    float64
	length float64
}

var sounds = []sound{
	{"tick", []note{{1200, 0}}, 0.012, 0.06},
	{"quarter", []note{{783.99, 0}, {1046.50, 0.09}}, 0.18, 0.6},
	{"done", []note{{523.25, 0}, {659.25, 0.09}, {783.99, 0.18}, {1046.50, 0.27}}, 0.35, 1.4},
	{"remind", []note{{329.63, 0}, {261.63, 0.16}}, 0.30, 0.9},
	{"alarm", []note{{440, 0}, {440, 0.22}, {440, 0.44}}, 0.12, 0.9},
}

func render(s sound) []int16 {
	n := int(s.length * rate)
	buf := make([]float64, n)
	for _, nt := range s.notes {
		for i := int(nt.start * rate); i < n; i++ {
			t := float64(i)/rate - nt.start
			tone := math.Sin(2*math.Pi*nt.freq*t) + 0.3*math.Sin(2*math.Pi*2*nt.freq*t)
			env := (1 - math.Exp(-t/0.004)) * math.Exp(-t/s.tau)
			buf[i] += 0.35 * tone * env
		}
	}
	fade := int(0.010 * rate)
	for i := 0; i < fade && i < n; i++ {
		buf[n-1-i] *= float64(i) / float64(fade)
	}
	peak := 0.0
	for _, v := range buf {
		peak = math.Max(peak, math.Abs(v))
	}
	out := make([]int16, n)
	for i, v := range buf {
		if peak > 0 {
			v = v / peak * 0.8
		}
		out[i] = int16(math.Round(v * 32767))
	}
	return out
}

func writeWAV(path string, samples []int16) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	data := len(samples) * 2
	w := func(v any) { _ = binary.Write(f, binary.LittleEndian, v) }
	f.WriteString("RIFF")
	w(uint32(36 + data))
	f.WriteString("WAVEfmt ")
	w(uint32(16))
	w(uint16(1)) // PCM
	w(uint16(1)) // mono
	w(uint32(rate))
	w(uint32(rate * 2))
	w(uint16(2))
	w(uint16(16))
	f.WriteString("data")
	w(uint32(data))
	w(samples)
	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gensounds <dir>")
		os.Exit(2)
	}
	dir := os.Args[1]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, s := range sounds {
		p := filepath.Join(dir, s.name+".wav")
		if err := writeWAV(p, render(s)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("wrote", p)
	}
}
