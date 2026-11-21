// Package trayicon renders the menu bar ring as a 44×44 template PNG
// (black with alpha; macOS tints it for the menu bar).
package trayicon

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"sync"
)

const size = 44

var (
	mu    sync.Mutex
	cache = map[string][]byte{}
)

// Key identifies a rendering; equal keys give equal images.
func Key(glyph string, progress float64, session bool) string {
	return fmt.Sprintf("%s/%d/%v", glyph, int(math.Round(math.Max(0, math.Min(1, progress))*20)), session)
}

func cov(edge float64) float64 { return math.Max(0, math.Min(1, 0.5-edge)) }

// Render draws the glyph: ring (progress 0..1, optional session dot), done,
// off (a crescent) or exam (a triangle).
func Render(glyph string, progress float64, session bool) []byte {
	key := Key(glyph, progress, session)
	mu.Lock()
	defer mu.Unlock()
	if b, ok := cache[key]; ok {
		return b
	}
	p := float64(int(math.Round(math.Max(0, math.Min(1, progress))*20))) / 20
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			a := alpha(glyph, p, session, float64(x), float64(y))
			if a > 0 {
				img.SetNRGBA(x, y, color.NRGBA{A: uint8(math.Round(a * 255))})
			}
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	cache[key] = buf.Bytes()
	return cache[key]
}

func alpha(glyph string, progress float64, session bool, x, y float64) float64 {
	dx, dy := x+0.5-22, y+0.5-22
	d := math.Hypot(dx, dy)
	const R, w = 13.0, 3.4
	ringCov := cov(math.Abs(d-R) - w/2)
	switch glyph {
	case "done":
		return math.Max(ringCov, cov(d-6.0))
	case "off":
		return cov(d-12) * (1 - cov(math.Hypot(dx-5, dy+3)-10.5))
	case "exam":
		// 3×3 supersampled triangle (22,9) (34,33) (10,33).
		n := 0
		for sy := 0; sy < 3; sy++ {
			for sx := 0; sx < 3; sx++ {
				if inTriangle(x+(float64(sx)+0.5)/3, y+(float64(sy)+0.5)/3) {
					n++
				}
			}
		}
		return float64(n) / 9
	}
	angle := math.Atan2(dx, -dy)
	if angle < 0 {
		angle += 2 * math.Pi
	}
	a := ringCov * 0.30
	if angle <= progress*2*math.Pi && progress > 0 {
		a = ringCov
	}
	if session {
		a = math.Max(a, cov(d-3.0))
	}
	return a
}

func inTriangle(px, py float64) bool {
	sign := func(x1, y1, x2, y2, x3, y3 float64) float64 { return (x1-x3)*(y2-y3) - (x2-x3)*(y1-y3) }
	d1 := sign(px, py, 22, 9, 34, 33)
	d2 := sign(px, py, 34, 33, 10, 33)
	d3 := sign(px, py, 10, 33, 22, 9)
	neg := d1 < 0 || d2 < 0 || d3 < 0
	pos := d1 > 0 || d2 > 0 || d3 > 0
	return !(neg && pos)
}
