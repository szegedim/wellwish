package drawing

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"path"
	"strings"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// TODO There is a small drawing PSNR at font edges to fix.

func SetupDrawing() {
	http.HandleFunc("/home.png", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(writer, request, "./drawing/res/home.png")
	})
	http.HandleFunc("/legal.png", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(writer, request, "./drawing/res/legal.png")
	})
	http.HandleFunc("/contact.png", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(writer, request, "./drawing/res/contact.png")
	})

	// Now this may actually break the server early.
	// Break fast, break hard allows users to enforce 2GB of RAM ensuring a stable run later.
	Loaded.Lock()
	fmt.Println("loading fonts")
	LoadFont(indexes, "./drawing/res/defaultfont.png", "")
	LoadFont("�", "./drawing/res/cursorwide.png", "")
	LoadSpace()
	fmt.Println("fonts loaded")
	Loaded.Unlock()
}

func DeclareForm(session *Session, pngFile string) {
	session.SelectedBox = -1
	session.BackgroundFile = pngFile

	ro := NoErrorImage(png.Decode(NoErrorFile(os.Open(session.BackgroundFile))))
	rw := image.NewRGBA64(image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: ro.Bounds().Dx(), Y: ro.Bounds().Dy()}})
	form := ImageSlice{Rgb: rw, Rect: ro.Bounds()}
	DrawImage(form, ro)

	form.Boxes = FindTextBoxes(form.Rgb)
	session.Form = form
	session.DirtyFrame = form.Rgb

	session.Text = map[int]Content{}
	session.SelectedBox = 0

	session.SignalRecalculate = func(session *Session) {
		fmt.Println(fmt.Sprintf("SignalRecalculate needs to be implemented."))
	}
	session.SignalFocusChanged = func(session *Session, from int, to int) {
		fmt.Println(fmt.Sprintf("SignalFocusChanged needs to be implemented."))
	}
	session.SignalClicked = func(session *Session, i int) {
		fmt.Println(fmt.Sprintf("SignalClicked needs to be implemented."))
	}
	session.SignalUploaded = func(session *Session, upload Upload) {
		fmt.Println(fmt.Sprintf("Uploaded %d bytes", len(upload.Body)))
		fmt.Println(fmt.Sprintf("SignalUploaded needs to be implemented."))
	}
	session.SignalClosed = func(session *Session) {
		session.SelectedBox = -1
	}
	session.SignalFullRedrawNeeded = func(session *Session) {
		RenderFullRedraw(session)
	}
	session.SignalPartialRedrawNeeded = func(session *Session, i int) {
		RenderSingleBoxChange(session, i)
	}
	session.SignalFocusLost = func(session *Session, i int) {
		TextFocusLost(session, i)
	}
	session.SignalFocusGot = func(session *Session, focused int) {
		if focused >= 0 && session.Text[focused].Editable {
			replace := session.Text[focused]
			if strings.HasPrefix(session.Text[focused].Text, RevertAndReturn) {
				replace.Text = "�"
			} else {
				if !strings.Contains(replace.Text, "�") {
					replace.Text = replace.Text + "�"
				}
			}
			session.Text[focused] = replace
			session.SignalPartialRedrawNeeded(session, focused)
		}
	}
	session.SignalTextChange = func(session *Session, i int, from string, to string) {
		session.SignalPartialRedrawNeeded(session, i)
	}
	session.SignalShiftFocus = func(session *Session) {
		session.SignalFocusLost(session, session.SelectedBox)
		for i := 0; i < len(session.Form.Boxes); i++ {
			candidate := (session.SelectedBox + 1) % len(session.Form.Boxes)
			text, ok := session.Text[candidate]
			if ok && text.Editable {
				session.SelectedBox = candidate
				if strings.HasPrefix(text.Text, RevertAndReturn) {
					session.SignalFocusGot(session, session.SelectedBox)
				} else {
					session.SignalPartialRedrawNeeded(session, session.SelectedBox)
				}
				break
			}
		}
	}
}

func PutText(session *Session, i int, t Content) int {
	if i == -1 {
		i = len(session.Text)
	}
	session.Text[i] = t
	if session.SelectedBox == -1 {
		session.SelectedBox = i
		session.SignalFocusGot(session, i)
	}
	return i
}

