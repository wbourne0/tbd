package lexer

import (
	"unicode"
)

// Parts of this file are sourced or directly copied from google's golang parser.
// https://github.com/golang/go/blob/3d7cb23e3d5e7880d582f1b0300064bd1138f3ee/src/go/token/token.go

// A lexical token
type Token int

// TODO: These descriptors would be much neater in a yaml file.

const (
	INVALID Token = iota
	EOF
	COMMENT

	values_begin
	// Identifiers and basic type literals
	// (these tokens stand for classes of literals)
	IDENTIFIER // something
	INT        // 543
	FLOAT      // 6543.23
	// CHAR   // 'a'  // TODO
	STRING // "blah"
	literal_end

	values_end

	// Operators and delimiters
	operators_begin
	ADD // +
	SUB // -
	MUL // *
	DIV // /
	MOD // %

	AND         // &
	OR          // |
	XOR         // ^
	LEFT_SHIFT  // <<
	RIGHT_SHIFT // >>
	AND_NOT     // &^

	assignment_operators_begin

	ADD_ASSIGN // +=
	SUB_ASSIGN // -=
	MUL_ASSIGN // *=
	QUO_ASSIGN // /=
	REM_ASSIGN // %=

	AND_ASSIGN         // &=
	OR_ASSIGN          // |=
	XOR_ASSIGN         // ^=
	LEFT_SHIFT_ASSIGN  // <<=
	RIGHT_SHIFT_ASSIGN // >>=
	AND_NOT_ASSIGN     // &^=

	assignment_operators_end

	BOOLEAN_AND // &&
	BOOLEAN_OR  // ||
	INCR        // ++
	DECR        // --

	EQL     // ==
	LESS    // <
	GREATER // >
	ASSIGN  // =
	NOT     // !

	NOT_EQL     // !=
	LESS_EQL    // <=
	GREATER_EQL // >=
	DEFINE      // :=
	TILDE       // ~
	ELLIPSIS    // ...

	OPAREN // (
	OBRACK // [
	OBRACE // {
	COMMA  // ,
	PERIOD // .

	CPAREN    // )
	CBRACK    // ]
	CBRACE    // }
	SEMICOLON // ;
	COLON     // :

	operators_end

	keyword_begin
	// Keywords
	CONST
	PUBLIC
	NIL
	// CASE
	// BREAK
	// CONTINUE

	// DEFAULT
	// DEFER
	ELSE
	// FALLTHROUGH
	// FOR

	FUNC
	// GO
	// GOTO
	IF
	IMPORT

	INTERFACE
	// MAP
	// PACKAGE
	// RANGE
	RETURN

	// SELECT
	STRUCT
	// SWITCH
	// TYPE
	VAR
	keyword_end
)

var (
	tokens = [...]string{

		EOF:     "EOF",
		COMMENT: "COMMENT",

		IDENTIFIER: "IDENTIFIER",
		INT:        "INT",
		FLOAT:      "FLOAT",
		STRING:     "STRING",

		ADD: "+",
		SUB: "-",
		MUL: "*",
		DIV: "/",
		MOD: "%",

		AND:         "&",
		OR:          "|",
		XOR:         "^",
		LEFT_SHIFT:  "<<",
		RIGHT_SHIFT: ">>",
		AND_NOT:     "&^",

		ADD_ASSIGN: "+=",
		SUB_ASSIGN: "-=",
		MUL_ASSIGN: "*=",
		QUO_ASSIGN: "/=",
		REM_ASSIGN: "%=",

		AND_ASSIGN:         "&=",
		OR_ASSIGN:          "|=",
		XOR_ASSIGN:         "^=",
		LEFT_SHIFT_ASSIGN:  "<<=",
		RIGHT_SHIFT_ASSIGN: ">>=",
		AND_NOT_ASSIGN:     "&^=",

		BOOLEAN_AND: "&&",
		BOOLEAN_OR:  "||",
		INCR:        "++",
		DECR:        "--",

		EQL:     "==",
		LESS:    "<",
		GREATER: ">",
		ASSIGN:  "=",
		NOT:     "!",

		NOT_EQL:     "!=",
		LESS_EQL:    "<=",
		GREATER_EQL: ">=",
		DEFINE:      ":=",
		TILDE:       "~",
		ELLIPSIS:    "...",

		OPAREN: "(",
		OBRACK: "[",
		OBRACE: "{",
		COMMA:  ",",
		PERIOD: ".",

		CPAREN:    ")",
		CBRACK:    "]",
		CBRACE:    "}",
		SEMICOLON: ";",
		COLON:     ":",

		CONST:  "const",
		PUBLIC: "public",
		NIL:    "nil",

		ELSE: "else",

		FUNC:   "func",
		IF:     "if",
		IMPORT: "import",

		INTERFACE: "interface",
		RETURN:    "return",

		STRUCT: "struct",
		VAR:    "var",
	}

	keywords            = make(map[string]Token)
	operatorLookupTable = [...]Token{
		'+':                ADD,
		'-':                SUB,
		'*':                MUL,
		'/':                DIV,
		'%':                MOD,
		'&':                AND,
		'|':                OR,
		'^':                XOR,
		('<' | ('<' << 7)): LEFT_SHIFT,
		('>' | ('>' << 7)): RIGHT_SHIFT,
		('&' | ('^' << 7)): AND_NOT,
		('&' | ('&' << 7)): BOOLEAN_AND,
		('|' | ('|' << 7)): BOOLEAN_OR,
		('+' | ('+' << 7)): INCR,
		('-' | ('-' << 7)): DECR,
		(':' | ('=' << 7)): DEFINE,
		('!' | ('=' << 7)): NOT_EQL,
		('=' | ('=' << 7)): EQL,

		('<' | ('=' << 7)): LESS_EQL,
		('>' | ('=' << 7)): GREATER_EQL,
		'<':                LESS,
		'>':                GREATER,
		'=':                ASSIGN,
		'!':                NOT,
		'(':                OPAREN,
		'[':                OBRACK,
		'{':                OBRACE,
		',':                COMMA,
		'.':                PERIOD,
		')':                CPAREN,
		']':                CBRACK,
		'}':                CBRACE,
		':':                COLON,
		';':                SEMICOLON,
		'~':                TILDE,
	}
)

