package drawing

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// So, why don't we use true type?
// It is that we do not want to deal with vectors.
// Servers may lack the necessary GPU resources.
// You can write a vector upgrade to reduce latency, if your servers have GPU.
// However, it will likely be more expensive for sporadic workloads.

func LoadFont(indexes string, name string, logs string) {
	f := NoErrorFile(os.Open(name))
	defer func() { _ = f.Close() }()
	full := NoErrorImage(png.Decode(f))
	glyph := image.NewRGBA64(full.Bounds())
	for j := glyph.Bounds().Min.Y; j < glyph.Bounds().Max.Y; j++ {
		for i := glyph.Bounds().Min.X; i < glyph.Bounds().Max.X; i++ {
			r, _, _, _ := full.At(i, j).RGBA()
			var c color.RGBA64
			if r < 128 {
				c = color.RGBA64{0, 0, 0, 0}
			} else {
				c = color.RGBA64{0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF}
			}
			glyph.Set(i, j, c)
		}
	}

	empty0 := true
	top := make([]int, 0)
	bottom := make([]int, 0)
	row := 0
	for j := glyph.Bounds().Min.Y; j < glyph.Bounds().Max.Y; j++ {
		empty1 := true
		for i := glyph.Bounds().Min.X; i < glyph.Bounds().Max.X; i++ {
			_, _, _, a := glyph.At(i, j).RGBA()
			if a > 0 {
				empty1 = false
				break
			}
		}
		if empty0 && !empty1 {
			top = append(top, j)
		} else if !empty0 && empty1 {
			bottom = append(bottom, j+1)
			row++
		}
		empty0 = empty1
	}
	for row := 0; row < len(top); row++ {
		left := -1
		empty0 := true
		col := 0
		for i := glyph.Bounds().Min.X; i < glyph.Bounds().Max.X; i++ {
			empty1 := true
			for j := top[row]; j < bottom[row]; j++ {
				_, _, _, a := glyph.At(i, j).RGBA()
				if a > 0 {
					empty1 = false
					break
				}
			}
			if empty0 && !empty1 {
				left = i
			} else if !empty0 && empty1 {
				rect := image.Rectangle{Min: image.Point{X: left, Y: top[row]}, Max: image.Point{X: i + 1, Y: bottom[row]}}
				src := glyph.SubImage(rect)
				const smoothing = uint32(7)
				var seed = 5
				item := quickbmp{bits: make([]bool, src.Bounds().Dx()*src.Bounds().Dy()), bounds: src.Bounds().Sub(src.Bounds().Min)}
				for jj := 0; jj < item.bounds.Dy(); jj++ {
					for ii := 0; ii < item.bounds.Dx(); ii++ {
						for kk := uint32(0); kk < smoothing; kk++ {
							tx := src.Bounds().Min.X + ii*src.Bounds().Dx()/item.bounds.Dx() + monteCarlo(src.Bounds().Dx()/1024, &seed)
							ty := src.Bounds().Min.Y + jj*src.Bounds().Dy()/item.bounds.Dy() + monteCarlo(src.Bounds().Dy()/1024, &seed)
							_, _, _, a := src.At(tx, ty).RGBA()
							if a > 0 {
								item.bits[jj*item.bounds.Dx()+ii] = true
							}
						}
					}
				}
				r, fName := fileName(indexes, col, row)
				fontCache[r] = item
				if logs != "" {
					character := image.NewRGBA64(rect.Sub(rect.Min))
					f := path.Join(logs, fName)
					DrawImage(ImageSlice{Rgb: character, Rect: character.Bounds()}, glyph.SubImage(rect))
					_ = png.Encode(NoErrorFile(os.Create(f)), character)
				}
				col++
			}
			empty0 = empty1
		}
	}
}

func DrawGlyph(target ImageSlice, r rune, palette *[]color.RGBA64) {
	glyph := fontCache[r]
	if glyph.bits == nil || len(glyph.bits) < glyph.bounds.Dy()*glyph.bounds.Dx() {
		return
	}
	src := glyph.bounds

	var smoothing = uint32(len((*palette)) - 1)
	var seed = 5

	for j := 0; j < target.Rect.Dy(); j++ {
		for i := 0; i < target.Rect.Dx(); i++ {
			wf := uint32(0)
			tx := i * src.Dx() / target.Rect.Dx()
			dx := src.Dx() / target.Rect.Dx()
			ty := j * src.Dy() / target.Rect.Dy()
			dy := src.Dy() / target.Rect.Dy()
			for k := uint32(0); k < smoothing; k++ {
				// Do a standard Monte-Carlo method to converge to the optimal speed
				a := glyph.bits[(ty+monteCarlo(dy, &seed))*glyph.bounds.Dx()+tx+monteCarlo(dx, &seed)]
				if a {
					wf++
				}
			}
			if wf > 0 {
				(*target.Rgb).Set(target.Rect.Min.X+i, target.Rect.Min.Y+j, (*palette)[wf])
			}
		}
	}
}

func LoadSpace() {
	fontGlyph := fontCache['A'].bounds
	fontCache[' '] = quickbmp{bits: make([]bool, fontGlyph.Dx()*fontGlyph.Dy()), bounds: fontGlyph}
	fontCache['\r'] = quickbmp{bits: make([]bool, fontGlyph.Dx()*fontGlyph.Dy()), bounds: fontGlyph}
	fontCache['\n'] = quickbmp{bits: make([]bool, fontGlyph.Dx()*fontGlyph.Dy()), bounds: fontGlyph}
	fontCache['\t'] = quickbmp{bits: make([]bool, fontGlyph.Dx()*fontGlyph.Dy()), bounds: fontGlyph}
	fontCache['\v'] = quickbmp{bits: make([]bool, 1*1024), bounds: image.Rectangle{Min: image.Point{}, Max: image.Point{X: 1, Y: 2024}}}
}

func fileName(indexes string, col int, row int) (rune, string) {
	scan := bufio.NewScanner(bytes.NewBufferString(indexes))
	line := 0
	for scan.Scan() {
		current := scan.Text()
		for i, c := range current {
			if i == col && line == row {
				if c >= 'a' && c <= 'z' {
					return c, fmt.Sprintf("%s_.png", string(c))
				}
				if c == '.' {
					return c, fmt.Sprintf("dot.png")
				}
				if c == '_' {
					return c, fmt.Sprintf("underscore.png")
				}
				if c == '\'' {
					return c, fmt.Sprintf("quote.png")
				}
				if c == '"' {
					return c, fmt.Sprintf("quotes.png")
				}
				if c == ',' {
					return c, fmt.Sprintf("comma.png")
				}
				return c, fmt.Sprintf("%s.png", string(c))
			}
		}
		line++
	}
	return ' ', "default.png"
}
