package generator

import (
	"fmt"
	"main/lexer"
	"main/parser"
	"reflect"
)

type Typed interface {
	Type() Type
}

type BinaryOperation struct {
	Left     Typed
	Right    Typed
	Operator lexer.Token
}

type UnaryOperation struct {
	Operand  Typed
	Operator lexer.Token
}

// ++ or --
type SuffixUnaryOperation struct {
	Target   Writeable
	Operator lexer.Token
}

func (s SuffixUnaryOperation) Type() Type {
	return s.Target.Type()
}

func (u UnaryOperation) Type() Type {
	return u.Operand.Type()
}

func (b BinaryOperation) Type() Type {
	if b.Left == nil {
		return &Generic{kind: invalid}
	}
	return b.Left.Type()
}

func (s *Scope) evaluateBinaryExpression(node parser.BinaryOperationNode) Typed {
	left := s.preEvaluate(node.Left)

	if left == nil {
		return nil
	}

	right := s.preEvaluate(node.Right)

	if right == nil {
		return nil
	}

	if !right.Type().AssignableTo(left.Type()) {
		s.error(node, "type mismatch: unable to resolve %s %s %s", left.Type().Name(), node.Operator.String(), right.Type().Name())
		return nil
	}

	if isConstant(left) && isConstant(right) {
		left, right := left.(ConstantValue), right.(ConstantValue)
		if node.Operator == lexer.BOOLEAN_AND {
			if reflect.ValueOf(left.Value()).IsZero() {
				return left
			}
			return right
		}

		if node.Operator == lexer.BOOLEAN_OR {
			if reflect.ValueOf(left.Value()).IsZero() {
				return right
			}
			return left
		}

		return s.resolveBinaryOperation(left, right, node)
	}

	return BinaryOperation{
		Left:     left,
		Right:    right,
		Operator: node.Operator,
	}
}

func (s *Scope) evaluateUnaryExpression(node parser.UnaryOperationNode) Typed {
	operand := s.preEvaluate(node.Operand)

	if operand == nil {
		return nil
	}

	if operand, ok := operand.(ConstantValue); ok {
		return s.resolveUnaryOperation(operand, node)
	}

	return UnaryOperation{
		Operand:  operand,
		Operator: node.Operator,
	}
}

func (s *Scope) evaluateUnarySuffixExpression(node parser.SuffixUnaryOperationNode) Typed {
	ident, ok := node.Operand.(parser.IdentifierNode)

	if !ok {
		s.error(node, "expected a name or identifier")
		return nil
	}

	val := s.lookupTyped(ident)

	if val == nil {
		return nil
	}

	if !val.Type().Kind().isNumeric() {
		s.error(node, "invalid operation for type %s: %s", val.Type().Name(), node.Operator)
		return nil
	}

	writeable, ok := val.(Writeable)

	if !ok {
		s.error(node, "not writeable: %s", ident.Target)
		return nil
	}

	return SuffixUnaryOperation{
		Target:   writeable,
		Operator: node.Operator,
	}
}

func (s *Scope) preEvaluate(val parser.ValueNode) Typed {
	switch node := val.(type) {
	case parser.BinaryOperationNode:

		return s.evaluateBinaryExpression(node)
	case parser.UnaryOperationNode:

		return s.evaluateUnaryExpression(node)
	case parser.SuffixUnaryOperationNode:
		return s.evaluateUnarySuffixExpression(node)
	case parser.IdentifierNode:
		return s.lookupTyped(node)
	case parser.StringNode:
		return ConstantValue{
			value: node.Value,
			typ:   genericString,
		}
	case parser.IntegerNode:
		return ConstantValue{
			value: node.Value,
			typ: untypedInt{
				bits: node.Bits,
			},
		}
	case parser.FloatNode:
		// todo:  probably worth making an untyped float
		return ConstantValue{
			value: node.Value,
			typ:   genericFloat64,
		}
	case parser.CallNode:
		return s.handleCall(node)
	default:
		// TODO: Slice, struct, index, etc.
		panic(fmt.Errorf("not implemented: evaluate %s", reflect.TypeOf(val).Name()))
	}
}