func SetImage(session *Session, i int, pngFile string, t Content) int {
	if i == -1 {
		i = len(session.Text)
	}
	t.BackgroundFile = pngFile
	ro := NoErrorImage(png.Decode(NoErrorFile(os.Open(t.BackgroundFile))))
	rw := image.NewRGBA64(session.Form.Boxes[i])
	t.Background = ImageSlice{Rgb: rw, Rect: rw.Bounds()}
	DrawImage(t.Background, ro)
	session.Text[i] = t
	return i
}

func ChangeFocus(session *Session, switchTo int) {
	if session.SelectedBox != switchTo {
		if session.Text[switchTo].Selectable || session.Text[switchTo].Editable {
			session.SignalFocusLost(session, session.SelectedBox)
			session.SignalFocusGot(session, switchTo)
			switchFrom := session.SelectedBox
			session.SelectedBox = switchTo
			if switchFrom != switchTo {
				session.SignalFocusChanged(session, switchFrom, switchTo)
			}
			session.SignalPartialRedrawNeeded(session, switchTo)
			return
		}
	}
	session.SignalClicked(session, switchTo)
}

func FillWithColor(target ImageSlice, color color.Color) {
	_, _, _, a := color.RGBA()
	if a > 0 {
		bounds := target.Rect
		for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
			for i := bounds.Min.X; i < bounds.Max.X; i++ {
				(*target.Rgb).Set(i, j, color)
			}
		}
	}
}

func EraseBorder(target ImageSlice) {
	bounds := target.Rect
	for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
		for i := bounds.Min.X; i < bounds.Max.X; i++ {
			if j == bounds.Min.Y || j == bounds.Max.Y-1 ||
				i == bounds.Min.X || i == bounds.Max.X-1 {
				original := (*target.Rgb).At(i-1, j-1)
				(*target.Rgb).Set(i, j, original)
			} else {
				i = bounds.Max.X - 2
			}
		}
	}
}

func EraseBox(session *Session, target ImageSlice) {
	bounds := target.Rect
	ro := NoErrorImage(png.Decode(NoErrorFile(os.Open(session.BackgroundFile))))
	for j := bounds.Min.Y; j < bounds.Max.Y; j++ {
		for i := bounds.Min.X; i < bounds.Max.X; i++ {
			original := ro.At(i, j)
			(*target.Rgb).Set(i, j, original)
		}
	}
}

func SampleImage(target ImageSlice, rectangle image.Rectangle) ImageSlice {
	ret := image.NewRGBA64(rectangle)
	src := target.Rgb.SubImage(rectangle)
	for j := rectangle.Min.Y; j < rectangle.Max.Y; j++ {
		for i := rectangle.Min.X; i < rectangle.Max.X; i++ {
			c := src.At(i, j)
			ret.Set(i, j, c)
		}
	}
	return ImageSlice{Rgb: ret, Rect: rectangle}
}

func DrawImage(target ImageSlice, img image.Image) {
	src := img.Bounds()

	const smoothing = uint32(7)
	var seed = 5
	for j := 0; j < target.Rect.Dy(); j++ {
		for i := 0; i < target.Rect.Dx(); i++ {
			final := color.NRGBA64{}
			for k := uint32(0); k < smoothing; k++ {
				// Do a standard Monte-Carlo method to converge to the optimal speed
				tx := src.Min.X + i*src.Dx()/target.Rect.Dx() + monteCarlo(src.Dx()/target.Rect.Dx(), &seed)
				ty := src.Min.Y + j*src.Dy()/target.Rect.Dy() + monteCarlo(src.Dy()/target.Rect.Dy(), &seed)
				r, g, b, a := img.At(tx, ty).RGBA()
				if a == 0xFFFF {
					final.R = uint16((uint32(final.R)*k + uint32(r)) / (k + 1))
					final.G = uint16((uint32(final.G)*k + uint32(g)) / (k + 1))
					final.B = uint16((uint32(final.B)*k + uint32(b)) / (k + 1))
					final.A = uint16((uint32(final.A)*k + uint32(a)) / (k + 1))
				}
			}
			if final.A > 0 && final.A < 0xFFFF {
				r, g, b, _ := target.Rect.At(i, j).RGBA()
				final.R = uint16((r + uint32(final.R)) / 2)
				final.G = uint16((g + uint32(final.G)) / 2)
				final.B = uint16((b + uint32(final.B)) / 2)
				final.A = 0xFFFF
				(*target.Rgb).Set(target.Rect.Min.X+i, target.Rect.Min.Y+j, final)
			}
			if final.A == 0xFFFF {
				(*target.Rgb).Set(target.Rect.Min.X+i, target.Rect.Min.Y+j, final)
			}
		}
	}
}