func init() {
	keywords = make(map[string]Token)

	for i := keyword_begin + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

func lookupOperator(char byte, next byte) (op Token, didReadNext bool) {
	// check if the char is ascii
	if char&0x80 == 0 {
		index := (uint16(next) << 7) | uint16(char)
		if index < uint16(len(operatorLookupTable)) {

			if maybeToken := operatorLookupTable[index]; maybeToken != INVALID {
				return maybeToken, true
			}
		}
	}

	return operatorLookupTable[char], false
}

func lookup(str string) Token {
	if tok, is_keyword := keywords[str]; is_keyword {
		return tok
	}

	return IDENTIFIER
}

func (t Token) IsValue() bool { return values_begin < t && t < values_end }

func (t Token) IsOperator() bool { return operators_begin < t && t < operators_end }

func (t Token) IsAssignmentOperator() bool {
	return assignment_operators_begin < t && t < assignment_operators_end
}

// Gets t as a non-assignment operator.  IE *= -> *
func (t Token) GetNonAssignmentOperator() Token {
	return t - (assignment_operators_begin - operators_begin)
}

func (t Token) InspectCustom() interface{} {
	if int(t) < len(tokens) && tokens[t] != "" {
		return tokens[t]
	}

	return int(t)
}

func (t Token) hasAssignmentOperator() bool {
	return operators_begin < t && t < assignment_operators_begin
}

func (t Token) getAssignmentOperator() Token {
	return t + (assignment_operators_begin - operators_begin)
}

func (t Token) IsKeyword() bool { return keyword_begin < t && t < keyword_end }

func isKeyword(str string) bool {
	_, ok := keywords[str]
	return ok
}

func isIdentifier(name string) bool {
	for i, c := range name {
		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
			return false
		}
	}
	return name != "" && !isKeyword(name)
}

// For now, this function is copied from https://github.com/golang/go/blob/f6647f2e3bc0b803a67c97a7d5d8733cefbd5d5b/src/go/token/token.go#L270
// (with some minor changes for variable names.)
// TODO: Optimize via a table containing the metadata of operators.

// Precedence returns the operator precedence of the binary
// operator op. If op is not a binary operator, the result
// is LowestPrecedence.
func (t Token) Precedence() int {
	switch t {
	case BOOLEAN_OR:
		return 1
	case BOOLEAN_AND:
		return 2
	case EQL, NOT_EQL, LESS, LESS_EQL, GREATER, GREATER_EQL:
		return 3
	case ADD, SUB, OR, XOR:
		return 4
	case MUL, DIV, MOD, LEFT_SHIFT, RIGHT_SHIFT, AND, AND_NOT:
		return 5
	case INCR, DECR:
		return 6
	case OPAREN:
		return 7
	}

	return -1
}

func (t Token) String() string {
	return tokens[t]
}
