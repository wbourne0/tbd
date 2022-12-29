package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"main/inspector"
	"main/lexer"
	"math/bits"
	"strconv"
)

type kind int
type any interface{}

type Parser struct {
	sc *lexer.Scanner

	didErr bool
	token  lexer.Token
	raw    string
	pos    token.Pos
}

func NewParser(src []byte, file *token.File) *Parser {
	scanner, err := lexer.NewScanner(src, file)

	// TODO: handle this in a better way
	if err != nil {
		panic(err)
	}

	parser := &Parser{
		sc: scanner,
	}

	parser.next()

	return parser
}

func (p *Parser) next() {
	p.token, p.raw, p.pos = p.sc.Next()
}

func (p *Parser) currentTokenString() string {
	if p.raw == "" {
		return p.token.String()
	}

	if p.token == lexer.STRING {
		return "\"" + p.raw + "\""
	}

	return p.raw
}

func (p *Parser) positonString() string {
	positionInfo := p.sc.PositionInfo(p.pos)

	return fmt.Sprintf("%d:%d", positionInfo.Line, positionInfo.Column)
}

var (
	unexpectedEOF = errors.New("unexpected EOF (end of file)")
	unexpectedEOL = errors.New("unexpected EOL (end of line)")
)

func (p *Parser) tokenEnd() token.Pos {
	switch p.token {
	case lexer.INT, lexer.FLOAT, lexer.IDENTIFIER:
		return p.pos + token.Pos(len(p.raw))
	case lexer.STRING:
		return p.pos + token.Pos(len(p.raw)) + 2
	}

	return p.pos + token.Pos(len(p.token.String()))
}

type parserError struct {
	error
}

func (p *Parser) err(start token.Pos, str string) parserError {
	return parserError{
		fmt.Errorf("err at %s: %s", p.positonString(), str),
	}
}

func (p *Parser) errf(start token.Pos, str string, values ...interface{}) parserError {
	return parserError{fmt.Errorf("err at %s: %s", p.positonString(), fmt.Sprintf(str, values...))}
}

const expr = `
a
  .b(3 / (10 * 2))
  .c(83 / ~15)
    * d[32 / (x + 5)]
`

// STEPS:
// resolve a.b
// resolve 10 * 2
// divide 3 by the result
// call b with the value
// get property c of the value returned from b
// bitwise invert 15
// divide 83 by the result
// call c with the result
// add 5 to x
// divide 32 by the result
// get the value of d at the index of the result
// multiply the returned alue from c by the value at the index of  d
// func (p *Parser) parseBinaryExpression(prec int) {
// 	switch p.tok.Kind {
// 	case lexer.OPAREN, lexer.OBRACE:

// 	}
// }

func (p *Parser) nodeAt(pos token.Pos) BaseNode {
	return BaseNode{start: pos}
}

func (p *Parser) assertValue(v ValueNode) ValueNode {
	if v, ok := v.(ValueNode); ok {
		return v
	}

	panic(p.err(v.Start(), "expected value"))

}

func (p *Parser) parseIdentifier() (node IdentifierNode) {
	if p.token != lexer.IDENTIFIER {
		panic(fmt.Errorf("not an identifier: %s", p.currentTokenString()))
	}

	node = IdentifierNode{
		BaseNode: p.nodeAt(p.pos),
		Target:   p.raw,
	}
	p.next()

	return
}

func (p *Parser) nodeHere() BaseNode {
	return p.nodeAt(p.pos)
}

func (p *Parser) parseInt() ValueNode {
	var err error
	node := IntegerNode{BaseNode: p.nodeHere(), strlen: len(p.raw)}

	if node.Value, err = strconv.ParseUint(p.raw, 0, 64); err != nil {
		panic(p.errf(p.pos, "unable to parse integer '%s': %s", p.currentTokenString(), err.Error()))
	}
	p.next()

	node.Bits = uint8(bits.Len64(node.Value))

	return node
}

func (p *Parser) parseFloat() ValueNode {
	var err error
	node := FloatNode{BaseNode: p.nodeHere()}

	if node.Value, err = strconv.ParseFloat(p.raw, 32); err != nil {
		if node.Value, err = strconv.ParseFloat(p.raw, 64); err != nil {
			panic(p.errf(p.pos, "unable to parse float '%s': %s", p.currentTokenString(), err.Error()))
		}
		node.isDouble = true
	}

	p.next()

	return node
}

