/*
* This work is released under CC BY-NC-SA 4.0
* Copyright © 2025 Nicolas Freundler
 */
package isocanonical

import (
	"slices"

	"github.com/nfreundl/rdf-tools/model"
)

type isocanizer struct {
	bnodeToStatements bnodeToStatements

	// initial maps
	blankNodeHashes map[string][][2]uint64
	otherHashes     map[string][][2]uint64
}

type bnodeToStatements map[string]struct {

	// blank node cannot be predicate, but they can be graph
	AsSubject []*model.Statement
	AsObject  []*model.Statement
	AsContext []*model.Statement
}

func newisocanizer() *isocanizer {
	return &isocanizer{
		blankNodeHashes:   make(map[string][][2]uint64),
		otherHashes:       make(map[string][][2]uint64),
		bnodeToStatements: make(bnodeToStatements),
	}
}

func (this *isocanizer) ingest(st *model.Statement) {

	bnode, ok := st.Subject.(model.BlankNode)
	if ok {
		this.blankNodeHashes[bnode.String()] = [][2]uint64{{0, 0}}
		if _, exists := this.bnodeToStatements[bnode.String()]; !exists {
			this.bnodeToStatements[bnode.String()] = struct {
				AsSubject []*model.Statement
				AsObject  []*model.Statement
				AsContext []*model.Statement
			}{}
		}
		ls := append(this.bnodeToStatements[bnode.String()].AsSubject, st)
		x := this.bnodeToStatements[bnode.String()]
		x.AsSubject = ls
		this.bnodeToStatements[bnode.String()] = x

	} else {
		this.otherHashes[st.Subject.String()] = [][2]uint64{hashTerm(st.Subject)}
	}

	bnode, ok = st.Predicate.(model.BlankNode)
	if ok {
		panic("blank node cannot be a blank node")

	} else {
		this.otherHashes[st.Predicate.String()] = [][2]uint64{hashTerm(st.Subject)}
	}

	bnode, ok = st.Object.(model.BlankNode)
	if ok {
		this.blankNodeHashes[bnode.String()] = [][2]uint64{{0, 0}}
		if _, exists := this.bnodeToStatements[bnode.String()]; !exists {
			this.bnodeToStatements[bnode.String()] = struct {
				AsSubject []*model.Statement
				AsObject  []*model.Statement
				AsContext []*model.Statement
			}{}
		}
		ls := append(this.bnodeToStatements[bnode.String()].AsObject, st)
		x := this.bnodeToStatements[bnode.String()]
		x.AsObject = ls
		this.bnodeToStatements[bnode.String()] = x

	} else {

		this.otherHashes[st.Object.String()] = [][2]uint64{hashTerm(st.Object)}
	}

	bnode, ok = st.Context.(model.BlankNode)
	if ok {
		this.blankNodeHashes[bnode.String()] = [][2]uint64{{0, 0}}
		if _, exists := this.bnodeToStatements[bnode.String()]; !exists {
			this.bnodeToStatements[bnode.String()] = struct {
				AsSubject []*model.Statement
				AsObject  []*model.Statement
				AsContext []*model.Statement
			}{}
		}
		ls := append(this.bnodeToStatements[bnode.String()].AsContext, st)
		x := this.bnodeToStatements[bnode.String()]
		x.AsContext = ls
		this.bnodeToStatements[bnode.String()] = x

	} else {

		if st.Context == nil {
			this.otherHashes[""] = [][2]uint64{hashNil()}
		} else {
			this.otherHashes[st.Context.String()] = [][2]uint64{hashTerm(st.Context)}
		}
	}
	return
}

