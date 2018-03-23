package main

import (
	"bufio"
	"fmt"
	"io"
	"unicode"

	"golang.org/x/text/unicode/rangetable"
)

const (
	OpenParen TokenType = iota
	CloseParen
	Identifier
	Number
	BinOp
	Keyword
)

const (
	Define KeywordType = iota
	Lambda
)

func (tt TokenType) String() string {
	names := [...]string{
		"OpenParen",
		"CloseParen",
		"Identifier",
		"Number",
		"BinOp",
		"Keyword",
	}

	if tt < OpenParen || tt > Keyword {
		return "Unknown"
	}

	return names[tt]
}

func (kt KeywordType) String() string {
	names := [...]string{
		"define",
		"lambda",
	}
	if kt < Define || kt > Lambda {
		return "Unknown"
	}
	return names[kt]
}

type (
	TokenType   int
	KeywordType int

	Lexer struct {
		bufrd *bufio.Reader
		err   error
		tok   string
		tt    TokenType
	}
)

func NewLexer(rd io.Reader) (lexer *Lexer) {
	return &Lexer{
		bufrd: bufio.NewReader(rd),
		err:   nil,
	}
}

func peekRuneOrPanic(rd *bufio.Reader) (p rune) {
	tmp, err := rd.Peek(1)
	if len(tmp) == 0 {
		panic(io.EOF)
	} else if err != nil {
		panic(err)
	}
	p = rune(tmp[0])
	return
}

func checkBufferForStr(str string, rd *bufio.Reader) (hasIt bool) {
	hasIt = false
	tmp, _ := rd.Peek(len(str))
	if str == string(tmp) {
		hasIt = true
	}
	return
}

func readKeyword(rd *bufio.Reader) (keyword string, err error) {
	err = fmt.Errorf("not a keyword")
	validKeywords := []KeywordType{
		Define,
		Lambda,
	}
	for _, kw := range validKeywords {
		if checkBufferForStr(kw.String(), rd) {
			rd.Discard(len(kw.String()))
			keyword = kw.String()
			err = nil
		}
	}
	return
}

func readIdentifier(rd *bufio.Reader) (id string, err error) {
	firstRt := rangetable.Merge(unicode.Upper, unicode.Lower)
	secondRt := rangetable.Merge(firstRt, unicode.Digit)
	// first rune is a special case
	if !unicode.In(peekRuneOrPanic(rd), firstRt) {
		err = fmt.Errorf("'%v' is not a valid start to an identifier")
		return
	}
	for unicode.In(peekRuneOrPanic(rd), secondRt) {
		var r rune
		r, _, err = rd.ReadRune()
		if err != nil {
			return
		}
		id += string(r)
	}
	return
}

// readNumber tries to read a string that contains a number value or returns an error
func readNumber(rd *bufio.Reader) (num string, err error) {
	err = fmt.Errorf("not a number")
	rt := rangetable.New('0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.')
	foundDecimalPoint := false
	for unicode.In(peekRuneOrPanic(rd), rt) {
		var r rune
		r, _, err = rd.ReadRune()
		if err != nil {
			break
		}

		if r == '.' && !foundDecimalPoint {
			foundDecimalPoint = true
		} else if r == '.' && foundDecimalPoint {
			err = fmt.Errorf("numbers may only have one decimal point")
			break
		}
		num += string(r)
		err = nil
	}
	return
}

func readBinop(rd *bufio.Reader) (sym string, err error) {
	err = fmt.Errorf("not a binary operator")
	rt := rangetable.New('+', '-', '*', '/')
	if unicode.In(peekRuneOrPanic(rd), rt) {
		var r rune
		r, _, err = rd.ReadRune()
		sym = string(r)
	}
	return
}

func (lexer *Lexer) Scan() (result bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if recovered != io.EOF {
				lexer.err = fmt.Errorf("%s", recovered)
			}
			result = false
		}
	}()

	lexer.tok = ""

	// find the first non-space rune
	r := peekRuneOrPanic(lexer.bufrd)
	for unicode.IsSpace(r) {
		lexer.bufrd.Discard(1)
		r = peekRuneOrPanic(lexer.bufrd)
	}
	// get the next token
	if r == '(' {
		lexer.tok += "("
		lexer.tt = OpenParen
		result = true
		lexer.bufrd.Discard(1)
	} else if r == ')' {
		lexer.tok += ")"
		lexer.tt = CloseParen
		result = true
		lexer.bufrd.Discard(1)
	} else if sym, err := readKeyword(lexer.bufrd); err == nil {
		lexer.tok = sym
		lexer.tt = Keyword
		result = true
	} else if id, err := readIdentifier(lexer.bufrd); err == nil {
		lexer.tok = id
		lexer.tt = Identifier
		result = true
	} else if num, err := readNumber(lexer.bufrd); err == nil {
		lexer.tok = num
		lexer.tt = Number
		result = true
	} else if sym, err := readBinop(lexer.bufrd); err == nil {
		lexer.tok = sym
		lexer.tt = BinOp
		result = true
	} else {
		panic("unknown token")
	}

	return result
}

func (lexer *Lexer) Token() string {
	return lexer.tok
}

func (lexer *Lexer) TokenType() TokenType {
	return lexer.tt
}

func (lexer *Lexer) Err() error {
	return lexer.err
}