func (p *Parser) parseString() ValueNode {
	node := StringNode{BaseNode: p.nodeHere()}

	if err := json.Unmarshal([]byte(p.raw), &node.Value); err != nil {
		panic(p.errf(p.pos, "error while parsing string: %s", err.Error()))
	}

	p.next()
	return node
}

func (p *Parser) parseValue() ValueNode {
	switch p.token {
	case lexer.IDENTIFIER:
		return p.parseIdentifier()
	case lexer.STRING:
		return p.parseString()
	case lexer.INT:
		return p.parseInt()
	case lexer.FLOAT:
		return p.parseFloat()
	}

	panic(p.errf(p.pos, "unexpected token: '%s'", p.currentTokenString()))
}

func (p *Parser) todo() InvalidNode {
	panic(errors.New("todo"))
}

func (p *Parser) parseType() TypeNode {
	switch p.token {
	case lexer.OBRACK:
		return p.parseSliceOrArrayPrefix()
	case lexer.IDENTIFIER:
		return p.parseIdentifier()
	case lexer.STRUCT, lexer.INTERFACE:
		return p.todo()
	}

	panic(p.errf(p.pos, "not a type: %s", p.currentTokenString()))
}

// parses a slice or array.
func (p *Parser) parseSliceOrArrayPrefix() TypeNode {
	if p.token != lexer.OBRACK {
		panic(fmt.Errorf("expected bracket at %s", p.positonString()))
	}
	start := p.pos
	p.next()

	// This is a slice
	if p.token == lexer.CBRACK {
		p.next()

		return SlicePrefixNode{
			BaseNode: p.nodeAt(start),
			SliceOf:  p.parseType(),
		}
	}

	// Size of the array is determined by the compiler.
	if p.token == lexer.ELLIPSIS {
		p.next()

		if p.token != lexer.CBRACK {
			panic(p.errf(p.pos, "expected closing bracket ']'; received: '%s'", p.currentTokenString()))
		}
		p.next()

		return ArrayPrefixNode{
			BaseNode: p.nodeAt(start),
			Len:      nil,
			ArrayOf:  p.parseType(),
		}
	}

	len := p.parseExpression()

	if p.token != lexer.CBRACK {
		panic(p.errf(p.pos, "expected closing bracket ']'; received '%s'", p.currentTokenString()))
	}

	p.next()

	return ArrayPrefixNode{
		BaseNode: p.nodeAt(start),
		Len:      len,
		ArrayOf:  p.parseType(),
	}
}

func (p *Parser) parseKeyedElements(allowUnkeyed bool) ElementListNode {
	if p.token != lexer.OBRACE {
		panic(fmt.Errorf("unable to parse keyed elements at %s", p.positonString()))
	}
	start := p.pos
	p.next()
	// {}
	if p.token == lexer.CBRACK {
		return ElementsNode{
			BaseNode: p.nodeAt(start),
			Elements: []ElementNode{},
		}
	}

	var (
		elemStart = p.pos
		key       = p.parseExpression()
	)

	// {value,
	if p.token == lexer.COMMA {
		if allowUnkeyed {
			p.next()
			node := p.parseUnkeyedElements()

			if node, ok := node.(ElementsNode); ok {
				node.start = start
				node.Elements = append([]ElementNode{{
					BaseNode: p.nodeAt(elemStart),
					Value:    key,
				}}, node.Elements...)
				return node
			}

			return node
		}

		panic(p.err(elemStart, "expected keyed elements"))
	}

	// {value}
	if p.token == lexer.CBRACE {
		if allowUnkeyed {
			p.next()

			return ElementsNode{
				BaseNode: p.nodeAt(start),
				Elements: []ElementNode{{
					BaseNode: p.nodeAt(elemStart),
					Value:    key,
				}},
			}
		}

		panic(p.err(elemStart, "expected a keyed element"))
	}

	if p.token != lexer.COLON {

		panic(p.errf(p.pos, "expected colon; received: '%s'", p.currentTokenString()))
	}

	p.next()
	elems := []ElementNode{{
		BaseNode: p.nodeAt(elemStart),
		Key:      key,
		Value:    p.parseExpression(),
	}}

	for {
		switch p.token {
		case lexer.CBRACE:
			p.next()
			return ElementsNode{
				BaseNode: p.nodeAt(start),
				Elements: elems,
			}
		case lexer.COMMA:
			p.next()
		default:
			panic(p.errf(p.pos, "unexpected token: '%s'", p.currentTokenString()))
		}

		elemStart = p.pos
		key = p.parseExpression()

		if p.token != lexer.COLON {
			panic(p.errf(p.pos, "expected keyed element"))
		}

		p.next()
		elems = append(elems, ElementNode{
			BaseNode: p.nodeAt(elemStart),
			Key:      key,
			Value:    p.parseExpression(),
		})
	}
}

