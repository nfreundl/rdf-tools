/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Edouard Martin Freundler
 */
package parser

type TokenType int

const (
	Prefix TokenType = iota
	Base
	IRI
	A
	String
	BlankNodeLabel
	BlankNodeOpening
	BlankNodeClosing
	BlankNodeAnonymous
	CollectionOpening
	CollectionClosing
	EmptyCollection
	Boolean
	Graph
	Point
	GraphOpening
	GraphClosing
	TripleTermOpening
	TripleTermClosing
	ReifiedTripleOpening
	ReifiedTripleClosing
	AnnotationOpening
	AnnotationClosing
	Number
	PNameNS
	PNameLN
	PrefixTag
	SemiColumn
	Coma
	Dot
	BaseTag
	SparqlBaseTag
	SparqlPrefixTag
)
const BufferSize int = 1 << 30

type Token struct {
	value     string
	tokenType TokenType
}

// var WS = [4]byte

var QUOTES = map[byte]struct{}{'"': struct{}{}, '\'': struct{}{}}

type Tokenizer struct {
	source    <-chan rune
	target    chan<- *Token
	tmp       []rune
	toResolve []rune
	curValue  string
}

func NewTokenizer(source <-chan rune, target chan<- *Token) *Tokenizer {
	return &Tokenizer{
		source: source,
		target: target,
		// longest terminal literals are """ ''' <<( and )>>
		tmp:       make([]rune, 0, 3),
		toResolve: make([]rune, 0, 3),
		curValue:  "",
	}
}

func (this *Tokenizer) start() {
	go this.run()
}

