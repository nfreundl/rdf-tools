/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */
package parser

import (
	"fmt"
	"strings"

	"github.com/nfreundl/rdf-tools/model"
)

type Parser struct {

	// source and target
	source <-chan *Token
	target chan<- *model.Statement

	// states
	baseUri       model.IRI
	namespaces    map[model.Prefix]model.IRI
	bnodeLabels   map[string]*model.LabelledBlankNode
	curSubject    model.RDFTerm
	curPredicate  model.RDFTerm
	curObject     model.RDFTerm
	curReifier    model.RDFTerm
	curTripleTerm model.TripleTerm
	curGraph      model.RDFTerm

	// stack for recursive blank node property list
	bnodeStack *BnodeStack

	// the PName factory
	newPrefixedName model.PNameConstructor
}

func NewParser(source <-chan *Token, target chan<- *model.Statement) *Parser {
	ret := &Parser{
		source:      source,
		target:      target,
		namespaces:  make(map[model.Prefix]model.IRI),
		bnodeLabels: make(map[string]*model.LabelledBlankNode),
		// all the rest is nil !
		bnodeStack: NewStack(),
	}

	ret.newPrefixedName = model.PrefixNameFactory(ret.namespaces)

	return ret

}

func (this *Parser) Start() {
	// with bufferless channels, this need to be started otherwise it waits
	go this.run()
}

func (this *Parser) run() {
	defer close(this.target)

	for {
		val, ok := <-this.source
		if !ok {
			return
		}

		if val.tokenType == Dot {
			if (this.curSubject != nil) && (this.curPredicate != nil) && (this.curObject != nil) {
				this.target <- &model.Statement{
					Subject:   this.curSubject,
					Object:    this.curObject,
					Predicate: this.curPredicate,
					Context:   this.curGraph,
				}
				this.curSubject = nil
				this.curPredicate = nil
				this.curObject = nil

			} else if (this.curSubject != nil) && (this.curObject == nil) {
				panic("unexpected .")
			} // else do nothing because it was already nil nil nil because of, e. g. a previous blank node

		} else if val.tokenType == SemiColumn {
			this.target <- &model.Statement{
				Subject:   this.curSubject,
				Object:    this.curObject,
				Predicate: this.curPredicate,
				Context:   this.curGraph,
			}

			this.curPredicate = nil
			this.curObject = nil
		} else if val.tokenType == Coma {
			this.target <- &model.Statement{
				Subject:   this.curSubject,
				Object:    this.curObject,
				Predicate: this.curPredicate,
				Context:   this.curGraph,
			}

			this.curObject = nil

		} else if val.tokenType == BlankNodeAnonymous {
			newBlankNode := model.NewAnonymousBlankNode()
			if (this.curSubject != nil) && (this.curPredicate != nil) {
				this.curObject = newBlankNode
				this.target <- &model.Statement{
					Subject:   this.curSubject,
					Object:    this.curObject,
					Predicate: this.curPredicate,
					Context:   this.curGraph,
				}
				this.curSubject = nil
				this.curPredicate = nil
				this.curObject = nil
			} else if this.curSubject == nil {
				this.curSubject = newBlankNode
			} else {
				panic("unexcpected blank node predicate")
			}
		} else if val.tokenType == BlankNodeOpening {
			newBlankNode := model.NewAnonymousBlankNode()
			if (this.curSubject != nil) && (this.curPredicate != nil) {
				this.curObject = newBlankNode
				this.target <- &model.Statement{
					Subject:   this.curSubject,
					Object:    this.curObject,
					Predicate: this.curPredicate,
					Context:   this.curGraph,
				}
				this.curSubject, this.curPredicate, this.curObject = nil, nil, nil

			} else if (this.curSubject == nil) && (this.curPredicate == nil) {
				this.curSubject = newBlankNode

			} else {
				panic("blank node as predicate not implemented")
			}
			// entering the blank node property list: save the current state by pushing to the stack, then initialize the current subject
			this.bnodeStack.Add(this.curSubject, this.curPredicate, this.curObject)
			this.curSubject = newBlankNode

			// this.runInsideBlankNode(newBlankNode)

		} else if val.tokenType == BlankNodeClosing {
			if (this.curSubject != nil) && (this.curPredicate != nil) && (this.curObject != nil) {
				this.target <- &model.Statement{
					Subject:   this.curSubject,
					Object:    this.curObject,
					Predicate: this.curPredicate,
					Context:   this.curGraph,
				}
			} else if (this.curSubject != nil) && (this.curObject == nil) {
				panic("closing blank node property list to early !")
			} // if all are nil, it means a Dot was pu

			this.curSubject, this.curPredicate, this.curObject = this.bnodeStack.Pop()

		} else if val.tokenType == PNameNS {
			// TODO investigate whether it should be an error outside prefixes declaration ?

			if this.curSubject == nil {
				this.curSubject = this.newPrefixedName(val.value, "")
			} else if this.curPredicate == nil {
				this.curPredicate = this.newPrefixedName(val.value, "")
			} else {
				this.curObject = this.newPrefixedName(val.value, "")
			}
			//|| (val.tokenType == IRI) || (val.tokenType == A)

		} else if val.tokenType == PNameLN {
			// TODO, it is a bit undoing what was done
			splt := strings.SplitN(val.value, ":", 2)
			prefix := splt[0] + ":"
			pnLocal := splt[1]
			if this.curSubject == nil {
				this.curSubject = this.newPrefixedName(prefix, pnLocal)
			} else if this.curPredicate == nil {
				this.curPredicate = this.newPrefixedName(prefix, pnLocal)
			} else {
				this.curObject = this.newPrefixedName(prefix, pnLocal)
			}
		} else if val.tokenType == A {
			if this.curSubject == nil {
				this.curSubject = model.A
			} else if this.curPredicate == nil {
				this.curPredicate = model.A
			} else {
				this.curObject = model.A
			}
		} else if val.tokenType == BaseTag {
			val = <-this.source
			if val.tokenType == PNameNS {
				val = <-this.source
				if val.tokenType == IRI {
					base := val.value
					val = <-this.source
					if val.tokenType == Dot {
						this.baseUri = model.IRI(base)
					} else {
						panic("error")
					}
				} else {
					panic("error")
				}
			} else {
				panic("error")
			}
		} else if val.tokenType == PrefixTag {
			val = <-this.source
			if val.tokenType == PNameNS {
				prefix := val.value
				val = <-this.source
				if val.tokenType == IRI {
					iri := val.value
					val = <-this.source
					if val.tokenType == Dot {
						this.namespaces[model.Prefix(prefix)] = model.IRI(iri)
					} else {
						panic(fmt.Sprintf("error, expected Dot, got %v", val.tokenType))
					}
				} else {
					panic("error")
				}
			} else {
				panic("error")
			}
		} else if val.tokenType == CollectionOpening {

			if (this.curSubject == nil) && (this.curPredicate == nil) {
				this.curSubject = model.NewAnonymousBlankNode()
				this.runInsidePropertyList(this.curSubject.(*model.AnonymousBlankNode))
			} else if (this.curSubject != nil) && (this.curPredicate != nil) && (this.curObject == nil) {
				this.curObject = model.NewAnonymousBlankNode()
				this.runInsidePropertyList(this.curObject.(*model.AnonymousBlankNode))
			} else {
				panic("unexpected (")
			}
		} else if val.tokenType == CollectionClosing {
			panic("unexpected )")
		} else if val.tokenType == EmptyCollection {
			// by convention the empty collection reduces to rdfs:nil TODO source this comment
			if (this.curSubject == nil) && (this.curPredicate == nil) {
				this.curSubject = model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#nil")
			} else if (this.curSubject != nil) && (this.curPredicate != nil) && (this.curObject == nil) {
				this.curObject = model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#nil")
			} else {
				panic("unexpected ()")
			}
		}

	}
}