func (p *Parser) parseUnkeyedElements() ElementListNode {
	node := ElementsNode{}

	for {
		if p.token == lexer.CBRACE {
			return node
		}

		elemStart := p.pos
		val := p.parseExpression()

		switch p.token {
		case lexer.COMMA:
			node.Elements = append(node.Elements, ElementNode{
				BaseNode: p.nodeAt(elemStart),
				Value:    val,
			})
			p.next()
			continue
		case lexer.CBRACE:
			node.Elements = append(node.Elements, ElementNode{
				BaseNode: p.nodeAt(elemStart),
				Value:    val,
			})
			p.next()
			return node
		case lexer.SEMICOLON:
			panic(p.err(p.pos, "expected all elements to be positional"))
		default:
			panic(p.errf(p.pos, "unexpected token: '%s'", p.currentTokenString()))
		}
	}
}

func (p *Parser) parseElementValue() ValueNode {
	if p.token == lexer.OBRACE {
		return p.parseKeyedElements(true)
	}

	return p.parseExpression()
}

func (p *Parser) parseSliceOrArray() ValueNode {
	start := p.pos
	prefix := p.parseSliceOrArrayPrefix()

	if p.token == lexer.OBRACE {
		if prefix, ok := prefix.(SlicePrefixNode); ok {
			return SliceValueNode{
				Prefix:   prefix,
				Elements: p.parseKeyedElements(true),
			}
		}

		if prefix, ok := prefix.(ArrayPrefixNode); ok {
			return ArrayValueNode{
				BaseNode: p.nodeAt(start),
				Prefix:   prefix,
				Elements: p.parseKeyedElements(true),
			}
		}
	}

	panic(p.errf(p.pos, "unexpected token: '%s'", p.currentTokenString()))
}

func (p *Parser) didTerminate() bool {
	switch p.token {
	// TODO: don't exit on eof; this is primarily for testing.
	case lexer.SEMICOLON, lexer.CBRACE, lexer.EOF:
		return true
	}

	if !p.token.IsOperator() && p.sc.DidReadNewline() {
		return true
	}

	return false
}

func (p *Parser) assertTerminator(n StepNode) StepNode {
	if p.didTerminate() {
		return n
	}
	panic(p.err(p.pos, "expected terminator"))
}

func (p *Parser) parseBinaryExpr(prec1 int) ValueNode {
	start := p.pos
	left := p.parseUnaryExpression()

	for {
		op := p.token
		opPrec := op.Precedence()

		if op == lexer.INCR || op == lexer.DECR {
			return p.parseSuffixOperation(left)
		}

		if opPrec < prec1 {
			return left
		}
		p.next()

		right := p.parseBinaryExpr(opPrec + 1)
		left = BinaryOperationNode{
			BaseNode: p.nodeAt(start),
			Left:     left,
			Right:    right,
			Operator: op,
		}
	}
}

func (p *Parser) parseSuffixOperation(left ValueNode) ValueNode {
	node := SuffixUnaryOperationNode{
		UnaryOperationNode: UnaryOperationNode{
			BaseNode: p.nodeAt(left.Start()),
			Operand:  left,
			Operator: p.token,
		},
		end: p.pos + token.Pos(len(p.token.String())),
	}
	p.next()

	return node
}

func (p *Parser) parseUnaryExpression() ValueNode {
	op, start := p.token, p.pos
	switch p.token {
	case lexer.ADD, lexer.SUB, lexer.TILDE, lexer.INCR, lexer.DECR, lexer.NOT:
		p.next()

		return UnaryOperationNode{
			BaseNode: p.nodeAt(start),
			Operator: op,
			Operand:  p.parseUnaryExpression(),
		}
	case lexer.AND:
		p.next()

		// this will break with stuff like &[]int
		if p.token != lexer.IDENTIFIER {
			panic(p.err(p.pos, "unable to get address of non-identifer"))
		}

		return UnaryOperationNode{
			BaseNode: p.nodeAt(start),
			Operator: op,
			Operand:  p.parseUnaryExpression(),
		}
	case lexer.MUL:
		p.next()

		return UnaryOperationNode{
			BaseNode: p.nodeAt(start),
			Operator: op,
			Operand:  p.parseUnaryExpression(),
		}
	}

	return p.parsePrimaryExpression()
}

