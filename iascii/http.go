//2,|gofmt
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iascii

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"text/template"
)

import (
	"appengine"
	"appengine/urlfetch"
	"resize"
)

var (
	errorTemplate  = template.Must(template.ParseFiles("error.html"))
	uploadTemplate = template.Must(template.ParseFiles("upload.html"))
)

func init() {
	http.HandleFunc("/", errorHandler(upload))
}

// handler for '/'; if the request is a POST try to convert the image
func upload(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.Method != "POST" {
		// No upload; show the upload form.
		if err := uploadTemplate.Execute(w, ""); err != nil {
			c.Errorf("Can't execute uploadTempl: ", err)
		}
		return
	}

	// Grab the image data
	var buf bytes.Buffer
	f, _, err := r.FormFile("image")
	if err != nil {
		u := r.FormValue("url")
		// c.Infof("about to fetch %v\n", u)
		if len(u) == 0 {
			if err = uploadTemplate.Execute(w, "");  err != nil {
				c.Errorf("Can't execute uploadTempalte:", err)
			}
			return
		}
		client := urlfetch.Client(c)
		resp, err := client.Get(u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		buf.Reset()
		io.Copy(&buf, resp.Body)
	} else {
		defer f.Close()
		io.Copy(&buf, f)
	}

	// c.Infof("length of buffer is: %d\n", buf.Len())
	i, _, err := image.Decode(&buf)
	check(err)

	const max = 300
	if b := i.Bounds(); b.Dx() > max || b.Dy() > max {
		// If it's gigantic, it's more efficient to downsample first
		// and then resize; resizing will smooth out the roughness.
		if b.Dx() > 2*max || b.Dy() > 2*max {
			w, h := max, max
			if b.Dx() > b.Dy() {
				h = b.Dy() * h / b.Dx()
			} else {
				w = b.Dx() * w / b.Dy()
			}
			i = resize.Resample(i, i.Bounds(), w, h)
			b = i.Bounds()
		}
		w, h := max/2, max/2
		if b.Dx() > b.Dy() {
			h = b.Dy() * h / b.Dx()
		} else {
			w = b.Dx() * w / b.Dy()
		}
		i = resize.Resize(i, i.Bounds(), w, h)
	}

	// Encode as a new ascii image
	buf.Reset()
	err = Encode(&buf, i)
	check(err)
	OutBuf := string(buf.Bytes())
	// c.Infof("Done converting; buf:\n%v\n", OutBuf)

	w.Header().Set("Content-type", "text/html")
	uploadTemplate.Execute(w, OutBuf)
}

// errorHandler wraps the argument handler with an error-catcher that
// returns a 500 HTTP error if the request fails (calls check with err non-nil).
func errorHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err, ok := recover().(error); ok {
				w.WriteHeader(http.StatusInternalServerError)
				errorTemplate.Execute(w, err)
			}
		}()
		fn(w, r)
	}
}

// check aborts the current execution if err is non-nil.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
