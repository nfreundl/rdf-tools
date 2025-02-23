/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Edouard Martin Freundler
 */
package parser

import (
	"bytes"
	"testing"
	"unicode/utf8"
)

func TestUtf8Source(t *testing.T) {
	testString := "I will be a valid stringééööΓ"
	x := []byte(testString)
	ioReader := bytes.NewReader(x)
	byteChan := make(chan byte, 1)

	go NewByteSource(ioReader, 1, byteChan)

	y := []byte{}
	for b := range byteChan {
		y = append(y, b)

	}

	if string(y) != testString {
		t.Errorf("result string %s is not equal to test string %s", string(y), testString)
	}

}

func TestHowUtf8Works(t *testing.T) {
	buff := make([]byte, 4)
	utf8.EncodeRune(buff, 'Γ')
	retRune, size := utf8.DecodeRune(buff)
	if retRune != 'Γ' {
		t.Errorf("Γ not returned, %c instead", retRune)
	}
	if size != 2 {
		t.Errorf("read %d instead of %d bytes", size, 2)
	}

}

func TestByteToUtf8(t *testing.T) {
	testString := "ΓééI will be a valid stringééööΓ"
	testBytes := []byte(testString)
	if testBytes[0] != 0xce {
		t.Errorf("%v is not equal to exected %v", testBytes[0], 0xce)
	}
	byteChan := make(chan byte, 4)
	runeChan := make(chan rune, 1)

	go func() {

		for _, b := range testBytes {
			byteChan <- b

		}

		close(byteChan)

	}()

	go NewRuneUtf8Source(byteChan, runeChan)

	res := []rune{}

	for r := range runeChan {
		res = append(res, r)

	}

	if testString != string(res) {
		t.Errorf("result string %s is not equal to test string %s", string(res), testString)
	}

}
