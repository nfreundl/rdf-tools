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
			// unfortunately, this will be deferred to when a , ; . is met, because reifier can change that
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

			this.runInsideBlankNode(newBlankNode.(*model.AnonymousBlankNode))
			continue
			// this.bnodeStack.Add(this.curSubject, this.curPredicate, this.curObject)
			//this.curSubject = newBlankNode

			// this.runInsideBlankNode(newBlankNode)

		} else if val.tokenType == BlankNodeClosing {

			panic("closing blank node property list to early !")

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
				this.runInsideCollection(this.curSubject.(*model.AnonymousBlankNode))
				continue
			} else if (this.curSubject != nil) && (this.curPredicate != nil) && (this.curObject == nil) {
				this.curObject = model.NewAnonymousBlankNode()
				this.runInsideCollection(this.curObject.(*model.AnonymousBlankNode))
				continue
			} else {
				panic("unexpected collection for the predicate")
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

	var curPredicate model.RDFTerm
	var curObject model.RDFTerm
	for {
		val, ok := <-this.source
		if !ok {
			panic("unexpected EOF in the blank node list")
		}

		if (val.tokenType == PNameLN) || (val.tokenType == IRI) || (val.tokenType == A) {
			var node model.RDFTerm
			switch val.tokenType {
			case IRI:
				node = model.IRI(val.value)
			case A:
				node = model.A
			case PNameLN:
				splt := strings.SplitN(val.value, ":", 2)
				prefix := splt[0] + ":"
				pnLocal := splt[1]
				node = this.newPrefixedName(prefix, pnLocal)
			}

			if curPredicate == nil {
				curPredicate = node

			} else if curObject == nil {

				curObject = node

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
			if (curPredicate != nil) && (curObject != nil) {
				this.target <- &model.Statement{
					Subject:   newBlankNode,
					Object:    curObject,
					Predicate: curPredicate,
					Context:   this.curGraph,
				}
				return
			} else {
				panic("unexpected ]")
			}

		} else if val.tokenType == BlankNodeAnonymous {
			newBlankNode2 := model.NewAnonymousBlankNode()
			if curPredicate != nil {
				curObject = newBlankNode2
				this.target <- &model.Statement{
					Subject:   newBlankNode,
					Object:    curObject,
					Predicate: curPredicate,
					Context:   this.curGraph,
				}
				curPredicate, curObject = nil, nil
			} else {
				panic("unexcpected blank node predicate")
			}
		} else if val.tokenType == BlankNodeOpening {
			newBlankNode2 := model.NewAnonymousBlankNode()
			if curPredicate != nil {
				curObject = newBlankNode2
			} else {
				panic("blank node as predicate not implemented")
			}
			// entering the blank node property list: save the current state by pushing to the stack, then initialize the current subject

			this.runInsideBlankNode(newBlankNode2.(*model.AnonymousBlankNode))
			continue
		} else if val.tokenType == CollectionOpening {

			newBlankNodeCollection := model.NewAnonymousBlankNode()
			if curPredicate == nil {
				panic("anonymous blanknode (from collection) cannot be a predicate")
			} else if curObject == nil {
				curObject = newBlankNodeCollection
				this.runInsideCollection(curObject.(*model.AnonymousBlankNode))
				continue
			}
		} else {
			panic("not implemented")
		}
	}
}

func (this *Parser) runInsideCollection(newBlankNode *model.AnonymousBlankNode) {

	curElm := model.BlankNode(newBlankNode)

	firstElm := true

	for {
		val, ok := <-this.source
		if !ok {
			panic("unexpected EOF in the collection")
		}
		if val.tokenType == CollectionClosing {
			this.target <- &model.Statement{
				Subject:   curElm,
				Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"),
				Object:    model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"),
			}
			return
		} else if (val.tokenType == PNameLN) || (val.tokenType == IRI) || (val.tokenType == A) || (val.tokenType == BlankNodeAnonymous) {
			var node model.RDFTerm
			switch val.tokenType {
			case A:
				node = model.A
			case IRI:
				node = model.IRI(val.value)
			case BlankNodeAnonymous:
				node = model.NewAnonymousBlankNode()
			case PNameLN:
				splt := strings.SplitN(val.value, ":", 2)
				prefix := splt[0] + ":"
				pnLocal := splt[1]
				node = this.newPrefixedName(prefix, pnLocal)
			}
			newElm := model.NewAnonymousBlankNode()
			if !firstElm {
				this.target <- &model.Statement{
					Subject:   curElm,
					Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"),
					Object:    newElm,
					Context:   this.curGraph,
				}
				curElm = newElm
			}
			this.target <- &model.Statement{
				Subject:   curElm,
				Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#first"),
				Object:    node,
				Context:   this.curGraph,
			}
			firstElm = false

		} else if val.tokenType == BlankNodeOpening {

			newBlankNode := model.NewAnonymousBlankNode()
			newElm := model.NewAnonymousBlankNode()
			if !firstElm {
				this.target <- &model.Statement{
					Subject:   curElm,
					Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"),
					Object:    newElm,
					Context:   this.curGraph,
				}
				curElm = newElm
			}
			this.target <- &model.Statement{
				Subject:   curElm,
				Object:    newBlankNode,
				Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#first"),
				Context:   this.curGraph,
			}

			firstElm = false
			this.runInsideBlankNode(newBlankNode.(*model.AnonymousBlankNode))
			continue

		} else if val.tokenType == CollectionOpening {

			newBlankNode := model.NewAnonymousBlankNode()
			newElm := model.NewAnonymousBlankNode()
			if !firstElm {
				this.target <- &model.Statement{
					Subject:   curElm,
					Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"),
					Object:    newElm,
					Context:   this.curGraph,
				}
				curElm = newElm
			}
			this.target <- &model.Statement{
				Subject:   curElm,
				Object:    newBlankNode,
				Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#first"),
				Context:   this.curGraph,
			}

			firstElm = false
			this.runInsideCollection(newBlankNode.(*model.AnonymousBlankNode))
			continue

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
