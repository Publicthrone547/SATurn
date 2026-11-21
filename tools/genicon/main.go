// Command genicon draws SATurn's app icon: a gold planet with an unfinished
// ring on a deep indigo rounded square.
//
//	go run ./tools/genicon build/appicon.png
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const (
	size = 1024
	ss   = 2 // supersampling per axis
)

type rgba struct{ r, g, b, a float64 }

func hex(s string, a float64) rgba {
	var r, g, b int
	fmt.Sscanf(s, "#%02x%02x%02x", &r, &g, &b)
	return rgba{float64(r) / 255, float64(g) / 255, float64(b) / 255, a}
}

// over composites c over dst (straight alpha).
func over(dst, c rgba) rgba {
	a := c.a + dst.a*(1-c.a)
	if a == 0 {
		return rgba{}
	}
	mix := func(x, y float64) float64 { return (x*c.a + y*dst.a*(1-c.a)) / a }
	return rgba{mix(c.r, dst.r), mix(c.g, dst.g), mix(c.b, dst.b), a}
}

func roundedSquare(x, y float64) bool {
	const half, r = 412.0, 185.0
	dx, dy := math.Abs(x-512)-(half-r), math.Abs(y-512)-(half-r)
	if dx <= 0 || dy <= 0 {
		return math.Abs(x-512) <= half && math.Abs(y-512) <= half
	}
	return dx*dx+dy*dy <= r*r
}

// ring returns whether (x, y) is on the ellipse stroke, which half it is on
// (front = below the planet's centre line in the rotated frame) and the
// angle along the ellipse in degrees, 0..360.
func ring(x, y float64) (on, front bool, angle float64) {
	const a, b, stroke = 330.0, 92.0, 30.0
	th := -18 * math.Pi / 180
	dx, dy := x-512, y-512
	u := dx*math.Cos(th) + dy*math.Sin(th)
	v := -dx*math.Sin(th) + dy*math.Cos(th)
	// Distance to the ellipse, approximated by scaling the normalised radius.
	k := math.Sqrt(u*u/(a*a) + v*v/(b*b))
	if k == 0 {
		return false, false, 0
	}
	gx, gy := u/(a*a), v/(b*b)
	grad := math.Sqrt(gx*gx+gy*gy) / k
	d := (k - 1) / grad
	angle = math.Mod(math.Atan2(v/b, u/a)*180/math.Pi+360, 360)
	return math.Abs(d) <= stroke/2, v > 0, angle
}

func sample(x, y float64) rgba {
	px := rgba{}
	if !roundedSquare(x, y) {
		return px
	}
	t := (y - 100) / 824
	top, bottom := hex("#171B4A", 1), hex("#0D1030", 1)
	px = rgba{top.r + (bottom.r-top.r)*t, top.g + (bottom.g-top.g)*t, top.b + (bottom.b-top.b)*t, 1}

	gold := hex("#F2C66D", 1)
	onRing, front, ang := ring(x, y)
	// The upper right quarter of the ring is unfinished.
	ringAlpha := 0.95
	if ang >= 270 {
		ringAlpha = 0.30
	}
	if onRing && !front {
		px = over(px, rgba{gold.r, gold.g, gold.b, ringAlpha})
	}
	dx, dy := x-512, y-512
	if dx*dx+dy*dy <= 170*170 {
		c := gold
		if band := math.Mod(y-512+400, 80); band >= 40 && band < 62 {
			c = hex("#D9B28A", 1)
		}
		// A little light from the top left.
		shade := 1 - 0.18*((dx+dy)/340+0.5)
		px = over(px, rgba{c.r * shade, c.g * shade, c.b * shade, 1})
	}
	if onRing && front {
		px = over(px, rgba{gold.r, gold.g, gold.b, ringAlpha})
	}
	return px
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: genicon <out.png>")
		os.Exit(2)
	}
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			var r, g, b, a float64
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					c := sample(float64(px)+(float64(sx)+0.5)/ss, float64(py)+(float64(sy)+0.5)/ss)
					r += c.r * c.a
					g += c.g * c.a
					b += c.b * c.a
					a += c.a
				}
			}
			n := float64(ss * ss)
			if a == 0 {
				continue
			}
			img.SetNRGBA(px, py, color.NRGBA{
				R: uint8(math.Round(r / a * 255)), G: uint8(math.Round(g / a * 255)),
				B: uint8(math.Round(b / a * 255)), A: uint8(math.Round(a / n * 255)),
			})
		}
	}
	f, err := os.Create(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("wrote", os.Args[1])
}
