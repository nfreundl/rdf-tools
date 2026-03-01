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
	switch ctxPair := [2]RDFTerm{this.Context, that.Context}; {
	case ctxPair == [2]RDFTerm{nil, nil}:
		return (this.Subject.Equals(that.Subject)) && (this.Predicate.Equals(that.Predicate)) && (this.Object.Equals(that.Object))
	case ctxPair[0] == nil:
		return false
	case ctxPair[1] == nil:
		return false
	default:
		return (this.Subject.Equals(that.Subject)) && (this.Predicate.Equals(that.Predicate)) && (this.Object.Equals(that.Object)) && (this.Context.Equals(that.Context))

	}

}