func (this *isocanizer) runOneRoundDeterministicHashing(blankNodeHashes, otherHashes map[string][][2]uint64, blankNodeIndex, otherHashesIdx int) (map[string][][2]uint64, map[string][][2]uint64) {

	for bnodeID, statements := range this.bnodeToStatements {
		acc := [2]uint64{0, 0}
		for _, statement := range statements.AsSubject {
			predicate := this.getCurrentHash(blankNodeHashes, otherHashes, blankNodeIndex, otherHashesIdx, statement.Predicate)
			object := this.getCurrentHash(blankNodeHashes, otherHashes, blankNodeIndex, otherHashesIdx, statement.Object)
			context := this.getCurrentHash(blankNodeHashes, otherHashes, blankNodeIndex, otherHashesIdx, statement.Context)
			acc = hashBag(acc, hash3Tuple(predicate, object, context))
		}
		for _, statement := range statements.AsObject {
			predicate := this.getCurrentHash(blankNodeHashes, otherHashes, blankNodeIndex, otherHashesIdx, statement.Predicate)
			subject := this.getCurrentHash(blankNodeHashes, otherHashes, blankNodeIndex, otherHashesIdx, statement.Subject)
			context := this.getCurrentHash(blankNodeHashes, otherHashes, blankNodeIndex, otherHashesIdx, statement.Context)
			acc = hashBag(acc, hash3Tuple(subject, predicate, context))
		}
		// check if it is necessary to add this distinguisher here
		for _, statement := range statements.AsContext {
			predicate := this.getCurrentHash(blankNodeHashes, otherHashes, blankNodeIndex, otherHashesIdx, statement.Predicate)
			subject := this.getCurrentHash(blankNodeHashes, otherHashes, blankNodeIndex, otherHashesIdx, statement.Subject)
			object := this.getCurrentHash(blankNodeHashes, otherHashes, blankNodeIndex, otherHashesIdx, statement.Object)
			acc = hashBag(acc, hash3TupleWithDistinguisher(subject, predicate, object))
		}
		x := blankNodeHashes[bnodeID]
		x = append(x, acc)
		blankNodeHashes[bnodeID] = x
	}
	return blankNodeHashes, otherHashes
}

func (this *isocanizer) testCondition(blankNodeHashes, otherHashes map[string][][2]uint64, blankNodeIndex, otherHashesIdx int) bool {
	// for all x,y; hash[i][x] == hash[i][y] <> hash[i-1][x] == hash[i-1][y]
	firstCond := true
	for _, hashes1 := range blankNodeHashes {
		for _, hashes2 := range blankNodeHashes {
			if (hashes1[blankNodeIndex] == hashes2[blankNodeIndex]) && (hashes1[blankNodeIndex-1] != hashes2[blankNodeIndex-1]) {
				firstCond = false
				break
			}
			if (hashes1[blankNodeIndex] != hashes2[blankNodeIndex]) && (hashes1[blankNodeIndex-1] == hashes2[blankNodeIndex-1]) {
				firstCond = false
				break
			}

		}

		if !firstCond {
			break
		}
	}

	secondCondition := true
	for bnode1, hashes1 := range blankNodeHashes {
		for bnode2, hashes2 := range blankNodeHashes {
			if (hashes1[blankNodeIndex] == hashes2[blankNodeIndex]) && (bnode1 != bnode2) {
				secondCondition = false
				break
			}
			if (hashes1[blankNodeIndex] != hashes2[blankNodeIndex]) && (bnode1 == bnode2) {
				secondCondition = false
				break
			}

		}
		if !secondCondition {
			break
		}

		for _, termHashes := range otherHashes {
			if hashes1[blankNodeIndex] == termHashes[otherHashesIdx] {
				secondCondition = false
				break
			}
		}
		if !secondCondition {
			break
		}

	}

	return firstCond || secondCondition
}

func (this *isocanizer) getCurrentHash(blankNodeHashes, otherHashes map[string][][2]uint64, blankNodeIndex, otherHashesIdx int, term model.RDFTerm) [2]uint64 {
	if bnode, ok := term.(model.BlankNode); ok {
		return blankNodeHashes[bnode.String()][blankNodeIndex]
	}
	return otherHashes[term.String()][otherHashesIdx]
}

// run algorithm 1 in paper
func (this *isocanizer) runDeterministicHashing(blankNodeHashes, otherHashes map[string][][2]uint64) (blankNodeDeterministHashes map[string][2]uint64) {
	blankNodeIdx := 0
	otherNodeIdx := 0

	for cond := true; cond; cond = this.testCondition(blankNodeHashes, otherHashes, blankNodeIdx, otherNodeIdx) {
		blankNodeIdx++
		this.runOneRoundDeterministicHashing(blankNodeHashes, otherHashes, blankNodeIdx, otherNodeIdx)
	}

	blankNodeDeterministHashes = make(map[string][2]uint64)

	for bnode, hashes := range blankNodeHashes {
		blankNodeDeterministHashes[bnode] = hashes[blankNodeIdx]
	}

	return

}

