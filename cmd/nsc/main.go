package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nathanjackson/nscheme"

	"llvm.org/llvm/bindings/go/llvm"
)

func main() {
	var targetTriple string
	flag.StringVar(&targetTriple, "target", llvm.DefaultTargetTriple(), "target triple")
	var cpu string
	flag.StringVar(&cpu, "cpu", "generic", "cpu model")
	var features string
	flag.StringVar(&features, "features", "", "target features")
	flag.Parse()
	positionalArgs := flag.Args()

	if len(positionalArgs) < 1 {
		fmt.Println("No file specified")
		return
	}

	srcFile, err := os.Open(positionalArgs[0])
	if err != nil {
		fmt.Printf("Could not open input file: %v", err)
		return
	}
	defer srcFile.Close()
	lex := nscheme.NewLexer(bufio.NewReader(srcFile))

	tokens := []string{}
	tokenTypes := []nscheme.TokenType{}
	for lex.Scan() {
		tok, tt := lex.Token(), lex.TokenType()
		tokens = append(tokens, tok)
		tokenTypes = append(tokenTypes, tt)
	}

	expression, _, err := nscheme.Parse(0, tokenTypes, tokens)
	if err != nil {
		fmt.Printf("Could not parse expression: %v\n", err)
		return
	}

	llvm.InitializeNativeTarget()
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	llvm.InitializeAllAsmPrinters()
	llvm.InitializeAllAsmParsers()

	context := llvm.GlobalContext()
	module := llvm.NewModule(positionalArgs[0])
	builder := context.NewBuilder()
	expression.Codegen(module, builder)

	target, err := llvm.GetTargetFromTriple(targetTriple)
	if err != nil {
		fmt.Printf("Could not get target from triple: %v\n", err)
		return
	}
	targetMachine := target.CreateTargetMachine(targetTriple, cpu, features,
		llvm.CodeGenLevelDefault, llvm.RelocDefault,
		llvm.CodeModelDefault)
	module.SetTarget(targetTriple)

	pass := llvm.NewPassManager()
	targetMachine.AddAnalysisPasses(pass)

	pass.Run(module)

	memoryBuffer, err := targetMachine.EmitToMemoryBuffer(module, llvm.ObjectFile)
	if err != nil {
		panic("could not generate object code")
	}

	inputFile := filepath.Base(positionalArgs[0])
	ext := filepath.Ext(positionalArgs[0])
	outputName := inputFile[0:len(inputFile)-len(ext)] + ".o"

	outputFile, err := os.Create(outputName)
	if err != nil {
		fmt.Printf("Could not open output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, bytes.NewReader(memoryBuffer.Bytes()))
	if err != nil {
		fmt.Printf("Could not write to output file: %v\n", err)
		return
	}
}
