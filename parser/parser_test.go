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

func TestParserWithOneLevelOfBlankNodes(t *testing.T) {

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

func TestParserWithOneLevelOfCollection(t *testing.T) {
	/*
		`@prefix pp: <http://example.com/> .
		pp:a a pp:xx ; pp:has ( pp:b pp:c pp:d ) .
		`
	*/

	x := []*Token{
		{tokenType: PrefixTag},
		{value: "pp:", tokenType: PNameNS},
		{value: "<http://example.com/>", tokenType: IRI},
		{tokenType: Dot},
		{value: "pp:a", tokenType: PNameLN},
		{tokenType: A},
		{value: "pp:xx", tokenType: PNameLN},
		{tokenType: SemiColumn},
		{value: "pp:has", tokenType: PNameLN},
		{tokenType: CollectionOpening},
		{value: "pp:b", tokenType: PNameLN},
		{value: "pp:c", tokenType: PNameLN},
		{value: "pp:d", tokenType: PNameLN},
		{tokenType: CollectionClosing},
		{tokenType: Dot},
	}

	newPname := model.PrefixNameFactory(map[model.Prefix]model.IRI{"pp:": "http://example.com/pp"})
	first := model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#first")
	rest := model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#rest")
	nihil := model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#nil")
	expected := []*model.Statement{
		{Subject: newPname("pp:", "a"), Predicate: model.A, Object: newPname("pp:", "xx")},
		{Subject: &model.AnonymousBlankNode{}, Predicate: first, Object: newPname("pp:", "b")},
		{Subject: &model.AnonymousBlankNode{}, Predicate: rest, Object: &model.AnonymousBlankNode{}},
		{Subject: &model.AnonymousBlankNode{}, Predicate: first, Object: newPname("pp:", "c")},
		{Subject: &model.AnonymousBlankNode{}, Predicate: rest, Object: &model.AnonymousBlankNode{}},
		{Subject: &model.AnonymousBlankNode{}, Predicate: first, Object: newPname("pp:", "d")},
		{Subject: &model.AnonymousBlankNode{}, Predicate: rest, Object: nihil},
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "has"), Object: &model.AnonymousBlankNode{}},
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

	// then check equality of non-blank nodes
	for i := range expected {
		if !statements[i].Predicate.Equals(expected[i].Predicate) {
			t.Errorf("the predicates %d are not equal", i)
		}

	}

	// then check the equality of blank nodes that should be equal
	if !statements[1].Subject.Equals(statements[2].Subject) {
		t.Error("the subject of statement 1 should be equal to the subject of statement 2")
	}
	if !statements[2].Object.Equals(statements[3].Subject) {
		t.Error("the object of statement 2 should be equal to the subject of statement 3")
	}
	if !statements[2].Object.Equals(statements[4].Subject) {
		t.Error("the object of statement 2 should be equal to the subject of statement 4")
	}
	if !statements[4].Object.Equals(statements[5].Subject) {
		t.Error("the object of statement 4 should be equal to the subject of statement 5")
	}
	if !statements[4].Object.Equals(statements[6].Subject) {
		t.Error("the object of statement 4 should be equal to the subject of statement 6")
	}
	if !statements[7].Object.Equals(statements[1].Subject) {
		t.Error("the object of statement 7 should be equal to the subject of statement 1")
	}

	// then check the inequality of blank nodes that should not be equal
	if statements[1].Subject.Equals(statements[2].Object) {
		t.Error("the subject of statement 1 should not be equal to the object of statement 2")
	}
	if statements[2].Object.Equals(statements[4].Object) {
		t.Error("the object of statement 2 should not be equal to the object of statement 4")
	}
	if statements[4].Object.Equals(statements[1].Subject) {
		t.Error("the object of statement 4 should not be equal to the subject of statement 1")
	}
}

func TestTripleTerm(t *testing.T) {
	/*
		`@prefix pp: <http://example.com/> .
		pp:a pp:b <<( pp:c pp:d <<( pp:e pp:f pp:g )>> )>> .
		`
	*/

	x := []*Token{
		{tokenType: PrefixTag},
		{value: "pp:", tokenType: PNameNS},
		{value: "<http://example.com/>", tokenType: IRI},
		{tokenType: Dot},
		{value: "pp:a", tokenType: PNameLN},
		{value: "pp:b", tokenType: PNameLN},
		{tokenType: TripleTermOpening},
		{value: "pp:c", tokenType: PNameLN},
		{value: "pp:d", tokenType: PNameLN},
		{tokenType: TripleTermOpening},
		{value: "pp:e", tokenType: PNameLN},
		{value: "pp:f", tokenType: PNameLN},
		{value: "pp:g", tokenType: PNameLN},
		{tokenType: TripleTermClosing},
		{tokenType: TripleTermClosing},
		{tokenType: Dot},
	}

	newPname := model.PrefixNameFactory(map[model.Prefix]model.IRI{"pp:": "http://example.com/pp"})
	expected := []*model.Statement{
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: &model.TripleTerm{
			Subject:   newPname("pp:", "c"),
			Predicate: newPname("pp:", "d"),
			Object: &model.TripleTerm{
				Subject:   newPname("pp:", "e"),
				Predicate: newPname("pp:", "f"),
				Object:    newPname("pp:", "g"),
			},
		}},
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

	// there is no blank node

	if !statements[0].Equals(expected[0]) {
		t.Errorf("the expected statement and the obtained statement are different")
	}

}

