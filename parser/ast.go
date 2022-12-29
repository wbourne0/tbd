package parser

import (
	"fmt"
	"go/token"
	"main/inspector"
	"main/lexer"
	"strings"
)

type AstNode interface {
	Start() token.Pos
	End() token.Pos
}

type BaseNode struct {
	start token.Pos
}

func (b BaseNode) Start() token.Pos {
	return b.start
}

// A node which can be accessed at the top level of a module.
type TopLevelNode interface {
	AstNode
	isTopLevelNode()
}

// A node which could represent any sort of value.
type ValueNode interface {
	AstNode
	isValueNode()
}

// A type declaration.
type TypeNode interface {
	AstNode
	isTypeNode()
}

// The declaration of a variable, constant, or import..
type DeclarationNode interface {
	AstNode
	TopLevelNode
	StepNode
	isDeclarationNode()
	// The name by which the declared can be referenced by.
	Name() string
}

// Elements of an array, struct, map, etc.
// This is either ElementsNode or InvalidNode.
type ElementListNode interface {
	AstNode
	ValueNode
	isElementListNode()
}

// A logical step - eg a function call, modifying a variable, or flow control.
type StepNode interface {
	AstNode
	isStepNode()
}

// A node used to represent invalid syntax or a parsing error.
type InvalidNode struct {
	BaseNode
	// The associated error.
	err error
	// The end of the invalid segment.
	end token.Pos
}

func (i InvalidNode) End() token.Pos {
	return i.end
}

func (InvalidNode) isStepNode()        {}
func (InvalidNode) isTypeNode()        {}
func (InvalidNode) isBlockNode()       {}
func (InvalidNode) isValueNode()       {}
func (InvalidNode) isTopLevelNode()    {}
func (InvalidNode) isFunctionNode()    {}
func (InvalidNode) isArgumentsNode()   {}
func (InvalidNode) isDeclarationNode() {}
func (InvalidNode) isElementListNode() {}

func (InvalidNode) Name() string { return "invalid" }

// An identity, which (if valid) references a function, variable, constant, import, etc.
type IdentifierNode struct {
	BaseNode
	// The target variable or constant.
	Target string
}

func (i IdentifierNode) End() token.Pos {
	return i.start + token.Pos(len(i.Target))
}

func (i IdentifierNode) InspectCustom() inspector.InspectString {
	return inspector.InspectString(i.Target)
}

func (IdentifierNode) isTypeNode()  {}
func (IdentifierNode) isValueNode() {}

// Access to a property of a referenced variable.
type PropertyAccessNode struct {
	BaseNode
	// The value which the property belongs to.
	PropertyOf ValueNode
	// The property that is being accessed.
	Property IdentifierNode
}

func (p PropertyAccessNode) End() token.Pos {
	return p.Property.End()
}

func (p PropertyAccessNode) InspectCustom() inspector.InspectString {
	return inspector.InspectString(inspector.Inspect(p.PropertyOf) + "." + inspector.Inspect(p.Property))
}

func (PropertyAccessNode) isValueNode() {}

// A constant integer value.
type IntegerNode struct {
	BaseNode
	// The value of the number number.
	Value uint64
	// The minimum amount of bits required to store the value.
	Bits uint8
	// The length of the raw integer string.
	strlen int
}

func (i IntegerNode) End() token.Pos {
	return i.start + token.Pos(i.strlen)
}

func (i IntegerNode) InspectCustom() uint64 {
	return i.Value
}

func (IntegerNode) isValueNode() {}

// A constant floating-point value.
type FloatNode struct {
	BaseNode
	// The stringified form of the number.
	Value float64
	// Whether the float is double precision (64 bit) or single-precision (32 bit)
	isDouble bool
	// The length of the raw float string.
	strlen int
}

func (f FloatNode) End() token.Pos {
	return f.start + token.Pos(f.strlen)
}

func (FloatNode) isValueNode() {}

// A constant string value.
type StringNode struct {
	BaseNode
	// The value of the string.
	Value string
	// The end of the string.
	end int
}

func (s StringNode) InspectCustom() string {
	return s.Value
}

func (s StringNode) End() token.Pos {
	// + 2 characters for quotes
	return s.start + token.Pos(len(s.Value))
}

