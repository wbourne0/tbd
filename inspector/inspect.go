package inspector

import (
	"fmt"
	"reflect"
	"strings"
	"time"
	"unsafe"
)

func joinRows(suffix string, rows []string, depth int) string {
	var str strings.Builder
	if len(rows) == 0 {
		return suffix
	}

	str.WriteByte('\n')

	for _, row := range rows {
		for i := 0; i < depth+1; i++ {
			str.WriteString("  ")
		}

		str.WriteString(row + "\n")
	}

	for i := 0; i < depth; i++ {
		str.WriteString("  ")
	}
	str.WriteString(suffix)

	return str.String()
}

var reflectValueType = reflect.TypeOf(reflect.Value{})
var typTime = reflect.TypeOf(time.Time{})

// clears the bit which marks a reflect value as private.
func clearReflectPrivateBit(v *reflect.Value) {
	type ptr unsafe.Pointer
	*(*uintptr)(ptr(uintptr(ptr(v)) + reflectValueType.Field(2).Offset)) &^= 96
}

// func makeAddressable(v *reflect.Value)

func addressOf(v reflect.Value, field int) uintptr {
	clearReflectPrivateBit(&v)

	return v.Field(field).UnsafeAddr()
}

func interfaceOf(v reflect.Value) any {
	clearReflectPrivateBit(&v)

	return v.Interface()
}

func format(msg string, colors bool, codes string) string {
	var str strings.Builder

	if !colors {
		return msg
	}

	str.Write([]byte{0x1b, '['})
	str.WriteString(codes)
	str.WriteString(msg)
	str.Write([]byte{0x1b, '[', '0', 'm'})

	return str.String()
}

type inspectColorOption int

const (
	ColorsDefault inspectColorOption = iota
	ColorsOff
	ColorsOn
)

// type InspectOptions struct {
// 	// Whether colors should be enabled or not.  Defaults to on if stdout is a tty; otherwise off.
// 	Colors inspectColorOption

// 	// The maximum amount of recursion with which structs will be resolved from.
// 	// zero values are transl.ated to 3.  Negative values represent infinity.
// 	MaxDepth int

// 	// The maximum amount of items which will be displayed in an array or struct.
// 	// 9 is translated to 10.   Negative values represent infinity.
// 	MaxItemCount int

// 	// The current depth of the inspector.  This is an internal configuration.
// 	depth int
// }

func nameof(t reflect.Type, colors bool) string {
	name := ""

	for t.Kind() == reflect.Ptr {
		name += "*"
		t = t.Elem()
	}

	if t.Name() == "" || t.Name() == t.Kind().String() {
		name += t.Kind().String()
	} else {
		name += t.Name()
	}

	return format(name, colors, "3;1m")
}

type InspectString string