func (this *Parser) runInsideBlankNode(newBlankNode *model.AnonymousBlankNode) {
	val, ok := <-this.source
	if !ok {
		panic("unexpected final opening blank node")
	}

	var curPredicate model.RDFTerm
	var curObject model.RDFTerm
	for {
		if (val.tokenType == PNameLN) || (val.tokenType == PNameNS) || (val.tokenType == IRI) || (val.tokenType == A) {
			if curPredicate == nil {
				curPredicate = model.A

			} else if curObject == nil {

				curObject = model.A

			} else {
				panic("expecting , ; or ]")
			}

		} else if val.tokenType == Coma {
			if (curObject == nil) || (curPredicate == nil) {
				panic("unexpected ,")
			}
			this.target <- &model.Statement{
				Subject:   newBlankNode,
				Predicate: curPredicate,
				Object:    curObject,
				Context:   this.curGraph,
			}
			curObject = nil

		} else if val.tokenType == SemiColumn {
			if (curObject == nil) || (curPredicate == nil) {
				panic("unexpected ;")
			}
			this.target <- &model.Statement{
				Subject:   newBlankNode,
				Predicate: curPredicate,
				Object:    curObject,
				Context:   this.curGraph,
			}
			curPredicate = nil
			curObject = nil
		} else if val.tokenType == BlankNodeClosing {
			return
		} else {
			panic("not implemented")
		}
		val = <-this.source
	}
}

func (this *Parser) runInsidePropertyList(newBlankNode *model.AnonymousBlankNode) {

	curElm := model.BlankNode(newBlankNode)
	val := <-this.source
	for {
		if val.tokenType == CollectionClosing {
			this.target <- &model.Statement{
				Subject:   curElm,
				Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"),
				Object:    model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"),
			}
		} else if (val.tokenType == PNameLN) || (val.tokenType == PNameNS) || (val.tokenType == IRI) || (val.tokenType == A) {
			newElm := model.NewAnonymousBlankNode()
			this.target <- &model.Statement{
				Subject:   curElm,
				Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#first"),
				Object:    model.A, //val
			}
			this.target <- &model.Statement{
				Subject:   curElm,
				Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"),
				Object:    newElm,
			}
			curElm = newElm

		} else {
			panic("not implemented")
		}

	}
}

type RuneSet map[rune]struct{}

func newSet(input ...rune) RuneSet {
	ret := make(RuneSet)

	for _, v := range input {
		ret[v] = struct{}{}

	}

	return ret

}

func (this RuneSet) add(input ...rune) RuneSet {

	for _, v := range input {
		this[v] = struct{}{}

	}
	return this

}
func (this RuneSet) contains(testRune rune) bool {

	_, ok := this[testRune]
	return ok

}

func (this RuneSet) addRange(lower rune, inclusiveUpper rune) RuneSet {
	for i := lower; i <= inclusiveUpper; i++ {
		this[i] = struct{}{}
	}
	return this
}

func (this RuneSet) remove(input ...rune) {
	for _, v := range input {
		delete(this, v)
	}
}

func (this RuneSet) removeRange(lower rune, inclusiveUpper rune) {
	{
		for i := lower; i <= inclusiveUpper; i++ {
			delete(this, i)
		}
	}
}

func (this RuneSet) copy() RuneSet {
	ret := newSet()

	for elm, _ := range this {
		ret.add(elm)
	}
	return ret
}
