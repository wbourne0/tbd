package generator

import (
	"errors"
	"fmt"
	"main/inspector"
)

type Generic struct {
	kind Kind
}

func (g Generic) Kind() Kind {
	return g.kind
}

func (g Generic) Zero() any {
	switch g.kind {
	case KindInt8:
		return int8(0)
	case KindInt16:
		return int16(0)
	case KindInt32:
		return int32(0)
	case KindInt64:
		return int64(0)
	case KindUint8:
		return uint8(0)
	case KindUint16:
		return uint16(0)
	case KindUint32:
		return uint32(0)
	case KindUint64:
		return uint64(0)
	case KindFloat32:
		return float32(0)
	case KindFloat64:
		return float64(0)
	case KindString:
		return ""
	case KindBool:
		return false
	}

	panic(errors.New("not a Generic"))
}

func (g Generic) Name() string {
	switch g.kind {
	case KindInt:
		return "int"
	case KindInt8:
		return "int8"
	case KindInt16:
		return "int16"
	case KindInt32:
		return "int32"
	case KindInt64:
		return "int64"
	case KindUint:
		return "uint64"
	case KindUint8:
		return "uint8"
	case KindUint16:
		return "uint16"
	case KindUint32:
		return "uint32"
	case KindUint64:
		return "uint64"
	case KindFloat32:
		return "float32"
	case KindFloat64:
		return "float64"
	case KindBool:
		return "bool"
	}

	panic(errors.New("not a Generic"))
}

// TODO: later it might be handy if we can assign a value of a lower byte count
// to a value with a higher one - eg assign an int8 to an int16
func (g *Generic) AssignableTo(other Type) bool {
	if other == nil {
		return true
	}
	v, _ := other.(*Generic)
	return g == v
}

func (k Kind) InspectCustom() inspector.InspectString {
	switch k {
	case KindInt:
		return "int"
	case KindInt8:
		return "int8"
	case KindInt16:
		return "int16"
	case KindInt32:
		return "int32"
	case KindInt64:
		return "int64"
	case KindUint:
		return "uint"
	case KindUint8:
		return "uint8"
	case KindUint16:
		return "uint16"
	case KindUint32:
		return "uint32"
	case KindUint64:
		return "uint64"
	case KindString:
		return "string"
	case KindFloat32:
		return "float32"
	case KindFloat64:
		return "float64"
	default:
		return inspector.InspectString(fmt.Sprintf("kind: %d", k))
	}
}

func (g Generic) InspectCustom() inspector.InspectString {
	return g.kind.InspectCustom()
}

var (
	genericInt     = &Generic{kind: KindInt}
	genericInt8    = &Generic{kind: KindInt8}
	genericInt16   = &Generic{kind: KindInt16}
	genericInt32   = &Generic{kind: KindInt32}
	genericInt64   = &Generic{kind: KindInt64}
	genericUint    = &Generic{kind: KindUint}
	genericUint8   = &Generic{kind: KindUint8}
	genericUint16  = &Generic{kind: KindUint16}
	genericUint32  = &Generic{kind: KindUint32}
	genericUint64  = &Generic{kind: KindUint64}
	genericString  = &Generic{kind: KindString}
	genericFloat32 = &Generic{kind: KindFloat32}
	genericFloat64 = &Generic{kind: KindFloat64}

	genericBool = &Generic{kind: KindBool}

	internalInvalid = &Generic{kind: KindBool}
)

var Generics = map[string]*Generic{
	"int":   genericInt,
	"int8":  genericInt8,
	"int16": genericInt16,
	"int32": genericInt32,
	"int64": genericInt64,

	"uint":    genericUint,
	"uint8":   genericUint8,
	"uint16":  genericUint16,
	"uint32":  genericUint32,
	"uint64":  genericUint64,
	"string":  genericString,
	"float32": genericFloat32,
	"float64": genericFloat64,

	"bool": genericBool,
}
