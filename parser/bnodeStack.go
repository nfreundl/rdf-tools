/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */
package parser

import "github.com/nfreundl/rdf-tools/model"

// instead of doing recursive function runInsideBnode, let's use a stack and add the stack to the Parser States
type bnodeStackElement struct {
	node    [3]model.RDFTerm
	nextElm *bnodeStackElement
}

type BnodeStack struct {
	*bnodeStackElement
}

func NewStack() *BnodeStack {
	return &BnodeStack{}

}

func (this *BnodeStack) Add(subject, predicate, object model.RDFTerm) {
	tmp := this.bnodeStackElement
	this.bnodeStackElement = &bnodeStackElement{}
	this.node = [3]model.RDFTerm{subject, predicate, object}
	this.nextElm = tmp
}

func (this *BnodeStack) Pop() (subject, predicate, object model.RDFTerm) {
	ret := this.node
	if this.nextElm != nil {
		this.bnodeStackElement = this.nextElm
	}
	return ret[0], ret[1], ret[2]
}
