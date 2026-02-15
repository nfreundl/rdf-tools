/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */
package model

type IRI string
type Prefix string
type a_ string

const A a_ = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"

func (this IRI) Equals(that RDFTerm) bool {
	thatIri, ok := that.(IRI)
	if ok {
		return this == thatIri
	}
	return false

}

func (this a_) Equals(that RDFTerm) bool {
	thatA, ok := that.(a_)
	if ok {
		return this == thatA
	}
	return false
}

func (this IRI) String() string {
	return "<" + string(this) + ">"
}

func (this a_) String() string {
	return "<" + string(this) + ">"
}
