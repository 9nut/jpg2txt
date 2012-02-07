//2,|gofmt
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package iascii

import (
	"errors"
	"image"
	"image/color"
	"io"
)

func Encode(w io.Writer, m image.Image) (err error) {
	b := m.Bounds()
	mw, mh := b.Dx(), b.Dy()
	if mw <= 0 || mh <= 0 {
		err = errors.New("Bad image bounds")
		return
	}

	cm := []byte(".ocOGDQ@")
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := color.GrayModel.Convert(m.At(x, y)).(color.Gray)
			c.Y >>= 5
			_, err = w.Write(cm[c.Y : c.Y+1])
			if err != nil {
				return
			}
		}
		_, err = w.Write([]byte("\n"))
		if err != nil {
			return
		}
	}
	return
}
