/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Edouard Martin Freundler
 */
package parser

import (
	"fmt"
)

func (tokenizer *Tokenizer) produceSpecialKeyword() {
	switch tokenizer.curValue {
	case "a":
		tokenizer.target <- &Token{tokenType: A}
		tokenizer.curValue = ""
	case "GRAPH":
		tokenizer.target <- &Token{tokenType: Graph}
		tokenizer.curValue = ""
	case "BASE":
		tokenizer.target <- &Token{tokenType: SparqlBaseTag}
		tokenizer.curValue = ""
	case "PREFIX":
		tokenizer.target <- &Token{tokenType: SparqlPrefixTag}
		tokenizer.curValue = ""
	default:

		panic(fmt.Sprintf("unexpectd %s", tokenizer.curValue))

	}
}

func (tokenizer *Tokenizer) checkDottingAndProducePnameNs() {

	dotting := tokenizer.curValue[len(tokenizer.curValue)-1] == '.'

	if dotting {
		tokenizer.curValue = tokenizer.curValue[:len(tokenizer.curValue)-1]

		tokenizer.produceSpecialKeyword()
		tokenizer.target <- &Token{tokenType: Dot}
		tokenizer.curValue = ""
	}
	// else keep curValue :)

}
