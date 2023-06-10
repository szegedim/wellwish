package drawing

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"gitlab.com/eper.io/engine/metadata"
	"image"
	"image/png"
	"io"
	random "math/rand"
	"net/http"
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
		fmt.Println(err)
		data, _ = os.Open("/dev/null")
		if data == nil {
			fmt.Println(err)
			return nil
		}
	}
	go func(f *os.File) {
		// Garbage collection
		time.Sleep(60 * time.Second)
		_ = f.Close()
	}(data)
	return data
}

func NoErrorWrite64(n int64, err error) {
	if err != nil {
		fmt.Errorf("%s\n", err)
	}
}

func NoErrorWrite(n int, err error) {
	if err != nil {
		fmt.Errorf("%s\n", err)
	}
}

func NoErrorVoid(err error) {
	if err != nil {
		fmt.Errorf("%s\n", err)
	}
}

func NoNilReader(response *http.Response) io.Reader {
	if response == nil {
		ret := bytes.Buffer{}
		return &ret
	}
	if response.Body == nil {
		ret := bytes.Buffer{}
		return &ret
	}
	return response.Body
}

func NoErrorBytes(data []byte, err error) []byte {
	if err != nil {
		fmt.Errorf("%s\n", err)
		return nil
	}
	return data
}

func NoErrorRequest(data *http.Request, err error) *http.Request {
	if err != nil {
		fmt.Errorf("%s\n", err)
		return nil
	}
	return data
}

func NoErrorResponse(data *http.Response, err error) *http.Response {
	if err != nil {
		fmt.Errorf("%s\n", err)
		return nil
	}
	return data
}

func NoErrorReader(data *os.File, err error) io.Reader {
	if err != nil {
		fmt.Errorf("%s\n", err)
		return nil
	}
	go func(f *os.File) {
		// Garbage collection
		time.Sleep(60 * time.Second)
		_ = f.Close()
	}(data)
	return data
}

func NoErrorWriter(data *os.File, err error) io.Writer {
	if err != nil {
		fmt.Errorf("%s\n", err)
		return nil
	}
	go func(f *os.File) {
		// Garbage collection
		time.Sleep(60 * time.Second)
		_ = f.Close()
	}(data)
	return data
}

func NoErrorString(data []byte, err error) string {
	if err != nil {
		fmt.Errorf("%s\n", err)
		return ""
	}
	return string(data)
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
	if uq == "" {
		return ""
	}
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
	// So we do not add much of a header suggesting it is the best solution.
	// Adding a header would increase the chance of randomly testing the
	// private key with sites to verify it works, practically leaking it.
	// Your internal context should tell where an api key is valid.

	// TODO Need to get a better seed from the internet
	x, _ := os.Stat(os.Args[0])
	seed := time.Now().UnixNano() ^ x.ModTime().UnixNano()
	random.Seed(seed)

	letters := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	key := make([]rune, 92)
	salt := metadata.RandomSalt
	for i := 0; i < 1000; i++ {
		for i := 0; i < 92; i++ {
			key[i] = letters[(((Random() ^ random.Uint32()) + uint32(salt[i])) % uint32(len(letters)))]
			time.Sleep(550 * time.Nanosecond)
		}
		if key[91] == 'A' {
			break
		}
	}
	key[91] = 'R'
	return string(key)
}
