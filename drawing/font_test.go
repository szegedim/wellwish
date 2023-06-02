package drawing

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

func TestConvertAll(t *testing.T) {
	// Loading a font should not take more than 20 seconds at startup
	start := time.Now()
	LoadFont(indexes, "res/defaultfont.png", "/tmp/")
	LoadSpace()
	LoadFont("�", "res/cursorwide.png", "/tmp/")
	if time.Now().Sub(start).Seconds() > 20 {
		t.Error("timeout 20s")
	}
}

// Counter-prussian scaling algorithm
func TestPhotonicOverlay(t *testing.T) {
	var seed = 5
	in, _ := png.Decode(NoErrorReader(os.Open("./res/tig.png")))
	weight := in.Bounds().Dx() * in.Bounds().Dy()
	mc := make([]int, weight)
	for i := 0; i < weight; i++ {
		mc[i] = monteCarlo2(weight, &seed)
	}
	standard := image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{1000, 1000}}
	out := image.NewRGBA(standard)
	a := time.Now()
	var done int64
	if 2 == 1 {
		// Full quality measurement
		for i := 0; i < weight; i++ {
			x := i % in.Bounds().Dy()
			y := i / in.Bounds().Dy()
			sx := x + in.Bounds().Min.X
			sy := y + in.Bounds().Min.Y
			o := out
			tx := x*o.Bounds().Dx()/in.Bounds().Dx() + in.Bounds().Min.X
			ty := y*o.Bounds().Dy()/in.Bounds().Dy() + in.Bounds().Min.Y
			c := in.At(sx+2, sy)
			out.Set(tx, ty, c)
			done = done + 4
		}
	} else {
		hz := 20
		periodus := int64(1000 * 1000 / hz)
		core := func(o *image.RGBA, donec *chan int64) {
			i := 0
			var d0 int64
			for {
				x := mc[i%weight] % in.Bounds().Dx()
				y := mc[(i+1)%weight] % in.Bounds().Dy()
				sx := x + in.Bounds().Min.X
				sy := y + in.Bounds().Min.Y
				tx := x*o.Bounds().Dx()/in.Bounds().Dx() + in.Bounds().Min.X
				ty := y*o.Bounds().Dy()/in.Bounds().Dy() + in.Bounds().Min.Y
				c := in.At(sx, sy)

				Set1(tx, ty, c, out)
				d0 = d0 + 4

				if i%1000 == 1 {
					if time.Now().Sub(a).Microseconds() > periodus {
						break
					}
				}
				i++
			}
			(*donec) <- d0
		}
		var ch = make(chan int64)
		for i := 0; i < 16; i++ {
			go core(out, &ch)
		}
		for i := 0; i < 16; i++ {
			d1 := <-ch
			done = done + d1
		}
	}
	msec := time.Now().Sub(a).Milliseconds()
	durationMs := int64(msec)
	fmt.Println(int64(done)*8/durationMs/1024, "MBps")

	accu := Black
	for i := 0; i < weight; i++ {
		x := i % in.Bounds().Dy()
		y := i / in.Bounds().Dy()
		c := out.At(x, y)
		if alpha(c.RGBA()) == 0 && alpha(accu.RGBA()) != 0 {
			c = accu
			out.Set(x, y, accu)
			continue
		}
		w, e, r, t := c.RGBA()
		accu = color.RGBA64{uint16(w), uint16(e), uint16(r), uint16(t)}
	}
	for i := 0; i < weight; i++ {
		x := i % in.Bounds().Dx()
		y := i / in.Bounds().Dx()
		c := out.At(x, y)
		if alpha(c.RGBA()) == 0 && alpha(accu.RGBA()) != 0 {
			c = accu
			out.Set(x, y, accu)
			continue
		}
		w, e, r, t := c.RGBA()
		accu = color.RGBA64{uint16(w), uint16(e), uint16(r), uint16(t)}
	}
	time.Sleep(1 * time.Millisecond)
	fmt.Println(msec, "ms")
	buf := bytes.NewBufferString("")
	NoErrorVoid(png.Encode(buf, out))
	NoErrorVoid(os.WriteFile("/tmp/prussian.png", buf.Bytes(), 0700))
}

func Set1(x, y int, c color.Color, p *image.RGBA) {
	r0, g0, b0, a0 := p.At(x, y).RGBA()
	r1, g1, b1, a1 := c.RGBA()
	c1 := color.RGBA64{
		R: uint16((r0*a0 + r1*a1) / (a0 + 1) / 255),
		G: uint16((g0*a0 + g1*a1) / (a0 + 1) / 256),
		B: uint16((b0*a0 + b1*a1) / (a0 + 1) / 256),
		A: uint16(a0 + 1)}
	p.Set(x, y, c1)
}

func alpha(r uint32, g uint32, b uint32, a uint32) uint32 {
	return a
}
