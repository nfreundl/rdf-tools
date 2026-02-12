/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */

package model

import (
	"testing"
)

func TestBlankNodesEqualities(t *testing.T) {
	a := NewAnonymousBlankNode()
	b := NewAnonymousBlankNode()
	c := NewBlankNodeFromLabel("c")
	d := NewBlankNodeFromLabel("d")
	e := NewBlankNodeFromLabel(a.getID())
	f := NewBlankNodeFromLabel("d")

	if !d.Equals(f) {
		t.Error("bnodes d and f should be equal")
	}
	if !a.Equals(e) {
		t.Error("bnodes a and e should be equal")
	}

	if a.Equals(b) {
		t.Error("bnodes a and b should be different")
	}

	if c.Equals(d) {
		t.Error("bnodes c and d should not be equal")
	}

}