func RGBMatch(c0 color.Color, c1 color.Color) bool {
	r0, g0, b0, a0 := c0.RGBA()
	r1, g1, b1, a1 := c1.RGBA()
	return r0 == r1 && g0 == g1 && b0 == b1 && a0 == a1
}

func FindTextBoxes(img image.Image) []image.Rectangle {
	// Active rectangles are marked with a border of this special color
	var colorKey = color.RGBA{R: 0xE9, G: 0x44, B: 0x20, A: 0xFF}
	ret := make([]image.Rectangle, 0)

	for j := img.Bounds().Min.Y; j < img.Bounds().Max.Y-2; j++ {
		for i := img.Bounds().Min.X; i < img.Bounds().Max.X-2; i++ {
			c := img.At(i+1, j)
			if !RGBMatch(c, colorKey) {
				continue
			}
			if img.At(i, j) == c &&
				img.At(i+1, j) == c &&
				img.At(i+2, j) == c &&
				img.At(i, j+1) == c &&
				img.At(i, j+2) == c &&
				img.At(i+1, j+1) != c {
				x := img.At(i, j)
				rect := image.Rectangle{Min: image.Point{X: i, Y: j}, Max: image.Point{X: i + 2, Y: j + 2}}
				for rect.Max.X = rect.Min.X + 2; rect.Max.X < img.Bounds().Max.X; rect.Max.X++ {
					if img.At(rect.Max.X, rect.Min.Y) != x {
						break
					}
				}
				for rect.Max.Y = rect.Min.Y + 2; rect.Max.Y < img.Bounds().Max.Y; rect.Max.Y++ {
					if img.At(rect.Min.X, rect.Max.Y) != x {
						break
					}
				}
				found := false
				for _, r := range ret {
					if rect.In(r) {
						found = true
						break
					}
				}
				if !found {
					ret = append(ret, rect)
				}
			}
		}
	}
	return ret
}

func FindGridTextBoxes(target ImageSlice, grid *Grid) {
	width := grid.Width
	height := grid.Height
	thickness := grid.Thickness

	tw := (len(width) + 1) * thickness
	for _, c := range width {
		tw = tw + c
	}
	th := (len(height) + 1) * thickness
	for _, c := range height {
		th = th + c
	}

	temp := make([]image.Rectangle, 0)
	isum := thickness
	jsum := thickness
	for _, j := range height {
		for _, i := range width {
			rect := image.Rectangle{Min: image.Point{X: isum, Y: jsum}, Max: image.Point{X: isum + i, Y: jsum + j}}
			isum = isum + i + thickness
			temp = append(temp, rect)
		}
		isum = thickness
		jsum = jsum + j + thickness
	}

	nominator := target.Rect.Dx()
	denominator := tw
	if 1000*target.Rect.Dy()/th < 1000*nominator/denominator {
		nominator = target.Rect.Dy()
		denominator = th
	}
	scaledThickness := thickness * nominator / denominator

	grid.Boxes = make([]image.Rectangle, 0)
	grid.Borders = make([]image.Rectangle, 0)
	for i := range temp {
		box := image.Rectangle{Min: temp[i].Min.Mul(nominator).Div(denominator).Add(target.Rect.Min), Max: temp[i].Max.Mul(nominator).Div(denominator).Add(target.Rect.Min)}
		grid.Boxes = append(grid.Boxes, box)
		border := image.Rectangle{Min: box.Min.Add(image.Point{X: -scaledThickness, Y: -scaledThickness}), Max: box.Max.Add(image.Point{X: scaledThickness, Y: scaledThickness})}
		grid.Borders = append(grid.Borders, border)
	}

	shift := image.Point{}
	if grid.Alignment == 0 {
		shift = target.Rect.Max.Sub(grid.Borders[len(grid.Borders)-1].Max).Div(2)
	} else if grid.Alignment == 1 {
		shift = target.Rect.Max.Sub(grid.Borders[len(grid.Borders)-1].Max)
	}
	for i := range grid.Borders {
		grid.Boxes[i] = grid.Boxes[i].Add(shift)
		grid.Borders[i] = grid.Borders[i].Add(shift)
	}
}