func isNil(v reflect.Value, vType reflect.Type) bool {
	switch vType.Kind() {
	case reflect.Slice, reflect.Ptr, reflect.Map, reflect.Func, reflect.Chan, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

func getCustomInspect(v reflect.Value, vType reflect.Type) (reflect.Value, bool) {
	clearReflectPrivateBit(&v)
	if !isNil(v, vType) {
		if meth, hasMeth := vType.MethodByName("InspectCustom"); hasMeth && meth.Type.NumOut() == 1 {
			var r []reflect.Value
			if meth.Type.NumIn() == 1 {
				r = v.Method(meth.Index).Call([]reflect.Value{})
			}
			return r[0], true
		}
	}

	return reflect.Value{}, false
}

func inspectReflectVal(v reflect.Value, depth int, colors, shouldHideName bool, parents []uintptr) string {
	var str strings.Builder
	if v.Kind() == reflect.Invalid {
		return format("nil", colors, "35;3m")
	}

	vType := v.Type()

	if vType.AssignableTo(reflect.TypeOf(InspectString(""))) {
		return v.String()
	}

	for vType.Kind() == reflect.Ptr || vType.Kind() == reflect.Interface {
		if custom, hasCustom := getCustomInspect(v, vType); hasCustom {
			str.WriteString(inspectReflectVal(custom, depth, colors, shouldHideName, parents))

			return str.String()
		}
		if v.IsZero() {
			str.WriteString(nameof(vType, colors))
			str.WriteByte(' ')
			str.WriteString(format("nil", colors, "35;3m"))

			return str.String()
		}

		if vType.Kind() == reflect.Ptr {

			clearReflectPrivateBit(&v)
			ptr := v.Pointer()
			for _, v := range parents {
				if v == ptr {
					str.WriteString(format("[recursive]", colors, "34m"))
					return str.String()
				}
			}
			parents = append(parents, ptr)

			if !shouldHideName {
				str.WriteByte('*')
			}
		}
		v = v.Elem()
		vType = v.Type()
	}

	if custom, hasCustom := getCustomInspect(v, vType); hasCustom {
		// str.WriteString(nameof(vType))
		// str.WriteByte(' ')
		str.WriteString(inspectReflectVal(custom, depth, colors, shouldHideName, parents))

		return str.String()
	}

	if vType == typTime {
		clearReflectPrivateBit(&v)
		tim := v.Interface().(time.Time)

		return format(tim.Format(time.RFC3339), colors, "34m")
	}

	switch vType.Kind() {
	case reflect.Func:
		str.WriteString("[function]")
	case reflect.Struct:
		if !shouldHideName {
			str.WriteString(nameof(vType, colors) + " {")
		} else {
			str.WriteByte('{')
		}

		rows := []string{}

		for num := 0; num < vType.NumField(); num++ {
			field := v.Field(num)

			rows = append(rows, fmt.Sprintf("%s: %s,", vType.Field(num).Name, inspectReflectVal(field, depth+1, colors, false, parents)))

		}

		str.WriteString(joinRows("}", rows, depth))
	case reflect.Slice, reflect.Array:
		if vType.Name() != "" {
			str.WriteString(nameof(vType, colors) + " {")
		} else {
			if v.Kind() == reflect.Array {
				str.WriteString(fmt.Sprintf("[%d]", v.Len()))
			} else {
				str.WriteString("[]")
			}
			str.WriteString(fmt.Sprintf("%s {", nameof(vType.Elem(), colors)))
		}

		rows := []string{}

		for i := 0; i < v.Len(); i++ {
			rows = append(rows, inspectReflectVal(v.Index(i), depth+1, colors, vType.Elem().Kind() != reflect.Interface, parents)+",")
		}

		str.WriteString(joinRows("}", rows, depth))
	case reflect.Map:
		if vType.Name() != "" {
			str.WriteString(nameof(vType, colors) + " {")
		} else {
			str.WriteString(format("map", colors, "31;1m"))
			str.WriteByte('[')
			str.WriteString(nameof(vType.Key(), colors))
			str.WriteByte(']')
			str.WriteString(nameof(vType.Elem(), colors))

		}

		str.WriteByte('{')
		rows := []string{}

		for items := v.MapRange(); items.Next(); {
			rows = append(rows, fmt.Sprintf(
				"%s: %s,",
				inspectReflectVal(items.Key(), depth+1, colors, false, parents),
				inspectReflectVal(items.Value(), depth+1, colors, false, parents),
			))
		}

		str.WriteString(joinRows("}", rows, depth))

	case reflect.String:
		if vType.Name() != vType.Kind().String() {
			str.WriteString(nameof(vType, colors) + " ")

		}
		str.WriteString(format("\""+fmt.Sprint(interfaceOf(v))+"\"", colors, "33m"))
	// numbers
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		if vType.Name() != vType.Kind().String() {
			str.WriteString(nameof(vType, colors) + " ")
		}
		str.WriteString(format(fmt.Sprint(interfaceOf(v)), colors, "32m"))

	default:
		if vType.Name() != vType.Kind().String() {
			str.WriteString(nameof(vType, colors))
			str.WriteByte(' ')
		}

		str.WriteString(fmt.Sprint(interfaceOf(v)))
	}

	return str.String()
}

func Inspect(val any) string {
	return inspectReflectVal(reflect.ValueOf(val), 0, true, false, []uintptr{})
}

func InspectBland(val any) string {
	return inspectReflectVal(reflect.ValueOf(val), 0, false, false, []uintptr{})
}

func Println(vals ...any) {
	strs := make([]string, len(vals))

	for i, val := range vals {
		strs[i] = Inspect(val)
	}

	fmt.Println(strings.Join(strs, " "))
}
