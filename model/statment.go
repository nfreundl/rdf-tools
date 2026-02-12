/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */
package model

type Statement struct {
	Subject   RDFTerm
	Predicate RDFTerm
	Object    RDFTerm
	Context   RDFTerm
}

func (this *Statement) Equals(that *Statement) bool {
	return (this.Subject.Equals(that.Subject)) && (this.Predicate.Equals(that.Predicate)) && (this.Object.Equals(that.Object)) && (this.Context.Equals(that.Context))
}
