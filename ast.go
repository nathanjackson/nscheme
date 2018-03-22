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
