/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */
package isocanonical

import (
	"encoding/binary"

	"github.com/nfreundl/rdf-tools/model"
	"github.com/twmb/murmur3"
)

const distinguisher = byte('@')

func hashTerm(term model.RDFTerm) [2]uint64 {
	h1, h2 := murmur3.StringSum128(term.String())
	return [2]uint64{h1, h2}
}

func hashPreviousHash(g [2]uint64) [2]uint64 {
	buffer := make([]byte, 16)
	binary.LittleEndian.PutUint64(buffer, g[0])
	binary.LittleEndian.PutUint64(buffer[8:], g[1])
	h1, h2 := murmur3.Sum128(buffer)
	return [2]uint64{h1, h2}
}

// order matters
func hash3Tuple(g1, g2, g3 [2]uint64) [2]uint64 {
	buffer := make([]byte, 48)
	binary.LittleEndian.PutUint64(buffer, g1[0])
	binary.LittleEndian.PutUint64(buffer[8:], g1[1])
	binary.LittleEndian.PutUint64(buffer[16:], g2[0])
	binary.LittleEndian.PutUint64(buffer[24:], g2[1])
	binary.LittleEndian.PutUint64(buffer[32:], g3[0])
	binary.LittleEndian.PutUint64(buffer[38:], g3[1])
	h1, h2 := murmur3.Sum128(buffer)
	return [2]uint64{h1, h2}
}

func hashWithDistinguisher(g [2]uint64) [2]uint64 {
	buffer := make([]byte, 17)
	binary.LittleEndian.PutUint64(buffer, g[0])
	binary.LittleEndian.PutUint64(buffer[8:], g[1])
	buffer[16] = distinguisher
	h1, h2 := murmur3.Sum128(buffer)
	return [2]uint64{h1, h2}
}

// associative an commutative
func hashBag(h1, h2 [2]uint64) [2]uint64 {
	return [2]uint64{h1[0] ^ h2[0], h1[1] ^ h2[1]}
}

func greater(h1, h2 [2]uint64) bool {
	if h1[0] > h2[0] {
		return true
	}
	return h1[1] > h2[1]
}
