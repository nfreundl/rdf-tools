package graph

import "github.com/nfreundl/rdf-tools/model"

type Graph interface {

	// only the context can be nil, it means a triple without context
	Insert(*model.Statement) error

	// any can be nil, means any subject, predicate, object or context. nil nil nil nil means SELECT *
	Select(*model.Statement) ([]*model.Statement, error)

	// only the context can be nil, it means a triple without conte
	Delete(*model.Statement) error
}

type DummyGraph struct {
}

func (this *DummyGraph) Insert(*model.Statement) error {
	return nil

}
func (this *DummyGraph) Select(*model.Statement) ([]model.Statement, error) {
	return []model.Statement{}, nil
}
func (this *DummyGraph) Delete(*model.Statement) error {
	return nil
}

type FlatGraph struct {
	namespace  map[model.Prefix]model.IRI
	statements []*model.Statement
}

func (this *FlatGraph) Insert(st *model.Statement) error {
	for _, v := range this.statements {
		if v.Equals(st) {
			return nil
		}
	}
	this.statements = append(this.statements, st)
	return nil
}
func (this *FlatGraph) Select(st *model.Statement) ([]*model.Statement, error) {
	ret := []*model.Statement{}
	for _, v := range this.statements {
		if (st.Subject != nil) && !(st.Subject.Equals(v.Subject)) {
			continue
		}
		if (st.Predicate != nil) && !(st.Predicate.Equals(v.Predicate)) {
			continue
		}
		if (st.Object != nil) && !(st.Object.Equals(v.Object)) {
			continue
		}
		if (st.Context != nil) && !(st.Context.Equals(v.Context)) {
			continue
		}
		ret = append(ret, v)
	}
	return ret, nil
}
func (this *FlatGraph) Delete(st *model.Statement) error {
	idx := -1

	for i, v := range this.statements {
		if v.Equals(st) {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil
	}
	// length := len(this.statements)

	left := this.statements[:idx]
	right := this.statements[idx+1:]
	this.statements = append(left, right...)
	return nil
}

// using maps of maps of maps instead of one big array
type MultiIndexGraph struct {
}
