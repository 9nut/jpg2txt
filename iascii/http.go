//2,|gofmt
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package iascii

import (
	"bytes"
	"http"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"old/template"
	"os"
)

import (
	"appengine"
	"appengine/urlfetch"
)

var (
	uploadTemplate *template.Template
	errorTemplate  = template.MustParseFile("error.html", nil)
)

func init() {
	http.HandleFunc("/", errorHandler(upload))
	uploadTemplate = template.New(nil)
	uploadTemplate.SetDelims("«", "»")
	if err := uploadTemplate.ParseFile("upload.html"); err != nil {
		panic("can't parse upload.html: " + err.String())
	}
}

// handler for '/'; if the request is a POST try to convert the image
func upload(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.Method != "POST" {
		// No upload; show the upload form.
		uploadTemplate.Execute(w, "")
		return
	}

	// Grab the image data
	var buf bytes.Buffer
	f, _, err := r.FormFile("image")
	if err != nil {
		u := r.FormValue("url")
		// c.Infof("about to fetch %v\n", u)
		if len(u) == 0 {
			uploadTemplate.Execute(w, "")
			return
		}
		client := urlfetch.Client(c)
		resp, err := client.Get(u)
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
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
			if err, ok := recover().(os.Error); ok {
				w.WriteHeader(http.StatusInternalServerError)
				errorTemplate.Execute(w, err)
			}
		}()
		fn(w, r)
	}
}

// check aborts the current execution if err is non-nil.
func check(err os.Error) {
	if err != nil {
		panic(err)
	}
}
