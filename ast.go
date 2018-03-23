package main

import (
	"llvm.org/llvm/bindings/go/llvm"
)

// Expr is the interface that AST nodes must implement.
type Expr interface {
	Codegen(module llvm.Module, builder llvm.Builder) llvm.Value
}

// NumExpr represents a constant numeric value.
type NumExpr float64

func (num *NumExpr) Codegen(llvm.Module, llvm.Builder) llvm.Value {
	return llvm.ConstFloat(llvm.DoubleType(), float64(*num))
}

type Operator int

const (
	AddOp Operator = iota
	SubOp
	MulOp
	DivOp
)

type BinOpExpr struct {
	Operator
	lhs Expr
	rhs Expr
}

func (binop *BinOpExpr) Codegen(module llvm.Module, builder llvm.Builder) (op llvm.Value) {
	lhsIR := binop.lhs.Codegen(module, builder)
	rhsIR := binop.rhs.Codegen(module, builder)
	switch binop.Operator {
	case AddOp:
		op = builder.CreateFAdd(lhsIR, rhsIR, "addOp")
	case SubOp:
		op = builder.CreateFSub(lhsIR, rhsIR, "subOp")
	case MulOp:
		op = builder.CreateFMul(lhsIR, rhsIR, "mulOp")
	case DivOp:
		op = builder.CreateFDiv(lhsIR, rhsIR, "divOp")
	}
	return
}

type DefineExpr struct {
	name       string
	expression Expr
}

func (defn *DefineExpr) Codegen(module llvm.Module, builder llvm.Builder) llvm.Value {
	val := defn.expression.Codegen(module, builder)
	global := llvm.AddGlobal(module, val.Type(), defn.name)
	global.SetInitializer(val)
	return global
}

type LambdaExpr struct {
	args []string
	body Expr
}

func (lexpr *LambdaExpr) Codegen(module llvm.Module, builder llvm.Builder) llvm.Value {
	argTypes := []llvm.Type{}
	for i := 0; i < len(lexpr.args); i++ {
		argTypes = append(argTypes, llvm.DoubleType())
	}
	prototype := llvm.FunctionType(llvm.DoubleType(), argTypes, false)
	fn := llvm.AddFunction(module, "_lambda", prototype)
	block := llvm.AddBasicBlock(fn, "")
	builder.SetInsertPoint(block, block.FirstInstruction())
	builder.CreateRet(lexpr.body.Codegen(module, builder))
	return fn
}
