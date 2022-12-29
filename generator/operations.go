package generator

import (
	"errors"
	"fmt"
	"main/lexer"
	"main/parser"
	"math/bits"
	"reflect"
	"unsafe"
)

func constantOf(v any, typ Type) ConstantValue {
	return ConstantValue{
		value: v,
		typ:   typ,
	}
}

// TODO: binary ops need to be improved.

func (s *Scope) resolveBinaryUintOperation(left, right uint64, operator lexer.Token, typ Type) Value {
	var val uint64
	switch operator {
	case lexer.ADD:
		val = left + right
	case lexer.SUB:
		val = left - right
	case lexer.MUL:
		val = left * right
	case lexer.DIV:
		val = left / right
	case lexer.MOD:
		val = left % right
	case lexer.AND:
		val = left & right
	case lexer.OR:
		val = left | right
	case lexer.XOR:
		val = left ^ right
	case lexer.AND_NOT:
		val = left &^ right
	case lexer.LEFT_SHIFT:
		val = left << right
	case lexer.RIGHT_SHIFT:
		val = left >> right
	case lexer.EQL:
		return constantOf(left == right, genericBool)
	case lexer.NOT_EQL:
		return constantOf(left != right, genericBool)
	case lexer.GREATER:
		return constantOf(left > right, genericBool)
	case lexer.LESS:
		return constantOf(left < right, genericBool)
	case lexer.GREATER_EQL:
		return constantOf(left >= right, genericBool)
	case lexer.LESS_EQL:
		return constantOf(left <= right, genericBool)
	default:
		panic(errors.New("not an operator"))
	}

	if t, ok := typ.(untypedInt); ok {
		t.bits = uint8(bits.Len64(uint64(-val)))
	}

	return constantOf(val, typ)
}

func (s *Scope) resolveBinaryIntOperation(left, right int64, operator lexer.Token, typ Type) Value {
	var val int64
	switch operator {
	case lexer.ADD:
		val = left + right
	case lexer.SUB:
		val = left - right
	case lexer.MUL:
		val = left * right
	case lexer.DIV:
		val = left / right
	case lexer.MOD:
		val = left % right
	case lexer.AND:
		val = left & right
	case lexer.OR:
		val = left | right
	case lexer.XOR:
		val = left ^ right
	case lexer.AND_NOT:
		val = left &^ right
	case lexer.LEFT_SHIFT:
		val = left << right
	case lexer.RIGHT_SHIFT:
		val = left >> right
	case lexer.EQL:
		return constantOf(left == right, genericBool)
	case lexer.NOT_EQL:
		return constantOf(left != right, genericBool)
	case lexer.GREATER:
		return constantOf(left > right, genericBool)
	case lexer.LESS:
		return constantOf(left < right, genericBool)
	case lexer.GREATER_EQL:
		return constantOf(left >= right, genericBool)
	case lexer.LESS_EQL:
		return constantOf(left <= right, genericBool)
	default:
		panic(errors.New("not an operator"))
	}

	if t, ok := typ.(untypedInt); ok {
		t.negative = val < 0
		if t.negative {
			t.bits = uint8(bits.Len64(uint64(-val)))
		} else {
			t.bits = uint8(bits.Len64(uint64(val)))
		}
	}

	return constantOf(val, typ)
}

func (s *Scope) resolveBinaryOperation(left, right ConstantValue, op parser.BinaryOperationNode) Value {
	switch l, r := reflect.ValueOf(left.value), reflect.ValueOf(right.value); l.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return s.resolveBinaryUintOperation(l.Uint(), r.Uint(), op.Operator, left.Type())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return s.resolveBinaryIntOperation(l.Int(), r.Int(), op.Operator, left.Type())
	case reflect.String:
		if op.Operator != lexer.ADD {
			// filled in by caller
			s.error(op, "invalid operation for type string: %s", op.Operator.String())
			return nil
		}

		return ConstantValue{
			typ:   left.Type(),
			value: l.String() + r.String(),
		}
	default:
		panic(fmt.Errorf("invalid operation: %s %s %s", left.Type().Name(), op.Operator.String(), right.Type().Name()))
	}
}

func (s *Scope) resolveUnaryOperation(operand ConstantValue, node parser.UnaryOperationNode) Value {
	if node.Operator == lexer.NOT {
		return ConstantValue{
			value: reflect.ValueOf(operand.value).IsZero(),
			typ:   genericBool,
		}
	}

	if operand.typ.Kind() == KindString {
		if node.Operator != lexer.ADD {
			s.error(node, "invalid operation for type string: %s", node.Operator.String())

			return nil
		}

		return operand
	}

	switch node.Operator {
	case lexer.ADD:
		return operand
	case lexer.SUB:
		switch val := reflect.ValueOf(operand.value); operand.typ.Kind() {
		case KindInt8, KindInt16, KindInt32, KindInt64:
			operand.value = -val.Int()
			return operand
		default:
			s.error(node, "invalid operation for type %s: %s", operand.typ.Name(), node.Operator.String())
			return nil
		}
	case lexer.TILDE:

		switch val := reflect.ValueOf(operand.value); val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			n := new(uint64)

			*(*int64)(unsafe.Pointer(n)) = val.Int()

			return s.resolveBinaryUintOperation(^*n, 0, lexer.OR, operand.typ)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

			return s.resolveBinaryUintOperation(^val.Uint(), 0, lexer.OR, operand.typ)

		default:
			s.error(node, "invalid operation for type %s: %s", operand.typ.Name(), node.Operator.String())
			return nil
		}
	default:
		s.error(node, "invalid operation for type %s: %s", operand.typ.Name(), node.Operator.String())
		return nil
	}
}

type untypedInt struct {
	bits     uint8
	negative bool
}

func (u untypedInt) Zero() any {
	return int64(0)
}

func (u untypedInt) Name() string {
	return "untyped int"
}

func (u untypedInt) Kind() Kind {
	return kindUntypedInt
}

func (u untypedInt) AssignableTo(target Type) bool {
	if target == nil {
		return true
	}

	switch bits := u.bits; target.Kind() {
	case KindUint8:
		return !u.negative && bits <= 8
	case KindUint16:
		return !u.negative && bits <= 16
	case KindUint32, KindUint:
		return !u.negative && bits <= 32
	case KindUint64:
		return !u.negative && bits <= 64
	case KindInt8:
		return bits < 8
	case KindInt16:
		return bits < 16
	case KindInt32, KindInt:
		return bits < 32
	case KindInt64:
		return bits < 64
	case kindUntypedInt:
		return true
	default:
		return false
	}
}
