/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Edouard Martin Freundler
 */
package parser

import (
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
}

func newParser(source <-chan *Token, target chan<- *model.Statement) *Parser {
	return &Parser{
		source: source,
		target: target,
		// all the rest is nil !
	}

}

func (this *Parser) run() {
	defer close(this.target)
	val := <-this.source
	for {

		if val.tokenType == Dot {
			this.target <- &model.Statement{
				Subject:   this.curSubject,
				Object:    this.curObject,
				Predicate: this.curPredicate,
				Context:   this.curGraph,
			}
			this.curSubject = nil
			this.curPredicate = nil
			this.curObject = nil
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
		} else if val.tokenType == BlankNodeClosing {
			this.target <- &model.Statement{
				Subject:   this.curSubject,
				Object:    this.curObject,
				Predicate: this.curPredicate,
				Context:   this.curGraph,
			}
			this.curSubject = nil
			this.curPredicate = nil
			this.curObject = nil

		} else if val.tokenType == BlankNodeOpening {
			newBlankNode := &model.AnonymousBlankNode{}
			if (this.curSubject != nil) && (this.curPredicate != nil) {
				this.curObject = newBlankNode
				this.target <- &model.Statement{
					Subject:   this.curSubject,
					Object:    this.curObject,
					Predicate: this.curPredicate,
					Context:   this.curGraph,
				}

			} else if (this.curSubject == nil) && (this.curPredicate == nil) {
				this.curSubject = newBlankNode

			} else {
				panic("not implemented")
			}

			this.runInsideBlankNode(newBlankNode)

		} else if val.tokenType == BlankNodeClosing {
			panic("unexpected ]")
		} else if val.tokenType == PNameNS {

			if this.curSubject == nil {
				this.curSubject = &model.PrefixedName{Prefix: val.value}
			} else if this.curPredicate == nil {
				this.curPredicate = &model.PrefixedName{Prefix: val.value}
			} else {
				this.curObject = &model.PrefixedName{Prefix: val.value}
			}
			//|| (val.tokenType == IRI) || (val.tokenType == A)

		} else if val.tokenType == PNameLN {

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
		} else if val.tokenType == CollectionOpening {

			if (this.curSubject == nil) && (this.curPredicate == nil) {
				this.curSubject = &model.AnonymousBlankNode{}
				this.runInsidePropertyList(this.curSubject.(*model.AnonymousBlankNode))
			} else if (this.curSubject != nil) && (this.curPredicate != nil) && (this.curObject == nil) {
				this.curObject = &model.AnonymousBlankNode{}
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
	val := <-this.source

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

	curElm := newBlankNode
	val := <-this.source
	for {
		if val.tokenType == CollectionClosing {
			this.target <- &model.Statement{
				Subject:   curElm,
				Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest"),
				Object:    model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#nil"),
			}
		} else if (val.tokenType == PNameLN) || (val.tokenType == PNameNS) || (val.tokenType == IRI) || (val.tokenType == A) {
			newElm := &model.AnonymousBlankNode{}
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
