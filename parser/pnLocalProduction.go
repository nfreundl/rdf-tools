/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */
package parser

// returns the value and true if nothing happens, returns the next value after emitting tokens and true if the value is not exhausted, rune 0 and false if the source was exhausted
func (tokenizer *Tokenizer) pnLocalProduction(val rune) (rune, bool) {
	defer func() { tokenizer.curValue = "" }()
	ok := false
	if PN_CHARS_U.addRange('0', '9').contains(tokenizer.ifPlxEsc(val)) {
		tokenizer.curValue += string(val)

		val, ok = <-tokenizer.source
		if !ok {
			tokenizer.target <- &Token{tokenType: PNameLN, value: tokenizer.curValue}
			return 0, false
		}

		for PN_CHARS.copy().addRange('0', '9').add('.').add(':').contains(tokenizer.ifPlxEsc(val)) {
			tokenizer.curValue += string(val)
			val, ok = <-tokenizer.source
			if !ok {

				tokenizer.checkDottingAndProducePnLocal()
				return 0, false
			}
		}

		tokenizer.checkDottingAndProducePnLocal()

		return val, true

	} else {
		tokenizer.target <- &Token{tokenType: PNameNS, value: tokenizer.curValue}

		return val, true
	}
}

func (tokenizer *Tokenizer) checkDottingAndProducePnLocal() {
	if tokenizer.curValue[len(tokenizer.curValue)-1] == '.' {
		tokenizer.target <- &Token{tokenType: PNameLN, value: tokenizer.curValue[:len(tokenizer.curValue)-1]}
		tokenizer.target <- &Token{tokenType: Dot}
	} else {
		tokenizer.target <- &Token{tokenType: PNameLN, value: tokenizer.curValue}
	}
}
