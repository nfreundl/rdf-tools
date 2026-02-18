/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */
package parser

import (
	"fmt"
	"testing"

	"github.com/nfreundl/rdf-tools/model"
)

func TestParser(t *testing.T) {

	/*
		`@prefix pp: <http://example.com/> .
		[ ] a pp:xx , pp:yy , [ a pp:zz ] .
		pp:aa a pp:zz ; pp:has [ ] .
		`
	*/
	x := []*Token{
		{tokenType: PrefixTag},
		{value: "pp:", tokenType: PNameNS},
		{value: "<http://example.com/>", tokenType: IRI},
		{tokenType: Dot},
		{tokenType: BlankNodeAnonymous},
		{tokenType: A},
		{value: "pp:xx", tokenType: PNameLN},
		{tokenType: Coma},
		{value: "pp:yy", tokenType: PNameLN},
		{tokenType: Coma},
		{tokenType: BlankNodeOpening},
		{tokenType: A},
		{value: "pp:zz", tokenType: PNameLN},
		{tokenType: BlankNodeClosing},
		{tokenType: Dot},
		{tokenType: PNameLN, value: "pp:aa"},
		{tokenType: A},
		{value: "pp:zz", tokenType: PNameLN},
		{tokenType: SemiColumn},
		{tokenType: PNameLN, value: "pp:has"},
		{tokenType: BlankNodeAnonymous},
		{tokenType: Dot},
	}

	newPname := model.PrefixNameFactory(map[model.Prefix]model.IRI{"pp:": "http://example.com/pp"})

	expected := []*model.Statement{
		{Subject: &model.AnonymousBlankNode{}, Predicate: model.A, Object: newPname("pp:", "xx")},
		{Subject: &model.AnonymousBlankNode{}, Predicate: model.A, Object: newPname("pp:", "yy")},
		{Subject: &model.AnonymousBlankNode{}, Predicate: model.A, Object: &model.AnonymousBlankNode{}},
		{Subject: &model.AnonymousBlankNode{}, Predicate: model.A, Object: newPname("pp:", "zz")},
		{Subject: &model.AnonymousBlankNode{}, Predicate: model.A, Object: newPname("pp:", "zz")},
		{Subject: newPname("pp:", "aa"), Predicate: newPname("pp:", "has"), Object: &model.AnonymousBlankNode{}},
	}

	source := make(chan *Token)

	go func() {
		for _, tk := range x {
			source <- tk

		}
		close(source)
	}()

	target := make(chan *model.Statement)

	parser := NewParser(source, target)

	parser.Start()

	statements := []*model.Statement{}

	for statement := range target {
		fmt.Printf("got statement !\n")
		statements = append(statements, statement)
	}

	// first check the lengths
	if len(statements) != len(expected) {
		t.Errorf("The number of statements is not correct; should be %d, got %d", len(expected), len(statements))
	}

	// check deep equal of all non-blank nodes
	// no blank nodes in property
	for i := 0; i < len(expected); i++ {
		if !expected[i].Predicate.Equals(statements[i].Predicate) {
			t.Errorf("the %dth predicates are not the same", i+1)

		}
	}
	if !expected[0].Object.Equals(statements[0].Object) {
		t.Errorf("the firs.Equals(objects are not equal")
	}
	if !expected[1].Object.Equals(statements[1].Object) {
		t.Errorf("the second objects are not equal")
	}
	if !expected[3].Object.Equals(statements[3].Object) {
		t.Errorf("the fourth objects are not equal")
	}
	if !expected[4].Object.Equals(statements[4].Object) {
		t.Errorf("the fifth objects are not equal")
	}
	if !expected[5].Subject.Equals(statements[5].Subject) {
		t.Errorf("the sixth subjects are not equal")
	}

	// some blank nodes must be equal (pointer equality, maybe later we will use UUID for internal IDs)

	if !statements[0].Subject.Equals(statements[1].Subject) {
		t.Errorf("first and second subject blank nodes are not equal")
	}

	if !statements[1].Subject.Equals(statements[2].Subject) {
		t.Errorf("second subject and third subject blank node are not equal")
	}

	if !statements[2].Object.Equals(statements[3].Subject) {
		t.Errorf("second object and fourth subject blank node are not equal")
	}

	// some blank node cannot be equal

	if statements[0].Subject.Equals(statements[2].Object) {
		t.Errorf("first subject and third object are equal")
	}

	if statements[0].Subject.Equals(statements[5].Object) {
		t.Errorf("first subject and third object are equal")
	}

	if statements[2].Object.Equals(statements[5].Object) {
		t.Errorf("third object and third object are equal")
	}

}