func (StringNode) isValueNode() {}

// A function call.
type CallNode struct {
	BaseNode
	// The function being called.
	Callee ValueNode
	// The arguments the target function is being called with.
	Arguments []ValueNode
	// The end of the function call (ie: the terminating ')').
	end token.Pos
}

func (c CallNode) End() token.Pos {
	return c.end
}

func (c CallNode) InspectCustom() inspector.InspectString {
	argStrings := make([]string, len(c.Arguments))

	for i, arg := range c.Arguments {
		argStrings[i] = inspector.Inspect(arg)
	}

	return inspector.InspectString(fmt.Sprintf("%s(%s)", inspector.Inspect(c.Callee), strings.Join(argStrings, ", ")))
}

func (CallNode) isStepNode()  {}
func (CallNode) isValueNode() {}

// An import to a module or external package.
// TODO: Do
type ImportDeclarationNode struct {
	BaseNode
	// The path of the import.
	Path string
	// The alias of the package.  Optional.
	Alias string
}

func (ImportDeclarationNode) isTopLevelNode()    {}
func (ImportDeclarationNode) isDeclarationNode() {}

// A declaration of a constant value.
type ConstantDeclarationNode struct {
	BaseNode
	// The type of the variable (can be empty).
	Type TypeNode
	// The name of the constant.
	name string
	// The value of the constant variable.
	Value ValueNode
}

func (c ConstantDeclarationNode) InspectCustom() inspector.InspectString {
	if c.Type != nil {
		return inspector.InspectString(fmt.Sprintf("const %s %s = %s", c.name, inspector.Inspect(c.Type), inspector.Inspect(c.Value)))
	}

	return inspector.InspectString(fmt.Sprintf("const %s = %s", c.name, inspector.Inspect(c.Value)))
}

func (c ConstantDeclarationNode) End() token.Pos {
	return c.Value.End()
}

func (ConstantDeclarationNode) isStepNode()        {}
func (ConstantDeclarationNode) isTopLevelNode()    {}
func (ConstantDeclarationNode) isDeclarationNode() {}

// The name of the constant
func (c ConstantDeclarationNode) Name() string {
	return c.name
}

// A declaration of a mutable variable.
type VariableDeclarationNode struct {
	BaseNode
	// The name of the variable.
	name string
	// The type of the variable.
	Type TypeNode
	// The initial value of the variable (if any).
	Value ValueNode
}

func (v VariableDeclarationNode) InspectCustom() inspector.InspectString {
	var str strings.Builder
	str.WriteString("var " + v.name)

	if v.Type != nil {
		str.WriteByte(' ')
		str.WriteString(inspector.Inspect(v.Type))
	}

	if v.Value != nil {
		str.WriteString(" = ")
		str.WriteString(inspector.Inspect(v.Value))
	}

	return inspector.InspectString(str.String())
}

func (v VariableDeclarationNode) End() token.Pos {
	if v.Value != nil {
		return v.Value.End()
	}

	return v.Type.End()
}

func (VariableDeclarationNode) isStepNode()        {}
func (VariableDeclarationNode) isTopLevelNode()    {}
func (VariableDeclarationNode) isDeclarationNode() {}

// The name of the variable
func (v VariableDeclarationNode) Name() string {
	return v.name
}

// A declaration of a function argument.
type ArgumentDeclarationNode struct {
	BaseNode
	// The name of the argument.
	name string
	// The type of the argument.
	Type TypeNode
}

func (a ArgumentDeclarationNode) InspectCustom() inspector.InspectString {
	if a.Type != nil {
		return inspector.InspectString(fmt.Sprintf("%s %s", a.name, inspector.Inspect(a.Type)))
	}

	return inspector.InspectString(a.name)
}

func (a ArgumentDeclarationNode) End() token.Pos {
	if a.Type != nil {
		return a.Type.End()
	}

	return a.start + token.Pos(len(a.name))
}

// The name of the argument.
func (arg ArgumentDeclarationNode) Name() string {
	return arg.name
}

func (ArgumentDeclarationNode) isDeclarationNode() {}

// The arguments of a function.
type ArgumentDeclarationsNode struct {
	BaseNode
	// The arguments of the function.
	Arguments []ArgumentDeclarationNode
	// The end of the argument list (ie, closing ')').
	end token.Pos
}

