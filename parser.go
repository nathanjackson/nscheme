package main

import (
	"fmt"
	"strconv"
)

var symbolToOperator = map[string]Operator{
	"+": AddOp,
	"-": SubOp,
	"*": MulOp,
	"/": DivOp,
}

func Parse(offset int, types []TokenType, toks []string) (expr Expr, end int, err error) {
	switch types[offset] {
	case OpenParen:
		// Start a new expression
		expr, end, err = Parse(offset+1, types, toks)
	case BinOp:
		// LHS
		var lhs, rhs Expr
		lhs, end, err = Parse(offset+1, types, toks)
		// RHS
		rhs, end, err = Parse(end+1, types, toks)

		// Construct a binary operator
		var tmp BinOpExpr
		tmp.Operator = symbolToOperator[toks[offset]]
		tmp.lhs = lhs
		tmp.rhs = rhs
		expr = &tmp
		end += 1
	case Number:
		var tmp float64
		tmp, err = strconv.ParseFloat(toks[offset], 64)
		nexpr := NumExpr(tmp)
		expr = &nexpr
		end = offset
	case Keyword:
		switch toks[offset] {
		case Define.String():
			if types[offset+1] != Identifier {
				err = fmt.Errorf("define must be followed by an identifier")
			}
			var defexpr DefineExpr
			defexpr.name = toks[offset+1]
			defexpr.expression, end, err = Parse(offset+2, types, toks)
			expr = &defexpr
		}
	}

	return
}
