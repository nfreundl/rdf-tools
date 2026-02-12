/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */
package model

type prefixedName struct {
	Prefix     string
	Localname  string
	Namespaces map[Prefix]IRI
}
type PNameConstructor func(string, string) *prefixedName

func PrefixNameFactory(ns map[Prefix]IRI) PNameConstructor {
	return func(p string, lname string) *prefixedName {
		_, ok := ns[Prefix(p)]
		if !ok {
			panic("prefix not in the namespaces map")
		}

		return &prefixedName{Prefix: p, Localname: lname, Namespaces: ns}
	}
}

func (this *prefixedName) Equals(that RDFTerm) bool {
	thatPname, ok := that.(*prefixedName)
	if ok {
		return (this.Prefix == thatPname.Prefix) && (this.Localname == thatPname.Localname)
	}
	_, ok2 := that.(IRI)
	if ok2 {
		// TODO
		return false
	}
	return false

}