func (a ArgumentDeclarationsNode) InspectCustom() inspector.InspectString {
	args := make([]string, len(a.Arguments))

	for i, arg := range a.Arguments {
		args[i] = inspector.Inspect(arg)
	}

	return inspector.InspectString(fmt.Sprintf("(%s)", strings.Join(args, ", ")))
}

func (a ArgumentDeclarationsNode) End() token.Pos {
	return a.end
}

func (a ArgumentDeclarationsNode) isArgumentsNode() {}

// An abstract declaration of a function.
type FunctionNode struct {
	BaseNode

	// The arguments of the function.
	Arguments ArgumentDeclarationsNode
	// The function's block.
	Block BlockNode

	// The return type of the function.
	Returns IdentifierNode
}

func (f FunctionNode) End() token.Pos {
	return f.Block.End()
}

func (FunctionNode) isStepNode()     {}
func (FunctionNode) isValueNode()    {}
func (FunctionNode) isFunctionNode() {}

func (f FunctionNode) InspectCustom() inspector.InspectString {
	return inspector.InspectString(fmt.Sprintf("%s {\n%s\n}", inspector.Inspect(f.Arguments), inspector.Inspect(f.Block)))
}

// A declaration of a top-level function.
type ModuleFunctionDeclarationNode struct {
	BaseNode
	Body FunctionNode
	// The name of the function.
	name string
}

func (m ModuleFunctionDeclarationNode) InspectCustom() inspector.InspectString {
	return inspector.InspectString("func " + m.name + inspector.Inspect(m.Body))
}
func (m ModuleFunctionDeclarationNode) End() token.Pos {
	return m.Body.End()
}
func (ModuleFunctionDeclarationNode) isTopLevelNode()    {}
func (ModuleFunctionDeclarationNode) isDeclarationNode() {}

// The name of the function.
func (m ModuleFunctionDeclarationNode) Name() string {
	return m.name
}

// A declaration of a method.
type MethodDeclarationNode struct {
	BaseNode
	Body FunctionNode
	// The name of the method.
	name string

	// The name of the type which the method belongs to.
	MethodOf string
	// Whether the function has a pointer receiver (this is the address of a value)
	// or not (this is a copy of a value of a type).
	IsPtrReceiver bool
}

func (m MethodDeclarationNode) Name() string {
	return "<method>"
}

func (m MethodDeclarationNode) InspectCustom() inspector.InspectString {
	if m.IsPtrReceiver {
		return inspector.InspectString(fmt.Sprintf("func *%s.%s%s", m.MethodOf, m.name, inspector.Inspect(m.Body)))
	}

	return inspector.InspectString(fmt.Sprintf("func %s.%s%s", m.MethodOf, m.name, inspector.Inspect(m.Body)))
}

func (m MethodDeclarationNode) End() token.Pos {
	return m.Body.End()
}
func (MethodDeclarationNode) isTopLevelNode()    {}
func (MethodDeclarationNode) isDeclarationNode() {}

// An assignment to a variable or property.
type AssignmentNode struct {
	BaseNode
	// The reference to the item being assigned a value.
	Assignee IdentifierNode
	// The value which is to be assigned to the variable.
	Value ValueNode
	end   token.Pos
}

func (a AssignmentNode) InspectCustom() inspector.InspectString {
	return inspector.InspectString(fmt.Sprintf("%s = %s", inspector.Inspect(a.Assignee), inspector.Inspect(a.Value)))
}
func (a AssignmentNode) End() token.Pos {
	return a.Value.End()
}
func (AssignmentNode) isStepNode() {}

// An operation consisting of left and right sides.
type BinaryOperationNode struct {
	BaseNode
	// The value on the left of the operator.
	Left ValueNode
	// The value on the right of the operator.
	Right ValueNode
	// The operator used in the operation.
	Operator lexer.Token
}

func (b BinaryOperationNode) End() token.Pos {
	return b.Right.End()
}

