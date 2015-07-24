// Copyright (c) 2014 David R. Jenni. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"testing"
)

type offsetTest struct {
	data       []byte
	offset     int
	byteOffset int
}

var offsetTests = []offsetTest{
	{[]byte("abcdef"), 0, 0},
	{[]byte("abcdef"), 1, 1},
	{[]byte("abcdef"), 5, 5},
	{[]byte("日本語def"), 0, 0},
	{[]byte("日本語def"), 1, 3},
	{[]byte("日本語def"), 5, 11},
}

func TestByteOffset(t *testing.T) {
	for _, test := range offsetTests {
		off, err := byteOff(test.offset, bytes.NewReader(test.data))
		if err != nil {
			t.Errorf("got error %v", err)
		}
		if off != test.byteOffset {
			t.Errorf("expected byte offset %d, got %d", test.byteOffset, off)
		}
	}
}
