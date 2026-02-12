/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */

package model

import "github.com/nfreundl/rdf-tools/utility"

// BlankNodes

type BlankNode interface {
	getID() string
	Equals(RDFTerm) bool
}

type LabelledBlankNode struct {
	id string
}

func (this *LabelledBlankNode) getID() string {
	return this.id
}

type AnonymousBlankNode struct {
	id string
}

func (this *AnonymousBlankNode) getID() string {
	return this.id
}

func NewBlankNodeFromLabel(label string) BlankNode {
	utility.FeedFromExternalSource(label)
	return &LabelledBlankNode{id: label}
}
func NewAnonymousBlankNode() BlankNode {
	return &AnonymousBlankNode{id: utility.GetNewID()}

}

func (this *AnonymousBlankNode) Equals(that RDFTerm) bool {
	bnode, ok := that.(BlankNode)
	if !ok {
		return false
	}
	return this.getID() == bnode.getID()
}

func (this *LabelledBlankNode) Equals(that RDFTerm) bool {
	bnode, ok := that.(BlankNode)
	if !ok {
		return false
	}
	return this.getID() == bnode.getID()
}