func (b BinaryOperationNode) InspectCustom() inspector.InspectString {
	var left, right string

	if _, ok := b.Left.(BinaryOperationNode); ok {
		left = fmt.Sprintf("(%s)", inspector.Inspect(b.Left))
	} else {
		left = inspector.Inspect(b.Left)
	}

	if _, ok := b.Right.(BinaryOperationNode); ok {
		right = fmt.Sprintf("(%s)", inspector.Inspect(b.Right))
	} else {
		right = inspector.Inspect(b.Right)
	}

	return inspector.InspectString(fmt.Sprintf("%s %s %s", left, b.Operator.String(), right))
}

func (BinaryOperationNode) isValueNode() {}

// An operation on a single value.
type UnaryOperationNode struct {
	BaseNode
	// The value which is being operated on.
	Operand ValueNode
	// The operator used in the operation.
	Operator lexer.Token
}

func (u UnaryOperationNode) End() token.Pos {
	return u.Operand.End()
}

func (u UnaryOperationNode) InspectCustom() inspector.InspectString {
	if _, ok := u.Operand.(IdentifierNode); ok {
		return inspector.InspectString(u.Operator.String() + inspector.Inspect(u.Operand))
	}

	return inspector.InspectString(fmt.Sprintf("%s(%s)", u.Operator.String(), inspector.Inspect(u.Operand)))
}

func (UnaryOperationNode) isValueNode() {}

type SuffixUnaryOperationNode struct {
	UnaryOperationNode
	// The end of the operation (ie, last character of the operator).
	end token.Pos
}

func (s SuffixUnaryOperationNode) End() token.Pos {
	return s.end
}

func (u SuffixUnaryOperationNode) InspectCustom() inspector.InspectString {
	if _, ok := u.Operand.(IdentifierNode); ok {
		return inspector.InspectString(inspector.Inspect(u.Operand) + u.Operator.String())
	}

	return inspector.InspectString(fmt.Sprintf("(%s)%s", inspector.Inspect(u.Operand), u.Operator.String()))
}

func (SuffixUnaryOperationNode) isValueNode() {}
func (SuffixUnaryOperationNode) isStepNode()  {}

// A slice prefix.
type SlicePrefixNode struct {
	BaseNode
	// The type of element in the array.
	SliceOf TypeNode
}

func (s SlicePrefixNode) InspectCustom() inspector.InspectString {

	return inspector.InspectString("[]" + inspector.Inspect(s.SliceOf))
}
func (s SlicePrefixNode) End() token.Pos {
	return s.SliceOf.End()
}
func (SlicePrefixNode) isTypeNode() {}

// An array prefix.
type ArrayPrefixNode struct {
	BaseNode
	// The length of the array.
	// if nil,  the length of the array is defined by the compiler.
	Len ValueNode
	// The type of element in the array.
	ArrayOf TypeNode
}

func (a ArrayPrefixNode) InspectCustom() inspector.InspectString {
	if a.Len == nil {
		return inspector.InspectString("[...]" + inspector.Inspect(a.ArrayOf))
	}

	return inspector.InspectString(fmt.Sprintf("[%s]%s", inspector.Inspect(a.Len), inspector.Inspect(a.ArrayOf)))
}

func (a ArrayPrefixNode) End() token.Pos {
	return a.ArrayOf.End()
}
func (ArrayPrefixNode) isTypeNode() {}

// A slice prefix and contents.
type SliceValueNode struct {
	BaseNode

	// The type of slice.
	Prefix SlicePrefixNode
	// The elements in the slice.
	Elements ElementListNode
}

func (s SliceValueNode) InspectCustom() inspector.InspectString {
	return inspector.InspectString(inspector.Inspect(s.Prefix) + inspector.Inspect(s.Elements))
}

func (s SliceValueNode) End() token.Pos {
	return s.Elements.End()
}

func (SliceValueNode) isValueNode() {}

// An array of items.
type ArrayValueNode struct {
	BaseNode

	// The type of array.
	Prefix ArrayPrefixNode
	// The elements of the array.
	Elements ElementListNode
}

func (a ArrayValueNode) InspectCustom() inspector.InspectString {
	return inspector.InspectString(inspector.Inspect(a.Prefix) + inspector.Inspect(a.Elements))
}

func (a ArrayValueNode) End() token.Pos {
	return a.Elements.End()
}
func (ArrayValueNode) isValueNode() {}

// Reperents the literal value `nil`
type NilNode struct {
	BaseNode
}

func (n NilNode) End() token.Pos {
	return n.start + token.Pos(len(lexer.NIL.String()))
}

