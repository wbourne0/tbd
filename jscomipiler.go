package main

import (
	"fmt"
	"main/generator"
	"main/inspector"
	"strings"
)

func stringifyTyped(val generator.Typed) string {
	var content strings.Builder

	switch val := val.(type) {
	case generator.BinaryOperation:

		if _, ok := val.Left.(generator.BinaryOperation); ok {
			content.WriteByte('(')
			content.WriteString(stringifyTyped(val.Left))
			content.WriteByte(')')
		} else {
			content.WriteString(stringifyTyped(val.Left))
		}

		content.WriteString(val.Operator.String())

		if _, ok := val.Right.(generator.BinaryOperation); ok {
			content.WriteByte('(')
			content.WriteString(stringifyTyped(val.Right))
			content.WriteByte(')')
		} else {
			content.WriteString(stringifyTyped(val.Right))
		}
	case generator.UnaryOperation:
		content.WriteString(val.Operator.String())
		content.WriteByte('(')
		content.WriteString(stringifyTyped(val.Operand))
		content.WriteByte(')')
	case generator.Call:
		if val.Target.Name != "" {
			content.WriteString(val.Target.Name)
		} else {
			content.WriteString("(function(")
			for i, v := range val.Target.Args {
				content.WriteString(v.Name)
				if i != len(val.Target.Args)-1 {
					content.WriteByte(',')
				}
			}

			content.WriteByte(')')
			content.WriteString(stringifyBlock(val.Target.Steps))
		}

		content.WriteByte('(')
		for i, v := range val.Arguments {
			content.WriteString(stringifyTyped(v))
			if i != len(val.Target.Args)-1 {
				content.WriteByte(',')
			}
		}

		content.WriteByte(')')
	case *generator.Variable:
		fmt.Println(inspector.Inspect(val))
		return val.Name
	case *generator.Argument:
		return val.Name

	case generator.Value:
		content.WriteString(inspector.InspectBland(val.Value()))
	}

	return content.String()
}

func stringifyBlock(steps []generator.Step) string {
	var content strings.Builder

	content.WriteByte('{')

	for _, step := range steps {
		switch st := step.(type) {
		case generator.Declare:
			content.WriteString("let ")
			content.WriteString(st.Name)

			if st.InitialValue != nil {
				content.WriteByte('=')
				content.WriteString(stringifyTyped(st.InitialValue))
			}
			content.WriteByte(';')

		case generator.Assign:
			content.WriteString(st.Target)
			content.WriteByte('=')
			content.WriteString(stringifyTyped(st.Value))
			content.WriteByte(';')
		case generator.If:
			content.WriteString("if(")
			content.WriteString(stringifyTyped(st.Condition))
			content.WriteByte(')')
			content.WriteString(stringifyBlock(st.Then.Steps))

			for _, elif := range st.ElseIf {
				content.WriteString("else if(")
				content.WriteString(stringifyTyped(elif.Condition))
				content.WriteString(stringifyBlock(elif.Then.Steps))
			}

			if st.Else != nil {
				content.WriteString("else")
				content.WriteString(stringifyBlock(st.Else.Steps))
			}

		case generator.Block:
			content.WriteString("if(1)")
			content.WriteString(stringifyBlock(st.Steps))
		case generator.Call:
			if st.Target.Name != "" {
				content.WriteString(st.Target.Name)
			} else {
				content.WriteString("(function(")
				for i, v := range st.Target.Args {
					content.WriteString(v.Name)
					if i != len(st.Target.Args)-1 {
						content.WriteByte(',')
					}
				}

				content.WriteByte(')')
				content.WriteString(stringifyBlock(st.Target.Steps))
			}

			content.WriteByte('(')
			for i, v := range st.Arguments {
				content.WriteString(stringifyTyped(v))
				if i != len(st.Target.Args)-1 {
					content.WriteByte(',')
				}
			}

			content.Write([]byte{')', ';'})
		case generator.Return:
			content.WriteString("return ")
			content.WriteString(stringifyTyped(st.Value))
		}
	}

	content.WriteByte('}')

	return content.String()
}

func stringify(mod generator.Module) string {
	var content strings.Builder

	for _, dec := range mod.Declarations {
		if dec == (generator.Declare{}) {
			continue
		}

		content.WriteString("let ")
		content.WriteString(dec.Name)

		if dec.InitialValue != nil {
			content.WriteByte('=')
			content.WriteString(stringifyTyped(dec.InitialValue))
		}
		content.WriteByte(';')
	}

	for name, ident := range mod.Scope.Identifiers {
		fn, ok := ident.(*generator.Function)

		if !ok {
			continue
		}

		content.WriteString("function ")
		content.WriteString(name)
		content.WriteByte('(')

		for i, arg := range fn.Args {
			content.WriteString(arg.Name)
			if i != len(fn.Args)-1 {
				content.WriteByte(',')
			}
		}
		content.WriteByte(')')

		content.WriteString(stringifyBlock(fn.Steps))
	}

	// content.WriteString("main();")

	return content.String()
}