type Content struct {
	Text           string
	BackgroundFile string
	Background     ImageSlice

	Editable   bool
	Selectable bool
	Lines      int

	Alignment       int
	FontColor       color.Color
	BackgroundColor color.Color
}

func DrawTextWithFillWithErase(session *Session, target ImageSlice, format Content) {
	EraseBox(session, target)
	EraseBorder(target)
	DrawTextWithFill(target, format)
}

func DrawTextWithFill(target ImageSlice, format Content) {
	if format.Background.Rgb != nil {
		DrawTextOnImage(target, format)
		return
	}
	//FillWithColor(target, format.BackgroundColor)
	DrawText(target, format)
}

func DrawTextOnImage(target ImageSlice, format Content) {
	DrawImage(target, format.Background.Rgb.SubImage(format.Background.Rect))
	DrawText(target, format)
}

type Grid struct {
	Width     []int
	Height    []int
	Thickness int
	Alignment int
	Boxes     []image.Rectangle
	Borders   []image.Rectangle
}

func DrawTextGridWithFill(target ImageSlice, grid *Grid, pos image.Point, format Content) {
	if grid.Boxes == nil {
		FindGridTextBoxes(target, grid)
	}
	DrawGridBorder(target, grid, pos, format)
	index := len(grid.Width)*pos.Y + pos.X
	DrawTextWithFill(ImageSlice{Rgb: target.Rgb, Rect: grid.Boxes[index]}, format)
}

func DrawGridBorder(target ImageSlice, grid *Grid, pos image.Point, format Content) {
	if grid.Boxes == nil || grid.Borders == nil {
		FindGridTextBoxes(target, grid)
	}

	index := len(grid.Width)*pos.Y + pos.X
	FillWithColor(ImageSlice{Rgb: target.Rgb, Rect: grid.Borders[index]}, format.FontColor)
	FillWithColor(ImageSlice{Rgb: target.Rgb, Rect: grid.Boxes[index]}, format.BackgroundColor)
}

func DrawGridPadding(target ImageSlice, grid *Grid, format Content) {
	if grid.Boxes == nil || grid.Borders == nil {
		FindGridTextBoxes(target, grid)
	}

	FillWithColor(ImageSlice{Rgb: target.Rgb, Rect: target.Rect}, format.BackgroundColor)
}

