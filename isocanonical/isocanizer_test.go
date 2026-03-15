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
		{Subject: model.NewBlankNodeFromLabel("1"), Predicate: model.A, Object: newPname("pp:", "xx")},
		{Subject: model.NewBlankNodeFromLabel("1"), Predicate: model.A, Object: newPname("pp:", "yy")},
		{Subject: model.NewBlankNodeFromLabel("1"), Predicate: model.A, Object: model.NewBlankNodeFromLabel("2")},
		{Subject: model.NewBlankNodeFromLabel("2"), Predicate: model.A, Object: newPname("pp:", "zz")},
		{Subject: newPname("pp:", "aa"), Predicate: newPname("pp:", "has"), Object: model.NewBlankNodeFromLabel("3")},
		{Subject: model.NewBlankNodeFromLabel("4"), Predicate: model.A, Object: model.NewBlankNodeFromLabel("5")},
		{Subject: model.NewBlankNodeFromLabel("6"), Predicate: model.A, Object: model.NewBlankNodeFromLabel("7")},
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
