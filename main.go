package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nfreundl/rdf-tools/isocanonical"
	"github.com/nfreundl/rdf-tools/model"
	"github.com/nfreundl/rdf-tools/parser"
)

func main() {
	inputFile := flag.String("input", "", "input file")
	outputFile := flag.String("output", "", "output file")
	canonicalFile := flag.String("canonicalMap", "", "canonical map")

	flag.Parse()
	filedesc, err := os.Open(*inputFile)
	if err != nil {
		panic("cannot open inputfile")
	}

	runeChan := parser.NewRuneReader(filedesc, 1024, 1024, 1024)

	tokenizerChan := make(chan *parser.Token, 1024)

	tokenizer := parser.NewTokenizer(runeChan, tokenizerChan)
	tokenizer.Start()

	tripleChan := make(chan *model.Statement, 1024)

	p := parser.NewParser(tokenizerChan, tripleChan)
	p.Start()

	outputFileDesc, err := os.OpenFile(*outputFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		panic("cannot open outputfile")
	}

	statements := []*model.Statement{}
	for statement := range tripleChan {
		fmt.Fprint(outputFileDesc, "%v\n", statement)
		statements = append(statements, statement)
	}

	tripleChan2 := make(chan *model.Statement, 1024)

	go func() {
		for _, statement := range statements {
			tripleChan2 <- statement
		}
	}()
	res := isocanonical.CanonicalizeGraph(tripleChan2)

	canonicalFileDesc, err := os.OpenFile(*canonicalFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(0644))
	if err != nil {
		panic("cannot open canonicalFile")
	}
	for key, value := range res {
		canonicalFileDesc.WriteString(fmt.Sprintf("{%v:%v}", key, value))
	}

}