func (p *Parser) parseParenthesizedExpression() ValueNode {
	if p.token != lexer.OPAREN {
		panic(fmt.Errorf("expected to be called at a parenthesis %s", p.positonString()))
	}

	p.next()
	expr := p.parseExpression()

	if p.token != lexer.CPAREN {
		panic(p.err(p.pos, "expected closing parenthesis"))
	}

	p.next()

	return expr
}

func (p *Parser) parseOperand() ValueNode {
	if p.token.IsValue() {
		return p.parseValue()
	}

	if p.token == lexer.OPAREN {
		return p.parseParenthesizedExpression()
	}

	if p.token == lexer.OBRACK {
		return p.parseSliceOrArray()
	}

	if p.token == lexer.FUNC {
		return p.todo()
	}

	panic(p.errf(p.pos, "unexpected token: '%s'", p.currentTokenString()))
}

func (p *Parser) parseCall(callee ValueNode) ValueNode {
	node := CallNode{
		BaseNode: p.nodeHere(),
		Callee:   callee,
	}
	p.next()

	if p.token == lexer.CPAREN {
		node.end = p.pos
		p.next()
		return node
	}

	for {
		node.Arguments = append(node.Arguments, p.parseExpression())

		switch p.token {
		case lexer.CPAREN:
			node.end = p.pos
			p.next()
			return node
		case lexer.COMMA:
			continue
		default:
			panic(p.errf(p.pos, "unexpected token: '%s'", p.currentTokenString()))
		}
	}
}

func (p *Parser) parsePrimaryExpression() ValueNode {
	node := p.parseOperand()

	for {
		switch p.token {
		case lexer.OPAREN:
			node = p.parseCall(node)
		case lexer.PERIOD:
			p.next()

			if p.token != lexer.IDENTIFIER {
				panic(p.errf(p.pos, "unexpected token: '%s'", p.currentTokenString()))
			}

			node = PropertyAccessNode{
				BaseNode:   p.nodeHere(),
				PropertyOf: node,
				Property:   p.parseIdentifier(),
			}
		case lexer.OBRACK:
			start := p.pos
			p.next()

			//TODO: handle semicolons - ie [1:2]
			key := p.parseExpression()

			if p.token != lexer.CBRACK {
				panic(p.errf(p.pos, "expecteed closing bracket; received '%s'", p.currentTokenString()))
			}

			end := p.pos
			p.next()

			return IndexNode{
				BaseNode: p.nodeAt(start),
				IndexOf:  node,
				Key:      key,
				end:      end,
			}
		default:
			return node
		}
	}
}

func (p *Parser) parseExpression() ValueNode {
	return p.parseBinaryExpr(0)
}

// func (p *Parser) parseKeyword() StepNode {
// 	switch p.token {
// 	case lexer.FUNC:
// 		return p.parseFunc()
// 	default:
// 		return p.errNodef(p.pos, "expected keyword; received: %s", p.currentTokenString())
// 	}
// }

func (p *Parser) parseVariableDeclaration() DeclarationNode {
	start := p.pos
	p.next()

	if p.token != lexer.IDENTIFIER {
		panic(p.errf(p.pos, "expected identifier; received '%s'", p.currentTokenString()))
	}

	v := VariableDeclarationNode{
		BaseNode: p.nodeAt(start),
		name:     p.raw,
	}

	p.next()

	if p.token == lexer.IDENTIFIER {
		v.Type = p.parseIdentifier()
	}

	if p.token != lexer.ASSIGN {
		if v.Type != nil && p.didTerminate() {
			return v
		}

		panic(p.err(start, "expected variable declaration to contain type or initial value"))
	}

	p.next()

	v.Value = p.parseExpression()

	return v
}

func (p *Parser) parseConstantDeclaration() DeclarationNode {
	start := p.pos
	p.next()

	if p.token != lexer.IDENTIFIER {
		panic(p.errf(p.pos, "expected identifier; received '%s'", p.currentTokenString()))
	}

	c := ConstantDeclarationNode{
		BaseNode: p.nodeAt(start),
		name:     p.raw,
	}

	p.next()

	if p.token == lexer.IDENTIFIER {
		c.Type = p.parseIdentifier()
		p.next()
	}

	if p.token != lexer.ASSIGN {
		panic(p.errf(start, "expected constant declaration to assign a value to the constant"))
	}

	p.next()
	c.Value = p.parseExpression()

	return c
}