// return true if the partition is fine (which means that all subset have a cardinality of one), else it returns the smallest partition
func (this *isocanizer) getSmallestNonTrivialSetIfPartitionIfNotFine(blankNodeDeterministHashes map[string][2]uint64) (bool, []string) {

	partition := make(map[[2]uint64][]string)
	nofBnodes := 0
	for bnode, hash := range blankNodeDeterministHashes {
		if _, ok := partition[hash]; !ok {
			partition[hash] = []string{}
		}
		x := partition[hash]
		x = append(x, bnode)
		partition[hash] = x
		nofBnodes++
	}

	minLength := nofBnodes + 1
	condition := true
	var candidate []string
	var candidateHash [2]uint64

	for hash, subSet := range partition {
		if (len(subSet) != 1) && (len(subSet) <= minLength) {
			condition = false
			// sort the smallest set,  if the size is equal , use the hash
			if (len(subSet) < minLength) || greater(candidateHash, hash) {
				candidate = subSet
				candidateHash = hash
				minLength = len(subSet)
			}
		}

	}
	return condition, candidate

}

func (this *isocanizer) prepareFromPreviousResult(blankNodeDeterministHashes map[string][2]uint64) (blankNodeHashes map[string][][2]uint64) {
	blankNodeHashes = map[string][][2]uint64{}

	for bnode, hash := range blankNodeDeterministHashes {
		blankNodeHashes[bnode] = [][2]uint64{hash}
	}
	return
}

func isLower(a, b map[string][2]uint64) bool {
	arrayA := [][2]uint64{}
	for _, hashA := range a {
		arrayA = append(arrayA, [2]uint64{hashA[0], hashA[1]})
	}
	arrayB := [][2]uint64{}
	for _, hashB := range b {
		arrayB = append(arrayB, [2]uint64{hashB[0], hashB[1]})
	}

	if len(arrayA) != len(arrayB) {
		panic("asserted to have the same length")
	}

	sortFunc := func(a, b [2]uint64) int {
		if (a[0] == b[0]) && (a[1] == b[1]) {
			return 0
		}
		if greater(a, b) {
			return -1
		}
		return 1
	}
	slices.SortFunc(arrayA, sortFunc)
	slices.SortFunc(arrayB, sortFunc)

	for i := range arrayA {

		switch sortFunc(arrayA[i], arrayB[i]) {
		case 1:
			return true
		case -1:
			return false
		case 0:
			continue
		}

	}
	return false
}

func (this *isocanizer) distinguish(otherHashes map[string][][2]uint64, blankNodeDeterministHashes map[string][2]uint64, smallestNonTrivial []string, candidateHashes map[string][2]uint64) map[string][2]uint64 {
	for _, bnode := range smallestNonTrivial {
		hash1 := make(map[string][2]uint64)
		for b, h := range blankNodeDeterministHashes {
			hash1[b] = [2]uint64{h[0], h[1]}
		}
		hash1[bnode] = hashWithDistinguisher(hash1[bnode])
		hash2 := this.runDeterministicHashing(this.prepareFromPreviousResult(hash1), otherHashes)
		if isFine, smallestNonTrivial1 := this.getSmallestNonTrivialSetIfPartitionIfNotFine(hash2); isFine {
			// because different branches may give valid results, we need to choose the lowest one (to stay determinist)
			if (candidateHashes == nil) || (isLower(hash2, candidateHashes)) {
				candidateHashes = hash2
			}
		} else {
			candidateHashes = this.distinguish(otherHashes, hash2, smallestNonTrivial1, candidateHashes)
		}

	}
	return candidateHashes
}

func CanonicalizeGraph(statements <-chan *model.Statement) map[string][2]uint64 {
	isocanizer := *newisocanizer()
	for statement := range statements {
		isocanizer.ingest(statement)
	}

	bnodesToHashes := isocanizer.runDeterministicHashing(isocanizer.blankNodeHashes, isocanizer.otherHashes)
	if isFine, smallestNonTrivial := isocanizer.getSmallestNonTrivialSetIfPartitionIfNotFine(bnodesToHashes); isFine {
		return bnodesToHashes
	} else {
		return isocanizer.distinguish(isocanizer.otherHashes, bnodesToHashes, smallestNonTrivial, nil)
	}

}
