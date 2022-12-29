package lexer

import (
	"errors"
	"fmt"
	"go/token"
)

// type Position token.Position

type PositionError struct {
	token.Position
	error
}

// func (p Position) WrapError(err error) PosError {
// 	return PosError{error: err, Position: p}
// }

// REFERENCE: https://github.com/golang/go/blob/7c5d7a4caffdb72ce252fb465ff4f7fd62a46c8a/src/go/scanner/scanner.go

// const (
// 	bom = 0xFEFF
// 	eof = -1
// )

type Scanner struct {
	src    []byte
	offset int
	char   byte

	eof            bool
	file           *token.File
	didReadNewline bool
}

func (s *Scanner) DidReadNewline() bool {
	return s.didReadNewline
}

var t token.File

func (s *Scanner) PositionInfo(p token.Pos) token.Position {
	return s.file.Position(p)
}

func NewScanner(src []byte, file *token.File) (scanner *Scanner, err error) {
	scanner = &Scanner{file: file, char: src[0]}
	scanner.src = src

	return
}

func (s *Scanner) error(offset int, msg string) PositionError {
	return PositionError{
		error:    errors.New(msg),
		Position: s.file.Position(s.file.Pos(offset)),
	}
}

func (s *Scanner) errorf(offset int, msg string, args ...interface{}) PositionError {
	return PositionError{
		error:    fmt.Errorf(msg, args...),
		Position: s.file.Position(s.file.Pos(offset)),
	}
}

func (s *Scanner) next() {

	if s.eof {
		return
	}

	s.offset++

	if s.char == '\n' {
		s.file.AddLine(s.offset)
		s.didReadNewline = true
	}

	if s.offset >= len(s.src) {

		s.eof = true
		s.char = 0
		return
	}

	s.char = s.src[s.offset]

}

func isDecimal(char byte) bool {
	return char >= '0' && char <= '9'
}

func (s *Scanner) peek() byte {
	if s.offset+1 >= len(s.src) {
		return 0
	}

	return s.src[s.offset+1]
}

func (s *Scanner) readIdentifier() string {
	start := s.offset

	for isLetter(s.char) || isDecimal(s.char) || s.char == '_' {
		s.next()
	}

	return string(s.src[start:s.offset])
}

// reads a string in the form of "{string}"
// (s.char should equal '"' prior to reading this)
func (s *Scanner) readInlineString() (str string) {
	start := s.offset

	for {
		s.next()

		if s.char == '"' {
			s.next()
			return string(s.src[start:s.offset])
		}

		if s.char == '\\' {
			s.next()
		}

		// todo: avoid panic and actually error properly...
		if s.char == '\n' {
			panic("not a string")
		}
	}
}

func isBinary(char rune) bool {
	return char == '0' || char == '1'
}

func isHexadecimal(char byte) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char <= 'F' && char >= 'A')
}

func isOctal(char byte) bool {
	return '0' <= char && char <= '7'
}

func (s *Scanner) readFor(check func(byte) bool) {
	for check(s.char) {
		s.next()
	}
}

func (s *Scanner) readNumber() (kind Token) {
	var (
		isExponential bool   // whether the number contains `e` or not.
		canHaveDot    = true // whether the number can have a dot or not.
	)

	kind = INT

	if s.char == '.' {
		s.next()
		s.readFor(isDecimal)
		return INT
	}

	if s.char == '0' {
		s.next()

		switch s.char {
		case 'x':
			s.next()
			s.readFor(isHexadecimal)
			return INT
		case 'o':
			s.next()
			s.readFor(isOctal)
			return INT
		case 'd':
			s.next()
			s.readFor(isDecimal)
			return INT
		}

		if !isDecimal(s.char) {
			return
		}
	}

	for {
		s.next()

		switch {
		case isDecimal(s.char), s.char == '_':
		case s.char == '.':
			if !canHaveDot {
				return
			}

			kind = FLOAT
			canHaveDot = false
		case s.char == 'e':
			if isExponential {
				return
			}

			isExponential = true
			canHaveDot = false
		default:
			return
		}

	}
}

// todo: operator shit, reference: https://github.com/golang/go/blob/1d004fa2015d128acf6302fc74b95f6a36c35680/src/go/scanner/scax#L762

func (s *Scanner) skipWhitespace() {
	for s.char == ' ' || s.char == '\t' || s.char == '\r' || s.char == '\n' {
		s.next()
	}
}

func isLetter(char byte) bool {
	// The lowercase variants of letters are exactly 32 higher than the same variant in
	/// uppercase.
	// &^ will convert those to lowercase.
	char &^= 32

	return 'A' <= char && char <= 'Z'
}

// Partially sourced from https://github.com/golang/go/blob/7c5d7a4caffdb72ce252fb465ff4f7fd62a46c8a/src/go/scanner/scanner.go#L829

func (s *Scanner) Next() (tok Token, raw string, pos token.Pos) {
	s.didReadNewline = false
	s.skipWhitespace()

	start := s.offset
	if s.eof {
		tok = EOF
		return
	}
	pos = s.file.Pos(start)

	switch char := s.char; {
	case isLetter(char), char == '_':
		ident := s.readIdentifier()

		if len(ident) > 1 {
			tok = lookup(ident)
			if !tok.IsKeyword() {
				raw = ident
			}
		} else {
			tok = IDENTIFIER
			raw = ident
		}

	case isDecimal(char), char == '.' && isDecimal(s.peek()):
		tok = s.readNumber()
		raw = string(s.src[start:s.offset])
	case char == '"':
		tok = STRING
		raw = s.readInlineString()
		pos = token.Pos(start)
	default:
		if s.char == '.' && s.peek() == '.' {
			s.next()

			if s.peek() == '.' {
				s.next()
				s.next()

				tok = ELLIPSIS
				return
			}

			tok = PERIOD
			return
		}
		didReadNewline := s.didReadNewline
		if op, didReadNext := lookupOperator(s.char, s.peek()); op != INVALID {
			if didReadNext {
				s.next()
			}

			s.next()
			if op.hasAssignmentOperator() && s.char == '=' {
				op = op.getAssignmentOperator()
				s.next()
			}

			tok = op

			if tok == SEMICOLON && didReadNewline {
				// honestly what goes through someone's head when they do this shit
				panic("terrible syntax decisions")
			}

			return
		}
	}

	return
}