func (p *Parser) tryParseAssignment(target IdentifierNode) StepNode {
	if p.token.IsAssignmentOperator() {
		value := BinaryOperationNode{
			BaseNode: p.nodeHere(),
			Left:     target,
			Operator: p.token.GetNonAssignmentOperator(),
		}
		p.next()

		value.Right = p.parseExpression()

		return p.assertTerminator(AssignmentNode{
			BaseNode: p.nodeAt(target.Start()),
			Assignee: target,
			Value:    value,
		})
	}

	if p.token == lexer.ASSIGN {
		node := AssignmentNode{
			BaseNode: p.nodeAt(target.Start()),
			Assignee: target,
		}

		p.next()
		node.Value = p.parseExpression()
		node.end = node.Value.End()

		return p.assertTerminator(node)
	}

	if p.token == lexer.DEFINE {
		p.next()

		return p.assertTerminator(VariableDeclarationNode{
			BaseNode: p.nodeAt(target.Start()),
			name:     target.Target,
			Value:    p.parseExpression(),
		})
	}

	return nil
}

func (p *Parser) parseStep() StepNode {
	var (
		target      ValueNode
		didIndirect bool
	)

	switch p.token {
	case lexer.VAR:
		return p.parseVariableDeclaration()
	case lexer.IDENTIFIER:
		target = p.parsePrimaryExpression()
	case lexer.RETURN:
		node := ReturnNode{BaseNode: p.nodeHere()}
		p.next()
		node.Value = p.parseExpression()
		return node
	case lexer.IF:
		return p.parseIf()
	default:
		panic(p.errf(p.pos, "unexpected token: %s", p.currentTokenString()))
	}

	switch target := target.(type) {
	case CallNode:
		if !didIndirect {
			return target
		}
	case BinaryOperationNode, ArrayValueNode, SliceValueNode,
		StringNode, IntegerNode, FloatNode:
		panic(p.errf(target.Start(), "nothing to do"))
	}

	if p.token == lexer.INCR || p.token == lexer.DECR {
		node := SuffixUnaryOperationNode{
			UnaryOperationNode: UnaryOperationNode{
				BaseNode: p.nodeHere(),
				Operand:  target,
				Operator: p.token,
			},
			end: p.pos + 2,
		}
		p.next()
		return p.assertTerminator(node)
	}

	if node := p.tryParseAssignment(target.(IdentifierNode)); node != nil {
		return node
	}

	if p.didTerminate() {
		panic(p.err(p.pos, "expected an action to be performed"))
	} else {
		panic(p.errf(target.Start(), "unexpected token: '%s'", p.currentTokenString()))
	}
}

func (p *Parser) parseBlock() BlockNode {
	if p.token != lexer.OBRACE {
		panic(fmt.Errorf("expected steps to start with brace at %s", p.positonString()))
	}

	node := BlockNode{BaseNode: p.nodeHere()}
	p.next()

	for p.token != lexer.CBRACE {
		node.Steps = append(node.Steps, p.parseStep())
		if p.token == lexer.SEMICOLON {
			p.next()
		}
	}

	node.end = p.pos + 1

	p.next()

	return node
}

func (p *Parser) parseFunctionArguments() ArgumentDeclarationsNode {
	if p.token != lexer.OPAREN {
		panic(p.err(p.pos, "expected function arguments"))
	}

	args := ArgumentDeclarationsNode{
		BaseNode: p.nodeHere(),
	}

	p.next()

	if p.token == lexer.CPAREN {
		p.next()
		return args
	}

	for {
		if p.token != lexer.IDENTIFIER {
			panic(p.err(p.pos, "expected identifier"))
		}

		arg := ArgumentDeclarationNode{
			BaseNode: p.nodeHere(),
			name:     p.raw,
		}

		p.next()

		if p.token == lexer.COMMA {
			p.next()
			args.Arguments = append(args.Arguments, arg)
			continue
		}

		// TODO: parse type.

		if p.token != lexer.IDENTIFIER {
			panic(p.err(p.pos, "expected identifier"))
		}

		arg.Type = p.parseIdentifier()

		args.Arguments = append(args.Arguments, arg)

		switch p.token {
		case lexer.CPAREN:
			p.next()
			return args
		case lexer.COMMA:
			p.next()
			continue
		default:
			panic(p.errf(p.pos, "expected comma or closing parenthesis; received %s", p.currentTokenString()))
		}
	}
}

