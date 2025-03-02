package parser

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTokenizer(t *testing.T) {
	x := `@prefix pp: <http://example.com/> .
[ ] a pp:xx , pp:yy , [ a pp:zz ] .
	`

	expected := []*Token{
		{tokenType: PrefixTag},
		{value: "pp:", tokenType: PNameNS},
		{value: "<http://example.com/>", tokenType: IRI},
		{tokenType: Point},
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
		{tokenType: Point},
	}
	source := make(chan rune)

	go func() {
		for _, v := range x {
			source <- v
		}
		close(source)
	}()

	target := make(chan *Token)

	tokenizer := NewTokenizer(source, target)
	tokenizer.start()
	result := []*Token{}
	for token := range target {
		fmt.Printf("got token %v %s\n", token.tokenType, token.value)
		result = append(result, token)
	}

	if len(expected) != len(result) {
		t.Errorf("the length of expected is not the length of result: %d, %d", len(expected), len(result))
	}

	if !reflect.DeepEqual(expected, result) {
		t.Error("the content of expected is not the same of this of result")
	}

}
