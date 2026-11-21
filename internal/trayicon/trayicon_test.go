package trayicon

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRender(t *testing.T) {
	for _, g := range []string{"ring", "done", "off", "exam"} {
		b := Render(g, 0.37, g == "ring")
		img, err := png.Decode(bytes.NewReader(b))
		if err != nil {
			t.Fatalf("%s: %v", g, err)
		}
		if r := img.Bounds(); r.Dx() != 44 || r.Dy() != 44 {
			t.Fatalf("%s: %v", g, r)
		}
		var ink int
		for y := 0; y < 44; y++ {
			for x := 0; x < 44; x++ {
				if _, _, _, a := img.At(x, y).RGBA(); a > 0 {
					ink++
				}
			}
		}
		if ink < 50 {
			t.Fatalf("%s: almost empty (%d pixels)", g, ink)
		}
	}
	if Key("ring", 0.36, false) != Key("ring", 0.37, false) {
		t.Fatal("nearby progress should share a key")
	}
	if !bytes.Equal(Render("ring", 0.5, false), Render("ring", 0.5, false)) {
		t.Fatal("cache")
	}
}
