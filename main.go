package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"llvm.org/llvm/bindings/go/llvm"
)

func main() {
	// Setup LLVM
	context := llvm.GlobalContext()

	fmt.Println("Nate's Scheme\n")
	run := true
	rd := bufio.NewReader(os.Stdin)

	for run {
		module := context.NewModule("top")
		builder := context.NewBuilder()

		tt := []TokenType{}
		toks := []string{}

		fmt.Print("nscheme> ")
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			fmt.Println("\nGoodbye!")
			break
		}

		lexer := NewLexer(strings.NewReader(line))

		for lexer.Scan() {
			tt = append(tt, lexer.TokenType())
			toks = append(toks, lexer.Token())
		}

		fmt.Printf("%v\n", tt)
		fmt.Printf("%v\n", toks)

		if err := lexer.Err(); err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		expr, _, err := Parse(0, tt, toks)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		prototype := llvm.FunctionType(llvm.DoubleType(), []llvm.Type{}, false)
		llvm.AddFunction(module, "expr", prototype)
		block := llvm.AddBasicBlock(module.NamedFunction("expr"), "entry")
		builder.SetInsertPoint(block, block.FirstInstruction())

		exprBody := expr.Codegen(module, builder)
		builder.SetInsertPointAtEnd(block)
		builder.CreateRet(exprBody)

		module.Dump()

		/*ee, err := llvm.NewExecutionEngine(module)
		if err != nil {
			fmt.Printf("Could not create ExecutionEngine: %v\n", err)
		}
		fun := ee.FindFunction("expr")
		result := ee.RunFunction(fun, []llvm.GenericValue{})
		fmt.Printf("%v\n", result.Float(llvm.DoubleType()))*/
	}
}
