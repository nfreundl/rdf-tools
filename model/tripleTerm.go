/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */
package model

import "fmt"

type TripleTerm struct {
	Subject   RDFTerm
	Object    RDFTerm
	Predicate RDFTerm
}

func (this *TripleTerm) Equals(other RDFTerm) bool {
	if otherTripleTerm, ok := other.(*TripleTerm); ok {
		return this.Subject.Equals(otherTripleTerm.Subject) && this.Predicate.Equals(otherTripleTerm.Predicate) && this.Object.Equals(otherTripleTerm.Object)

	}
	return false

}

func (this *TripleTerm) String() string {
	return fmt.Sprintf("<<( %s %s %s )>>", this.Subject, this.Predicate, this.Object)
}