func TestReifiedTriples(t *testing.T) {
	/*
		`@prefix pp: <http://example.com/> .
		pp:a pp:b << << pp:c pp:d pp:e >> pp:f pp:g ~ pp:h>> .
		`
	*/

	x := []*Token{
		{tokenType: PrefixTag},
		{value: "pp:", tokenType: PNameNS},
		{value: "<http://example.com/>", tokenType: IRI},
		{tokenType: Dot},
		{value: "pp:a", tokenType: PNameLN},
		{value: "pp:b", tokenType: PNameLN},
		{tokenType: ReifiedTripleOpening},
		{tokenType: ReifiedTripleOpening},
		{value: "pp:c", tokenType: PNameLN},
		{value: "pp:d", tokenType: PNameLN},
		{value: "pp:e", tokenType: PNameLN},
		{tokenType: ReifiedTripleClosing},
		{value: "pp:f", tokenType: PNameLN},
		{value: "pp:g", tokenType: PNameLN},
		{tokenType: ReifierTag},
		{value: "pp:h", tokenType: PNameLN},
		{tokenType: ReifiedTripleClosing},
		{tokenType: Dot},
	}

	newPname := model.PrefixNameFactory(map[model.Prefix]model.IRI{"pp:": "http://example.com/pp"})
	expected := []*model.Statement{
		{Subject: model.NewAnonymousBlankNode(), Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#reifies"), Object: &model.TripleTerm{
			Subject:   newPname("pp:", "c"),
			Predicate: newPname("pp:", "d"),
			Object:    newPname("pp:", "e"),
		}},
		{Subject: newPname("pp:", "h"), Predicate: model.IRI("http://www.w3.org/1999/02/22-rdf-syntax-ns#reifies"), Object: &model.TripleTerm{
			Subject:   model.NewAnonymousBlankNode(),
			Predicate: newPname("pp:", "f"),
			Object:    newPname("pp:", "g"),
		}},
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: newPname("pp:", "h")},
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

	// check the non-blank nodes
	if !statements[0].Predicate.Equals(expected[0].Predicate) {
		t.Errorf("should be equal")
	}
	if !statements[0].Object.Equals(expected[0].Object) {
		t.Errorf("should be equal")
	}
	if !statements[1].Subject.Equals(expected[1].Subject) {
		t.Errorf("should be equal")
	}
	if !statements[1].Predicate.Equals(expected[1].Predicate) {
		t.Errorf("should be equal")
	}
	if !statements[1].Object.(*model.TripleTerm).Predicate.Equals(expected[1].Object.(*model.TripleTerm).Predicate) {
		t.Errorf("should be equal")
	}
	if !statements[1].Object.(*model.TripleTerm).Object.Equals(expected[1].Object.(*model.TripleTerm).Object) {
		t.Errorf("should be equal")
	}

}

