/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Edouard Martin Freundler
 */
package model

type Statement struct {
	Subject   RDFTerm
	Predicate RDFTerm
	Object    RDFTerm
	Context   RDFTerm
}