func (NilNode) isValueNode() {}

// An element of a slice or array.
type ElementNode struct {
	BaseNode
	// The element's key, if any.  If null, it is determined by the position of elements.
	// If the first element has a key, all elements should; if the first element
	// has a nil key none of the elements should have non-nil keys.
	Key ValueNode
	// The value of the element.
	Value ValueNode
}

func (e ElementNode) InspectCustom() inspector.InspectString {
	if e.Key == nil {
		return inspector.InspectString(inspector.Inspect(e.Value))
	}

	return inspector.InspectString(fmt.Sprintf("%s: %s", inspector.Inspect(e.Key), inspector.Inspect(e.Value)))
}

func (e ElementNode) End() token.Pos {
	return e.Value.End()
}

// A list of elements.
type ElementsNode struct {
	BaseNode
	// The elements in the list.
	Elements []ElementNode
	// The closing }.
	end token.Pos
}

func (e ElementsNode) InspectCustom() inspector.InspectString {
	elemString := inspector.Inspect(e.Elements)
	return inspector.InspectString(fmt.Sprintf("{%s}", elemString[25:len(elemString)-1]))
}

func (e ElementsNode) End() token.Pos {
	return e.end
}

func (ElementsNode) isElementListNode() {}
func (ElementsNode) isValueNode()       {}

type BlockNode struct {
	BaseNode
	Steps []StepNode
	end   token.Pos
}

func (b BlockNode) InspectCustom() inspector.InspectString {
	steps := make([]string, len(b.Steps))

	for i, step := range b.Steps {
		steps[i] = fmt.Sprintf("%d. %s", i+1, inspector.Inspect(step))
	}

	return inspector.InspectString(strings.Join(steps, "\n"))
}

func (b BlockNode) End() token.Pos {
	return b.end
}
func (b BlockNode) isBlockNode() {}

// A return statement.
type ReturnNode struct {
	BaseNode
	// The value returned.
	Value ValueNode
}

func (r ReturnNode) End() token.Pos {
	return r.Value.End()
}
func (r ReturnNode) InspectCustom() inspector.InspectString {
	return inspector.InspectString("return " + inspector.Inspect(r.Value))
}
func (ReturnNode) isStepNode() {}

// A module.
type ModuleNode struct {
	Nodes []TopLevelNode
}

func (m ModuleNode) InspectCustom() inspector.InspectString {
	nodes := make([]string, len(m.Nodes))

	for i, node := range m.Nodes {
		nodes[i] = inspector.Inspect(node) + "\n"
	}

	return inspector.InspectString(strings.Join(nodes, "\n"))
}

// The indexing of a value.
type IndexNode struct {
	BaseNode
	// The value being indexed (presumably an array or slice).
	IndexOf ValueNode
	// The key by whch the target is being indexed by.
	Key ValueNode

	end token.Pos
}

func (i IndexNode) InspectCustom() inspector.InspectString {
	return inspector.InspectString(fmt.Sprintf("%s[%s]", inspector.Inspect(i.IndexOf), inspector.Inspect(i.Key)))
}
func (i IndexNode) End() token.Pos {
	return i.end
}
func (IndexNode) isValueNode() {}

// Marks a declaration as public.
type PublicNode struct {
	BaseNode
	// The declaration which should be public.
	Node TopLevelNode
}

func (p PublicNode) End() token.Pos {
	return p.Node.End()
}

func (p PublicNode) isTopLevelNode()    {}
func (p PublicNode) isDeclarationNode() {}
func (p PublicNode) InspectCustom() inspector.InspectString {
	return inspector.InspectString("public " + inspector.Inspect(p.Node))
}

// An if statment.
type IfNode struct {
	BaseNode
	// The declaration which should be public.
	Condition ValueNode
	// The block which should be evaluated provide that the condition is truthy.
	Then BlockNode
	// Other conditions that should be evaluated if the primary condition is false prior to evaluating `else`.
	ElseIf []IfNode
	// The block which should be evaluated provided that the condition is falsy.
	Else *BlockNode
}

func (i IfNode) End() token.Pos {
	if i.Else.start == 0 {
		return i.Then.end
	}

	return i.Else.end
}

func (IfNode) isStepNode() {}
