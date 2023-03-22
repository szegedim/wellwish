package drawing

import (
	"crypto/rand"
	"fmt"
	"image"
	"image/png"
	math "math/rand"
	"os"
	"time"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

type ImageSlice struct {
	Rgb   *image.RGBA64
	Rect  image.Rectangle
	Boxes []image.Rectangle
}

func NewImage(_image *image.RGBA64) ImageSlice {
	return ImageSlice{Rgb: _image, Rect: _image.Bounds()}
}

func NewImageSlice(_image *image.RGBA64, rect image.Rectangle) ImageSlice {
	return ImageSlice{Rgb: _image, Rect: rect}
}

func NoErrorFile(data *os.File, err error) *os.File {
	if err != nil {
		fmt.Errorf("%s\n", err)
		return nil
	}
	return data
}

func NoErrorBytes(data []byte, err error) []byte {
	if err != nil {
		fmt.Errorf("%s\n", err)
		return nil
	}
	return data
}

func NoErrorImage(data image.Image, err error) image.Image {
	if err != nil {
		fmt.Errorf("%s\n", err)
		return nil
	}
	return data
}

func NewImageSliceFromPng(fileName string) ImageSlice {
	img := NoErrorImage(png.Decode(NoErrorFile(os.Open(fileName))))
	ret := ImageSlice{Rgb: image.NewRGBA64(img.Bounds()), Rect: img.Bounds()}
	DrawImage(ret, img)
	return ret
}

func NewImageSliceDuplicated(slice ImageSlice) ImageSlice {
	ret := ImageSlice{Rgb: image.NewRGBA64(slice.Rect), Rect: slice.Rect}
	DrawImage(ret, slice.Rgb)
	return ret
}

func RedactPublicKey(uq string) string {
	return uq[0:6]
}

func Random() uint32 {
	buf := make([]byte, 4)
	n, err := rand.Read(buf)
	if err != nil || n != 4 {
		return 0
	}
	return uint32(buf[0])<<24 | uint32(buf[0])<<16 | uint32(buf[0])<<8 | uint32(buf[0])<<0
}

func GenerateUniqueKey() string {
	// TODO Need to get something better from the internet
	x, _ := os.Stat(os.Args[0])
	seed := time.Now().UnixNano() ^ x.ModTime().UnixNano()
	math.Seed(seed)

	letters := []rune("ABCDEFGHIJKLMNOPQRST")
	key := make([]rune, 64)
	for i := 0; i < 1000; i++ {
		for i := 0; i < 64; i++ {
			key[i] = letters[(Random()^math.Uint32())%uint32(len(letters))]
			time.Sleep(550 * time.Nanosecond)
		}
		if key[3] == 'A' {
			break
		}
	}
	key[4] = 'B'
	return string(key)
}