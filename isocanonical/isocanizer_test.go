package isocanonical

import (
	"testing"

	"github.com/nfreundl/rdf-tools/model"
)

func TestWithOneAmbiguity(t *testing.T) {
	/*
		`@prefix pp: <http://example.com/> .
		[ ] a pp:xx , pp:yy , [ a pp:zz ] .
		pp:aa a pp:zz ; pp:has [ ] .
		[] a [] .
		[] a [] .
		`
	*/

	newPname := model.PrefixNameFactory(map[model.Prefix]model.IRI{"pp:": "http://example.com/pp"})
	graph := []*model.Statement{
		{Subject: model.NewAnonymousBlankNode(), Predicate: model.A, Object: newPname("pp:", "xx")},
		{Subject: model.NewAnonymousBlankNode(), Predicate: model.A, Object: newPname("pp:", "yy")},
		{Subject: model.NewAnonymousBlankNode(), Predicate: model.A, Object: model.NewAnonymousBlankNode()},
		{Subject: model.NewAnonymousBlankNode(), Predicate: model.A, Object: newPname("pp:", "zz")},
		{Subject: model.NewAnonymousBlankNode(), Predicate: model.A, Object: newPname("pp:", "zz")},
		{Subject: newPname("pp:", "aa"), Predicate: newPname("pp:", "has"), Object: model.NewAnonymousBlankNode()},
		{Subject: model.NewAnonymousBlankNode(), Predicate: model.A, Object: model.NewAnonymousBlankNode()},
		{Subject: model.NewAnonymousBlankNode(), Predicate: model.A, Object: model.NewAnonymousBlankNode()},
	}

	channel := make(chan *model.Statement)

	go func() {
		for _, v := range graph {
			channel <- v
		}
		close(channel)
	}()

	res := CanonicalizeGraph(channel)
	if len(res) != 7 {
		t.Errorf("length should be 7, got %d", len(res))
	}

}
