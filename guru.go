// Copyright (c) 2014 David R. Jenni. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Guru is an experimental wrapper around oracle for use with editor buffers inside of Acme.

It executes oracle with the given arguments on the current cursor selection of the Acme buffer.

Usage:
	Guru <mode> <args> ...
For information about <mode> and <args>, read the oracle help (oracle -help).

Guru assumes that the following change is applied, which adds a mechanism for editor buffers:
https://go-review.googlesource.com/#/c/10400/
*/
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"9fans.net/go/acme"
)

type file struct {
	name       string
	body       []byte
	off0, off1 int // byte offsets of selection
}

func current() (*file, error) {
	win, err := openWin()
	if err != nil {
		return nil, err
	}
	defer win.CloseFiles()

	body, err := ioutil.ReadAll(&bodyReader{win})
	if err != nil {
		return nil, err
	}

	name, err := filename(win)
	if err != nil {
		return nil, err
	}

	off0, off1, err := selection(win, body)
	if err != nil {
		return nil, err
	}

	return &file{name, body, off0, off1}, nil
}

func openWin() (*acme.Win, error) {
	id, err := strconv.Atoi(os.Getenv("winid"))
	if err != nil {
		return nil, err
	}
	return acme.Open(id, nil)
}

func filename(win *acme.Win) (string, error) {
	b, err := win.ReadAll("tag")
	if err != nil {
		return "", err
	}
	tag := string(b)
	i := strings.Index(tag, " ")
	if i == -1 {
		return "", errors.New("cannot get filename from tag")
	}
	return tag[0:i], nil
}

func selection(win *acme.Win, body []byte) (int, int, error) {
	if _, _, err := win.ReadAddr(); err != nil {
		return 0, 0, err
	}
	if err := win.Ctl("addr=dot"); err != nil {
		return 0, 0, err
	}
	q0, q1, err := win.ReadAddr()
	if err != nil {
		return 0, 0, err
	}
	off0, err := byteOff(q0, bytes.NewReader(body))
	if err != nil {
		return 0, 0, err
	}
	off1, err := byteOff(q1, bytes.NewReader(body))
	if err != nil {
		return 0, 0, err
	}
	return off0, off1, nil
}

func byteOff(q int, r io.RuneReader) (off int, err error) {
	for i := 0; i != q; i++ {
		_, s, err := r.ReadRune()
		if err != nil {
			return 0, err
		}
		off += s
	}
	return
}

type bodyReader struct{ *acme.Win }

func (r bodyReader) Read(d []byte) (int, error) {
	return r.Win.Read("body", d)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: Guru <mode> <args> ...\n")
		os.Exit(1)
	}

	f, err := current()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open window: %v\n", err)
		os.Exit(1)
	}

	var args []string
	args = append(args, "-pos", fmt.Sprintf("%s:#%d,#%d", f.name, f.off0, f.off1))
	args = append(args, "-replaceset", f.name+",-")
	args = append(args, os.Args[1:]...)

	c := exec.Command("oracle", args...)
	c.Stderr = os.Stderr
	c.Stdout = os.Stderr
	c.Stdin = bytes.NewReader(f.body)

	if err = c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "oracle error: %v\n", err)
		os.Exit(1)
	}
}