func (this *Tokenizer) run() {
	defer close(this.target)
	val, ok := <-this.source
	if !ok {
		return
	}
	for {

		// wild white space

		if WS.contains(val) {
			val, ok = <-this.source
			if !ok {
				return
			}
		}

		// double quote string

		if val == '"' {
			this.curValue += string(val)
			val, ok = <-this.source
			if !ok {
				panic("unexpected EOF")
			}
			if val == '"' {
				this.curValue += string(val)
				val, ok = <-this.source
				if !ok {
					return
				}
				if val == '"' {
					this.curValue += string(val)

					//we are in a long double quote

					nofConsecutiveQuotes := 0
					for nofConsecutiveQuotes != 3 {
						val, ok = <-this.source
						if !ok {
							panic("unexpected EOF")
						}
						if val == '"' {
							nofConsecutiveQuotes += 1
						} else {
							nofConsecutiveQuotes = 0
						}

						this.curValue += string(val)
					}
					this.target <- &Token{
						value:     this.curValue,
						tokenType: String,
					}
					this.curValue = ""

				} else {
					// empty double quote
					this.target <- &Token{
						value:     this.curValue,
						tokenType: String,
					}
					this.curValue = ""
				}
			}

		}

		// simple quote string

		if val == '\'' {
			this.curValue += string(val)
			val, ok = <-this.source
			if !ok {
				panic("unexpected EOF")
			}
			if val == '\'' {
				this.curValue += string(val)
				if !ok {
					return
				}

				if val == '\'' {
					// we are in a long single quoted string
					nofConsecutiveQuotes := 0
					for nofConsecutiveQuotes != 3 {
						if val == '\'' {
							nofConsecutiveQuotes += 1
						} else {
							nofConsecutiveQuotes = 0
						}

						this.curValue += string(val)
						val = <-this.source
					}
					this.target <- &Token{
						value:     this.curValue,
						tokenType: String,
					}
					this.curValue = ""
				} else {
					// empty single quote
					this.target <- &Token{
						value:     this.curValue,
						tokenType: String,
					}
					this.curValue = ""
				}
			}

		}

		// blank node opening

		if val == '[' {
			this.curValue += string(val)
			val = <-this.source
			// skip white spaces
			_, ok := WS[val]
			for ok {

				val = <-this.source
				_, ok = WS[val]

			}

			if val == ']' {
				this.target <- &Token{
					tokenType: BlankNodeAnonymous,
				}
				this.curValue = ""
				val, ok = <-this.source
				if !ok {
					return
				}

			} else {
				this.target <- &Token{
					tokenType: BlankNodeOpening,
				}
				this.curValue = ""

			}

		}

		// IRI or reified

		if val == '<' {
			this.curValue += string(val)
			val = <-this.source
			if val == '<' {
				this.curValue += string(val)
				val = <-this.source
				if val == '(' {
					this.target <- &Token{tokenType: TripleTermOpening}
					this.curValue = ""
				} else {
					this.target <- &Token{tokenType: ReifiedTripleOpening}
					this.curValue = ""
				}
			} else {
				this.curValue += string(val)
				val = <-this.source
				for val != '>' {
					if forbiddenInIRI.contains(val) {
						panic("unexpected character i IRI")
					}
					this.curValue += string(val)
					val = this.ifUcharrEsc(val)
					val = <-this.source
				}
				this.curValue += string(val)

				this.target <- &Token{tokenType: IRI, value: this.curValue}
				this.curValue = ""
				val = <-this.source
			}
		}

		// closing )

		if val == ')' {
			this.curValue += string(val)
			val = <-this.source
			if val == '>' {
				this.curValue += string(val)
				val = <-this.source
				if val == '>' {
					this.target <- &Token{tokenType: TripleTermClosing}
					this.curValue = ""
				}
				// else error ?
			} else {
				this.target <- &Token{tokenType: CollectionClosing}
			}
		}

		// closing >

		if val == '>' {
			this.curValue += string(val)
			val = <-this.source
			if val == '>' {
				this.target <- &Token{tokenType: ReifiedTripleClosing}
				this.curValue = ""
			}
		}

		// closing ]

		if val == ']' {
			this.target <- &Token{tokenType: BlankNodeClosing}
			val, ok = <-this.source
			if !ok {
				return
			}
		}

		// {

		if val == '{' {
			this.curValue += string(val)
			val = <-this.source
			if val == '|' {
				this.target <- &Token{tokenType: AnnotationOpening}

			} else {
				this.target <- &Token{tokenType: GraphOpening}
			}
		}

		// _

		if val == '_' {
			this.curValue += string(val)
			val = <-this.source
			if val == ':' {
				this.curValue += string(val)
				val = <-this.source
				if ((val >= '0') && (val <= '9')) || PN_CHARS_U.contains(val) {
					this.curValue += string(val)
					val = <-this.source
					for PN_CHARS.copy().add('.').contains(val) {
						this.curValue += string(val)
						val = <-this.source
					}
					if this.curValue[len(this.curValue)-1] == '.' {
						panic("blank node label cannot finish with '.'")
					} else {
						this.target <- &Token{value: this.curValue, tokenType: BlankNodeLabel}
						this.curValue = ""
					}

				} else {
					panic("")
				}

			}

		}

		// numbers
		if (val == '+') || (val == '-') || ((val >= '0') && (val <= '9')) {
			this.curValue += string(val)
			val = <-this.source
			for (val >= '0') && (val <= '9') {
				this.curValue += string(val)
				val = <-this.source
			}
			if val == '.' {
				this.curValue += string(val)
				val = <-this.source
			}
			for (val >= '0') && (val <= '9') {
				this.curValue += string(val)
				val = <-this.source

			}
			if (val == 'e') || (val == 'E') {
				if (val == '+') || (val == '-') {
					this.curValue += string(val)
					val = <-this.source

				}
				for (val >= '0') && (val <= '9') {
					this.curValue += string(val)
					val = <-this.source

				}

			}
			this.target <- &Token{
				value:     this.curValue,
				tokenType: Number,
			}

		}

		// prefix and possible name

		// PN_CHARS_BASE ((PN_CHARS | '.')* PN_CHARS)?
		if PN_CHARS_BASE.contains(val) {
			this.curValue += string(val)
			val, ok = <-this.source
			if !ok {
				this.checkDottingAndProducePnameNs()
				return
			}
			for ok && PN_CHARS.copy().add('.').contains(val) {
				this.curValue += string(val)
				val, ok = <-this.source
			}
			if !ok {
				return
			}
			this.checkDottingAndProducePnameNs()
			if val == ':' {
				this.curValue += string(val)
				val = <-this.source
				val, ok = this.pnLocalProduction(val)
				if !ok {
					return
				}
			} else {
				this.produceSpecialKeyword()
				this.curValue = ""
			}

		}

		// @ lang dir or base or prefix tag
		// '@' [a-zA-Z]+ ('-' [a-zA-Z0-9]+)* ('--' [a-zA-Z]+)?
		if val == '@' {
			this.curValue += string(val)
			val = <-this.source
			for (val >= 'a' && val <= 'z') || (val >= 'A' && val >= 'Z') {
				this.curValue += string(val)
				val = <-this.source
				if this.curValue == "@base" {
					this.target <- &Token{tokenType: BaseTag}
					this.curValue = ""
					break
				}
				if this.curValue == "@prefix" {
					this.target <- &Token{tokenType: PrefixTag}
					this.curValue = ""
					break
				}
			}
			for val == '-' {
				this.curValue += string(val)
				val = <-this.source

				if val != '-' {
					if !(val >= 'a' && val <= 'z') || (val >= 'A' && val >= 'Z') || (val >= '0' && val <= '9') {
						panic("must be a alphadigit")
					}
					for (val >= 'a' && val <= 'z') || (val >= 'A' && val >= 'Z') || (val >= '0' && val <= '9') {
						this.curValue += string(val)
						val = <-this.source
					}
				} else {
					this.curValue += string(val)
					val = <-this.source
					if (val >= 'a' && val <= 'z') || (val >= 'A' && val >= 'Z') {
						panic("alphabet is needed here")
					}
					for (val >= 'a' && val <= 'z') || (val >= 'A' && val >= 'Z') {
						this.curValue += string(val)
						val = <-this.source
					}
					break
				}
			}
		}

		// collection

		if val == '(' {
			val = <-this.source
			_, ok := WS[val]

			for ok {
				val = <-this.source
				_, ok = WS[val]
			}
			if val == ')' {
				this.target <- &Token{tokenType: EmptyCollection}
				val = <-this.source
			} else {
				this.target <- &Token{tokenType: CollectionOpening}
				val = <-this.source
			}
		}

		if val == ',' {
			this.target <- &Token{tokenType: Coma}
			val, ok = <-this.source
			if !ok {
				panic("unexpected EOF")
			}
		}
		if val == ';' {
			this.target <- &Token{tokenType: SemiColumn}
			val, ok = <-this.source
			if !ok {
				panic("unexpected EOF")
			}
		}
		if val == '.' {
			this.target <- &Token{tokenType: Point}
			val, ok = <-this.source
			if !ok {
				return
			}

		}

	}

}