func TestGraphTriples(t *testing.T) {
	/*
		`@prefix pp: <http://example.com/> .
		pp:a pp:b pp:c .
		{ pp:a pp:b pp:c }
		{ pp:a pp:b pp:c . }
		GRAPH _:aa { pp:a pp:b pp:c }
		_:aa { pp:a pp:b pp:c . }
		`
	*/

	x := []*Token{
		{tokenType: PrefixTag},
		{value: "pp:", tokenType: PNameNS},
		{value: "<http://example.com/>", tokenType: IRI},
		{tokenType: Dot},
		{value: "pp:a", tokenType: PNameLN},
		{value: "pp:b", tokenType: PNameLN},
		{value: "pp:c", tokenType: PNameLN},
		{tokenType: Dot},
		{tokenType: GraphOpening},
		{value: "pp:a", tokenType: PNameLN},
		{value: "pp:b", tokenType: PNameLN},
		{value: "pp:c", tokenType: PNameLN},
		{tokenType: GraphClosing},
		{tokenType: GraphOpening},
		{value: "pp:a", tokenType: PNameLN},
		{value: "pp:b", tokenType: PNameLN},
		{value: "pp:c", tokenType: PNameLN},
		{tokenType: Dot},
		{tokenType: GraphClosing},
		{tokenType: Graph},
		{value: "aa", tokenType: BlankNodeLabel},
		{tokenType: GraphOpening},
		{value: "pp:a", tokenType: PNameLN},
		{value: "pp:b", tokenType: PNameLN},
		{value: "pp:c", tokenType: PNameLN},
		{tokenType: GraphClosing},
		{tokenType: Graph},
		{value: "aa", tokenType: BlankNodeLabel},
		{tokenType: GraphOpening},
		{value: "pp:a", tokenType: PNameLN},
		{value: "pp:b", tokenType: PNameLN},
		{value: "pp:c", tokenType: PNameLN},
		{tokenType: Dot},
		{tokenType: GraphClosing},
	}

	newPname := model.PrefixNameFactory(map[model.Prefix]model.IRI{"pp:": "http://example.com/pp"})

	expected := []*model.Statement{
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: newPname("pp:", "c")},
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: newPname("pp:", "c")},
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: newPname("pp:", "c")},
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: newPname("pp:", "c"), Context: model.NewBlankNodeFromLabel("aa")},
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: newPname("pp:", "c"), Context: model.NewBlankNodeFromLabel("aa")},
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

	for i, statement := range statements {
		if !statement.Equals(expected[i]) {
			t.Errorf("%d th statement is not equal to the expected", i)
		}
	}
}

func TestAssertedAndAnnotated(t *testing.T) {
	/*
		`@prefix pp: <http://example.com/> .
		pp:a pp:b pp:c ~ .
		pp:d pp:e pp:f ~ ~ .
		pp:g pp:h pp:i ~ {| pp:j pp:k , pp:l |} ~ pp:m .
		pp:n pp:o pp:p ~ pp:r {| pp:s pp:t |} .
		pp:u pp:v pp:w ~

		`
	*/

	x := []*Token{
		{tokenType: PrefixTag},
		{value: "pp:", tokenType: PNameNS},
		{value: "<http://example.com/>", tokenType: IRI},
		{tokenType: Dot},
		{value: "pp:a", tokenType: PNameLN},
		{value: "pp:b", tokenType: PNameLN},
		{value: "pp:c", tokenType: PNameLN},
		{tokenType: ReifierTag},
		{tokenType: Dot},
		{value: "pp:d", tokenType: PNameLN},
		{value: "pp:e", tokenType: PNameLN},
		{value: "pp:f", tokenType: PNameLN},
		{tokenType: ReifierTag},
		{tokenType: ReifierTag},
		{tokenType: Dot},
		{value: "pp:g", tokenType: PNameLN},
		{value: "pp:h", tokenType: PNameLN},
		{value: "pp:i", tokenType: PNameLN},
		{tokenType: ReifierTag},
		{tokenType: AnnotationOpening},
		{value: "pp:j", tokenType: PNameLN},
		{value: "pp:k", tokenType: PNameLN},
		{tokenType: Coma},
		{value: "pp:l", tokenType: PNameLN},
		{tokenType: AnnotationClosing},
		{tokenType: Dot},
		{tokenType: GraphClosing},
		{tokenType: GraphOpening},
		{value: "pp:a", tokenType: PNameLN},
		{value: "pp:b", tokenType: PNameLN},
		{value: "pp:c", tokenType: PNameLN},
		{tokenType: Dot},
		{tokenType: GraphClosing},
		{tokenType: Graph},
		{value: "aa", tokenType: BlankNodeLabel},
		{tokenType: GraphOpening},
		{value: "pp:a", tokenType: PNameLN},
		{value: "pp:b", tokenType: PNameLN},
		{value: "pp:c", tokenType: PNameLN},
		{tokenType: GraphClosing},
		{tokenType: Graph},
		{value: "aa", tokenType: BlankNodeLabel},
		{tokenType: GraphOpening},
		{value: "pp:a", tokenType: PNameLN},
		{value: "pp:b", tokenType: PNameLN},
		{value: "pp:c", tokenType: PNameLN},
		{tokenType: Dot},
		{tokenType: GraphClosing},
	}

	newPname := model.PrefixNameFactory(map[model.Prefix]model.IRI{"pp:": "http://example.com/pp"})

	expected := []*model.Statement{
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: newPname("pp:", "c")},
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: newPname("pp:", "c")},
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: newPname("pp:", "c")},
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: newPname("pp:", "c"), Context: model.NewBlankNodeFromLabel("aa")},
		{Subject: newPname("pp:", "a"), Predicate: newPname("pp:", "b"), Object: newPname("pp:", "c"), Context: model.NewBlankNodeFromLabel("aa")},
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

	for i, statement := range statements {
		if !statement.Equals(expected[i]) {
			t.Errorf("%d th statement is not equal to the expected", i)
		}
	}
}
