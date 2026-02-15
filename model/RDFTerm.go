/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */
package model

// RDF terms

type RDFTerm interface {
	Equals(RDFTerm) bool
	// string is not injective: prefixed name are turned into IRI, anonymous blank nodes are turned into labelled blank nodes
	String() string
}