// never complains if the following is not i th expected rtbnf'"\
func (this *Tokenizer) ifEcharEsc(val rune) rune {
	for val == '\\' {
		this.curValue += string(val)
		val = <-this.source
		this.curValue += string(val)
		val = <-this.source
	}
	return val
}

// does not complains in string. complains in IRI
func (this *Tokenizer) ifUcharrEsc(val rune) rune {
	for val == '\\' {
		this.curValue += string(val)
		val = <-this.source
		if val == 'u' {
			this.curValue += string(val)
			val = <-this.source
			for i := 0; i < 4; i++ {
				if HEX.contains(val) {
					this.curValue += string(val)
					val = <-this.source
				} else {
					panic("unexpected escaped char")
				}

			}

		} else if val == 'U' {
			this.curValue += string(val)
			val = <-this.source
			for i := 0; i < 8; i++ {
				if HEX.contains(val) {
					this.curValue += string(val)
					val = <-this.source
				} else {
					panic("unexpected escaped char")
				}

			}

		} else {
			panic("unexpected escaped char")
		}

	}
	return val
}

func (this *Tokenizer) ifEcharorUcharEsc(val rune) rune {
	for val == '\\' {
		this.curValue += string(val)
		val = <-this.source
		if (val == 'u') || (val == 'U') {
			this.curValue += string(val)
			val = <-this.source

			for HEX.contains(val) {

				this.curValue += string(val)
				val = <-this.source
			}
		} else {
			this.curValue += string(val)
			val = <-this.source
		}

	}
	return val
}

// always complains
func (this *Tokenizer) ifPlxEsc(val rune) rune {
	for (val == '\\') || (val == '%') {
		if val == '\\' {
			this.curValue += string(val)
			val = <-this.source
			if !PN_LOCAL_ESC.contains(val) {
				panic("")
			}
			this.curValue += string(val)
			val = <-this.source
		} else {
			this.curValue += string(val)
			val = <-this.source
			if !HEX.contains(val) {
				panic("")
			}
			this.curValue += string(val)
			val = <-this.source
			if !HEX.contains(val) {
				panic("")
			}
			this.curValue += string(val)
			val = <-this.source

		}
	}
	return val
}