func DrawText(target ImageSlice, format Content) {
	text := format.Text
	if text == "" {
		return
	}
	if format.Lines == 0 {
		return
	}
	for iteration := 0; iteration < 2; iteration++ {
		type character struct {
			rect image.Rectangle
			r    rune
			pos  image.Point
		}
		chars := make([]character, 0)
		cursor := image.Point{}
		bounds := image.Rectangle{}

		lines := 1
		nominalLineMultiplier := 12
		nominalLineDivider := 10
		for _, c := range text {
			glyph := fontCache[c]

			if glyph.bits == nil {
				continue
			}
			nominalLineHeight := 1024 * nominalLineMultiplier / nominalLineDivider
			nominalWidthShift := 1024 * glyph.bounds.Dx() / glyph.bounds.Dy()
			scaledGlyph := image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{nominalWidthShift, nominalLineHeight}}
			bounds = bounds.Union(scaledGlyph.Add(cursor))
			if c == '\r' {
				cursor = cursor.Sub(image.Point{X: cursor.X, Y: 0})
			} else if c == '\n' {
				cursor = cursor.Add(image.Point{X: -cursor.X, Y: nominalLineHeight})
				lines++
			} else {
				chars = append(chars, character{rect: scaledGlyph, r: c, pos: cursor})
				cursor = cursor.Add(image.Point{X: nominalWidthShift * 140 / 100, Y: 0})
			}
		}
		if lines > format.Lines {
			text = "ERROR Too Many New Lines"
			continue
		}

		if format.Alignment == -1 {
			reverse := make([]character, len(chars))
			padding := bounds.Max.X - chars[len(chars)-1].pos.X - chars[len(chars)-1].rect.Dx()
			for i := 0; i < len(chars); i++ {
				j := len(chars) - i - 1
				if i > 0 && chars[j+1].pos.Y > chars[j].pos.Y {
					padding = bounds.Max.X - chars[j].pos.X - chars[j].rect.Dx()
				}
				reverse[j] = character{rect: chars[j].rect, r: chars[j].r, pos: image.Point{X: chars[j].pos.X + padding, Y: chars[j].pos.Y}}
			}
			chars = reverse
		} else if format.Alignment == 0 {
			center := make([]character, len(chars))
			padding := bounds.Max.X - chars[len(chars)-1].pos.X - chars[len(chars)-1].rect.Dx()
			for i := 0; i < len(chars); i++ {
				j := len(chars) - i - 1
				if i > 0 && chars[j+1].pos.Y > chars[j].pos.Y {
					padding = bounds.Max.X - chars[j].pos.X - chars[j].rect.Dx()
				}
				center[j] = character{rect: chars[j].rect, r: chars[j].r, pos: image.Point{X: chars[j].pos.X + padding/2, Y: chars[j].pos.Y}}
			}
			chars = center
		}

		scaleNominator := lines
		scaleDenominator := format.Lines
		scaleNominator = scaleNominator * target.Rect.Dy()
		scaleDenominator = scaleDenominator * bounds.Dy()
		if scaleDenominator == 0 {
			continue
		}
		if bounds.Dx()*scaleNominator/scaleDenominator > target.Rect.Dx() {
			// Always fit regardless of shape.
			scaleNominator = target.Rect.Dx()
			scaleDenominator = bounds.Dx()
		}
		// TODO 0 align box

		r, g, b, a := format.FontColor.RGBA()
		br, bg, bb, ba := format.BackgroundColor.RGBA()
		const smoothing = uint32(7)
		palette := make([]color.RGBA64, smoothing+1)
		for wf := uint32(0); wf <= smoothing; wf++ {
			wb := smoothing - wf
			palette[wf] = color.RGBA64{
				R: uint16((r*wf + br*wb) / smoothing),
				G: uint16((g*wf + bg*wb) / smoothing),
				B: uint16((b*wf + bb*wb) / smoothing),
				A: uint16((a*wf + ba*wb) / smoothing)}
		}

		ready := make(chan int)
		for _, c := range chars {
			fitPos := c.pos.Mul(scaleNominator).Div(scaleDenominator)
			fitSize := image.Rectangle{Min: c.rect.Min.Mul(scaleNominator).Div(scaleDenominator), Max: c.rect.Max.Mul(scaleNominator).Div(scaleDenominator)}
			fitBounds := fitSize.Add(fitPos).Add(target.Rect.Min)
			if format.Alignment == 0 && fitBounds.Dy() < target.Rect.Dy() {
				fitBounds = fitBounds.Add(image.Point{X: 0, Y: (target.Rect.Dy() - fitBounds.Dy()) / 2})
			}
			go func(ready chan int, slice ImageSlice, r rune, palette *[]color.RGBA64) {
				DrawGlyph(slice, r, palette)
				ready <- 1
			}(ready, ImageSlice{target.Rgb, fitBounds, nil}, c.r, &palette)
		}
		for i := 0; i < len(chars); i++ {
			<-ready
		}
	}
}

func ChangedRectangle(a image.Image, b image.Image) image.Rectangle {
	ret := image.Rectangle{}
	for j := a.Bounds().Min.Y; j < a.Bounds().Max.Y; j++ {
		for i := a.Bounds().Min.X; i < a.Bounds().Max.X; i++ {
			if a.At(i, j) != b.At(i, j) {
				x := image.Rectangle{Min: image.Point{X: i, Y: j}, Max: image.Point{X: i + 1, Y: j + 1}}
				ret = ret.Union(x)
			}
		}
	}
	return ret
}

func TextFocusLost(session *Session, i int) {
	if i != -1 && session.Text[i].Editable {
		before, after, ok := strings.Cut(session.Text[i].Text, "�")
		if ok {
			replace := session.Text[i]
			replace.Text = before + after
			session.Text[i] = replace
			session.SignalPartialRedrawNeeded(session, i)
		}
	}
}

func CombineImages(pngFile string, overlay string) string {
	loaded := NewImageSliceFromPng(pngFile)
	checked := NewImageSliceFromPng(overlay)
	DrawImage(loaded, checked.Rgb)
	f := path.Join("/tmp", path.Base(pngFile))
	_ = png.Encode(NoErrorFile(os.Create(f)), loaded.Rgb)
	return f
}
