/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Edouard Martin Freundler
 */
package parser

import (
	"errors"
	"fmt"
	"io"

	"unicode/utf8"
)

func NewByteSource(reader io.Reader, buffSize int, channel chan<- byte) {

	buffer := make([]byte, buffSize)
	defer close(channel)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return
			}
			// TODO handle this
			panic(err)

		}
		if n == 0 {
			return
		}
		for i := 0; i < n; i++ {
			channel <- buffer[i]
		}

	}
}

func NewRuneUtf8Source(source <-chan byte, target chan<- rune) {
	defer close(target)
	tmp := make([]byte, 0, 4)

	for b := range source {
		tmp = processBytes(b, tmp, target)
	}
	if len(tmp) != 0 {
		panic(errors.New(fmt.Sprintf("Invalid UTF8 %v", tmp)))

	}

}

func processBytes(r byte, tmp []byte, target chan<- rune) []byte {
	tmp = append(tmp, r)
	retRune, _ := utf8.DecodeRune(tmp)
	if retRune == utf8.RuneError {
		if len(tmp) == 4 {
			panic(errors.New(fmt.Sprintf("Invalid UTF8 %v", tmp)))
		} else {
			return tmp
		}
	} else {

		target <- retRune
		// I want to clear without calling GC
		// that makes a copy and changes the copy
		// tmp = tmp[:0]
		return tmp[:0]
	}

}

func NewRuneReader(inputFile io.Reader, readerBufferSize int, byteChannelSize int, runeChannelSize int) <-chan rune {
	// TODO new error, buffers must be > 1

	byteChan := make(chan byte, byteChannelSize)
	runeChan := make(chan rune, runeChannelSize)

	go NewByteSource(inputFile, readerBufferSize, byteChan)
	go NewRuneUtf8Source(byteChan, runeChan)

	return runeChan
}