func (p *Parser) parseFunctionArgumentsAndBlock(start token.Pos) FunctionNode {
	node := FunctionNode{
		BaseNode: p.nodeAt(start),
	}

	if p.token == lexer.OPAREN {
		node.Arguments = p.parseFunctionArguments()
	} else {
		panic(p.errf(p.pos, "expected start of function arguments; received '%s'", p.currentTokenString()))
	}

	if p.token == lexer.IDENTIFIER {
		node.Returns = p.parseIdentifier()
	}

	if p.token == lexer.OBRACE {
		node.Block = p.parseBlock()
	} else {
		panic(p.errf(p.pos, "expected start of function block; received '%s'", p.currentTokenString()))
	}

	return node

}

// parses the part of a function's declaration from `func` to `(`
func (p *Parser) parseFunctionHead() TopLevelNode {
	var (
		baseNode      = p.nodeHere()
		isPtrReceiver = false
	)

	if p.token != lexer.FUNC {
		panic(fmt.Errorf("invalid usage of parseTopLevelFunc at %s", p.positonString()))
	}
	p.next()

	if p.token == lexer.MUL {
		isPtrReceiver = true
		p.next()
	}

	if p.token != lexer.IDENTIFIER {
		panic(p.errf(p.pos, "expected identifier; received '%s'", p.currentTokenString()))
	}

	ident := p.raw

	p.next()

	if p.token != lexer.PERIOD {
		if isPtrReceiver {
			panic(p.errf(p.pos, "expected method name in the form of '.method'; received '%s'", p.currentTokenString()))
		}

		return ModuleFunctionDeclarationNode{
			BaseNode: baseNode,
			name:     ident,
		}
	}

	p.next()

	if p.token != lexer.IDENTIFIER {
		panic(p.errf(p.pos, "expected method name; received '%s'", p.currentTokenString()))
	}
	node := MethodDeclarationNode{
		BaseNode:      baseNode,
		name:          p.raw,
		MethodOf:      ident,
		IsPtrReceiver: isPtrReceiver,
	}
	p.next()
	return node
}

func (p *Parser) parseTopLevelFunc() TopLevelNode {
	start := p.pos

	fn := p.parseFunctionHead()

	if p.token == lexer.OPAREN {
		body := p.parseFunctionArgumentsAndBlock(start)
		switch head := fn.(type) {
		case ModuleFunctionDeclarationNode:
			head.Body = body
			fn = head
		case MethodDeclarationNode:
			head.Body = body
			fn = head
		}
	}

	return fn
}

func (p *Parser) ParseModule() (mod ModuleNode) {

	for p.token != lexer.EOF {
		var (
			node     TopLevelNode
			isPublic bool
			start    = p.pos
		)

		if isPublic = p.token == lexer.PUBLIC; isPublic {
			p.next()
			if p.token == lexer.PUBLIC {
				panic(p.err(p.pos, "unexpected token"))
			}
		}

		switch p.token {
		case lexer.FUNC:
			node = p.parseTopLevelFunc()
		case lexer.VAR:
			node = p.parseVariableDeclaration()
		case lexer.CONST:
			node = p.parseConstantDeclaration()
		case lexer.SEMICOLON:
			p.next()
			continue
		default:
			panic(p.errf(p.pos, "unexpected identifier %s", inspector.Inspect(p.token)))
		}

		if isPublic {
			node = PublicNode{
				BaseNode: p.nodeAt(start),
				Node:     node,
			}
		}

		mod.Nodes = append(mod.Nodes, node)
	}

	return
}

func (p *Parser) parseIfOnly() IfNode {
	node := IfNode{
		BaseNode: p.nodeHere(),
	}

	p.next()

	node.Condition = p.parseExpression()

	if p.token != lexer.OBRACE {
		panic(p.err(p.pos, "expected start of block"))
	}

	node.Then = p.parseBlock()

	return node
}

// TODO: go's initialization in `if` statements is super handy,
func (p *Parser) parseIf() IfNode {
	node := p.parseIfOnly()

	for p.token == lexer.ELSE {
		p.next()
		switch p.token {
		case lexer.OBRACE:
			bl := p.parseBlock()
			node.Else = &bl
		case lexer.IF:
			node.ElseIf = append(node.ElseIf, p.parseIfOnly())
		default:
			panic(p.errf(p.pos, "unexpected token: %s", p.token.String()))
		}
	}

	return node
}
