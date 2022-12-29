package generator

import (
	"fmt"
	"go/token"
	"main/inspector"
	"main/lexer"
	"main/parser"
	"reflect"
)

type Kind uint8

const (
	invalid Kind = iota
	numeric_start
	KindInt
	KindInt8
	KindInt16
	KindInt32
	KindInt64
	KindUint
	KindUint8
	KindUint16
	KindUint32
	KindUint64
	kindUntypedInt // internal type for integer constants
	KindFloat32
	KindFloat64
	numeric_end

	KindBool
	KindSlice
	KindString
	KindStruct
	KindInterface
)

func (k Kind) isNumeric() bool {
	return numeric_start < k && k < numeric_end
}

type AstError struct {
	message string
	node    parser.AstNode
}

func (a *AstError) Error() string {
	return a.message
}

func (a *AstError) Format(file *token.File) string {
	start := file.Position(a.node.Start())
	end := file.Position(a.node.End())

	return fmt.Sprintf("[%d:%d, %d:%d): %s", start.Line, start.Column, end.Line, end.Column, a.message)
}

func isConstant(v Typed) bool {
	_, ok := v.(ConstantValue)
	return ok
}

type Type interface {
	Kind() Kind
	Zero() any
	Name() string
	AssignableTo(other Type) bool
}

type Scoped interface {
	error(parser.AstNode, string, ...interface{})
	lookupIdentifier(parser.IdentifierNode) any
	parent() Scoped
	Lookup(name string) any
}

type Module struct {
	*Scope
	Declarations []Declare
	Exports      []string
}

type ConstantValue struct {
	value any
	typ   Type
}

func (c ConstantValue) Value() any {
	return c.value
}

func (c ConstantValue) Type() Type {
	return c.typ
}

type Function struct {
	p       Scoped
	Scope   *Scope
	Name    string
	Args    []*Argument
	Steps   []Step
	Returns Type
}

func (fn *Function) parent() Scoped {
	return fn.p
}

func (fn *Function) lookupIdentifier(node parser.IdentifierNode) any {
	return fn.p.lookupIdentifier(node)
}

func (fn *Function) Lookup(name string) any {
	return fn.p.Lookup(name)
}

func (fn *Function) error(node parser.AstNode, str string, rest ...interface{}) {
	fn.p.error(node, str, rest...)
}

func newScope(inherits Scoped) *Scope {
	return &Scope{
		inherits:    inherits,
		Identifiers: map[string]any{},
	}
}

type Value interface {
	// The value's raw value (if any).
	Value() any
	// The type of the value.
	Type() Type
}

type Writeable interface {
	Typed
	isWriteable()
}

type Scope struct {
	inherits    Scoped
	Identifiers map[string]any
	Errors      []AstError
}

func (s Scope) parent() Scoped {
	return s.inherits
}

func (s *Scope) error(node parser.AstNode, str string, vals ...interface{}) {
	if s.inherits != nil {
		s.inherits.error(node, str, vals...)
		return
	}

	s.Errors = append(s.Errors, AstError{message: fmt.Sprintf(str, vals...), node: node})
}

func (s *Scope) Lookup(name string) any {
	if val, ok := s.Identifiers[name]; ok {
		return val
	}

	if s.parent() == nil {
		return nil
	}

	return s.parent().Lookup(name)
}

func (s *Scope) lookupIdentifier(node parser.IdentifierNode) any {
	if val, ok := s.Identifiers[node.Target]; ok {
		return val
	}

	if s.parent() != nil {
		return s.parent().lookupIdentifier(node)
	}

	if typ, ok := Generics[node.Target]; ok {
		return typ
	}

	s.error(node, "unable to resolve name: %s", inspector.Inspect(node))

	return nil
}

func (s *Scope) lookupType(node parser.IdentifierNode) Type {
	ident := s.lookupIdentifier(node)

	if ident == nil {
		return nil
	}

	typ, ok := ident.(Type)
	if !ok {
		s.error(node, "expected '%s' to be a type", node.Target)
		return nil
	}

	return typ
}

func (s *Scope) lookupTyped(node parser.IdentifierNode) Typed {
	ident := s.lookupIdentifier(node)

	if ident == nil {
		return nil
	}

	val, ok := ident.(Typed)
	if !ok {
		s.error(node, "expected '%s' to be a type or variable", node.Target)
		return nil
	}

	return val
}

func (s *Scope) getReturnType() Type {
	for scope := Scoped(s); scope != nil; scope = scope.parent() {
		if fn, ok := scope.(*Function); ok {
			return fn.Returns
		}
	}

	// This shouldn't happen, we don't handle `return` outside of functions.
	panic("unexpected return")
}

func (s *Scope) assignValue(node parser.AssignmentNode) Step {
	// tood: maybe handle par
	value := s.lookupTyped(node.Assignee)

	if value == nil {
		return nil
	}

	if new := s.preEvaluate(node.Value); new != nil {
		writeable, ok := value.(Writeable)

		if !ok {
			s.error(node, "unable to assign value to target '%s'", node.Assignee.Target)
			return nil
		}

		if !new.Type().AssignableTo(writeable.Type()) {
			s.error(node.Value, "unable to assign value of type %s to value of type %s", new.Type().Name(), value.Type().Name())
			return nil
		}

		return Assign{
			Target: node.Assignee.Target,
			Value:  new,
		}
	}

	return nil
}

type Variable struct {
	Name         string
	InitialValue Typed
	typ          Type
}

func (v Variable) Type() Type {
	return v.typ
}

func (Variable) isWriteable() {}

func (s *Scope) declareVariable(node parser.VariableDeclarationNode) (dec Declare) {
	name := node.Name()
	if _, ok := s.Identifiers[name]; ok {
		s.error(node, "cannot redeclare identifier '%s'", name)
		return
	}

	var (
		typ Type
		val Typed
		err *AstError
	)

	if node.Type != nil {
		// TODO: support inline types.
		if typ = s.lookupType(node.Type.(parser.IdentifierNode)); err != nil {
			return
		}
	}

	if node.Value != nil {
		if val = s.preEvaluate(node.Value); val == nil {
			return
		}

		if typ == nil {
			typ = val.Type()

			if untyped, isUntyped := typ.(untypedInt); isUntyped {
				if untyped.AssignableTo(genericInt) {
					typ = genericInt
				} else if untyped.AssignableTo(genericInt64) {
					typ = genericInt64
				} else if untyped.AssignableTo(genericUint64) {
					typ = genericUint64
				} else {
					s.error(node, "idk how this happened lol")
				}
			}
		} else if !val.Type().AssignableTo(typ) {
			s.error(node, "type %s is unassignable to %s", val.Type().Name(), typ.Name())
		}
	}

	v := &Variable{
		Name:         name,
		InitialValue: val,
		typ:          typ,
	}

	s.Identifiers[name] = v

	return Declare{
		Name:     name,
		Variable: v,
	}
}

func (s *Scope) declareConstant(node parser.ConstantDeclarationNode) {
	name := node.Name()
	if _, ok := s.Identifiers[name]; ok {
		s.error(node, "cannot redeclare identifier '%s'", name)
	}

	var (
		typ Type
		val Typed
	)

	if node.Type != nil {
		// TODO: support inline types.
		if typ = s.lookupType(node.Type.(parser.IdentifierNode)); typ == nil {
			return
		}
	}

	if node.Value != nil {
		if val = s.preEvaluate(node.Value); val == nil {
			return
		}

		if typ == nil {
			typ = val.Type()
		} else if !val.Type().AssignableTo(typ) {
			s.error(node, "not assignable to type '%s': '%s'", typ.Name(), val.Type().Name())
			return
		}

		if !isConstant(val) {
			s.error(node, "value is not a constant")
			return
		}
	}

	s.Identifiers[name] = ConstantValue{
		value: val.(ConstantValue).value,
		typ:   typ,
	}

	return
}

type OperationSteps []Operation

type Operation struct {
	Value    Value
	Operator lexer.Token
}

func (s *Scope) getType(node parser.TypeNode) Type {

	switch node := node.(type) {
	case parser.IdentifierNode:
		return s.lookupType(node)

	default:
		return nil
	}
}

func (s *Scope) getArguments(nodes []parser.ArgumentDeclarationNode) []*Argument {
	if len(nodes) == 0 {
		return []*Argument{}
	}

	args := make([]*Argument, len(nodes))

	var typ Type

	// lookup args in reverse so earlier args will
	for i := len(args) - 1; i >= 0; i-- {
		node := nodes[i]

		if node.Type == nil && typ == nil {
			s.error(node, "expected argument '%s' to have a type", node.Name())
		}

		maybeType := s.getType(node.Type)

		if maybeType != nil {
			typ = maybeType
		} else {
			if typ == nil {
				s.error(node, "expected argument '%s' to have a type", node.Name())
			}
		}

		args[i] = &Argument{
			Name: node.Name(),
			typ:  typ,
		}

	}

	return args
}

func (s *Scope) handleBlock(block parser.BlockNode) (steps []Step) {
	for _, step := range block.Steps {
		switch node := step.(type) {
		case parser.VariableDeclarationNode:
			steps = append(steps, s.declareVariable(node))

		case parser.ConstantDeclarationNode:
			s.declareConstant(node)
		case parser.AssignmentNode:
			steps = append(steps, s.assignValue(node))
		case parser.IfNode:
			steps = append(steps, s.handleIf(node))
		case parser.CallNode:
			steps = append(steps, s.handleCall(node))
		case parser.ReturnNode:
			steps = append(steps, s.handleReturn(node))
			return
		default:
			panic(fmt.Errorf("unhandled node: %s", reflect.TypeOf(node).Name()))
		}
	}

	return
}

func (s *Scope) handleReturn(node parser.ReturnNode) Step {
	typ := s.getReturnType()

	if typ == nil && node.Value != nil {
		s.error(node, "unexpected return value; function returns nothing")
		return nil
	}

	val := s.preEvaluate(node.Value)

	if val == nil {
		return nil
	}

	if !val.Type().AssignableTo(typ) {
		s.error(node, "invalid return: expected value of type '%s'; received value of type '%s';", typ.Name(), val.Type().Name())
	}

	return Return{Value: val}
}

func (s *Scope) handleTopLevelFunction(node parser.FunctionNode) *Function {
	fn := &Function{p: s}
	if node.Returns.Target != "" {
		fn.Returns = s.lookupType(node.Returns)
	}
	fn.Scope = newScope(fn)

	defer fn.Scope.removeConstants()

	fn.Args = s.getArguments(node.Arguments.Arguments)

	for _, arg := range fn.Args {
		fn.Scope.Identifiers[arg.Name] = arg
	}

	return fn
}

// TODO: do this better.
func (s *Scope) removeConstants() {
	consts := []string{}

	for name, val := range s.Identifiers {
		if _, ok := val.(ConstantValue); ok {
			consts = append(consts, name)
		}
	}

	for _, name := range consts {
		delete(s.Identifiers, name)
	}
}

func handleChildBlock(s Scoped, node parser.BlockNode) Block {
	child := newScope(s)

	return Block{
		Scope: child,
		Steps: child.handleBlock(node),
	}
}

func (s *Scope) handleIf(node parser.IfNode) Step {
	val := s.preEvaluate(node.Condition)

	if val, ok := val.(ConstantValue); ok {
		scope := newScope(s)

		if !reflect.ValueOf(val.value).IsZero() {
			return handleChildBlock(s, node.Then)
		}

		// We can omit the first block; it's unreachable code.
		if len(node.ElseIf) > 0 {
			new := node.ElseIf[0]

			if len(node.ElseIf) > 1 {
				new.ElseIf = node.ElseIf[1:]
			}

			new.Else = node.Else
			return s.handleIf(new)
		}

		if node.Else != nil {
			return Block{
				Scope: scope,
				Steps: scope.handleBlock(*node.Else),
			}
		}

		return nil
	}

	stp := If{
		Condition: val,
		Then:      handleChildBlock(s, node.Then),
		ElseIf:    make([]If, len(node.ElseIf)),
	}

	for i, elif := range node.ElseIf {
		stp.ElseIf[i] = If{
			Condition: s.preEvaluate(elif.Condition),
			Then:      handleChildBlock(s, elif.Then),
		}
	}

	if node.Else != nil {
		bl := handleChildBlock(s, *node.Else)
		stp.Else = &bl
	}

	return stp
}

func ProcessModule(ast parser.ModuleNode) (mod Module) {
	// TODO: consider unordered declarations.
	// This can be done by lazy-loading everything then in lookupValue evaluate
	// the variable / constant with a stack to detect recursive dependencies.
	//
	// e.g.:
	// ```
	// const a = b;
	// const b = 32;
	// ````

	scope := newScope(nil)
	defer scope.removeConstants()
	mod.Scope = scope

	for _, node := range ast.Nodes {
		if pub, ok := node.(parser.PublicNode); ok {
			mod.Exports = append(mod.Exports, pub.Node.(parser.DeclarationNode).Name())
			node = pub.Node
		}

		switch node := node.(type) {
		case parser.VariableDeclarationNode:

			mod.Declarations = append(mod.Declarations, scope.declareVariable(node))
		case parser.ConstantDeclarationNode:
			scope.declareConstant(node)
		case parser.ModuleFunctionDeclarationNode:
			// TODO: assert name doesn't already exist.
			fn := scope.handleTopLevelFunction(node.Body)
			fn.Name = node.Name()
			scope.Identifiers[node.Name()] = fn

			// Defer loading of the function's steps until we read every function
			defer func() { fn.Steps = fn.Scope.handleBlock(node.Body.Block) }()

		default:
			panic(fmt.Errorf("unhandled node: %s", reflect.TypeOf(node).Name()))
		}
	}

	return
}

func (s *Scope) handleInlineFunction(node parser.FunctionNode) (fn *Function) {
	fn.Scope = newScope(fn)
	scope := newScope(fn)
	fn.Args = scope.getArguments(node.Arguments.Arguments)
	fn.Steps = scope.handleBlock(node.Block)

	return
}

func (s *Scope) handleCall(node parser.CallNode) Call {
	var fn *Function
	switch callee := node.Callee.(type) {
	case parser.IdentifierNode:
		i := s.lookupIdentifier(callee)
		ident, ok := i.(*Function)

		if !ok {
			s.error(callee, "not a function")
			return Call{}
		}
		fn = ident
	case parser.FunctionNode:
		fn = s.handleInlineFunction(callee)
	default:
		panic("unexpected")
	}

	// TODO: if variadic functions are added, this logic needs to change.
	if len(node.Arguments) != len(fn.Args) {
		s.error(node, "incorrect number of arguments for function; expected %d", len(fn.Args))
		return Call{}
	}

	step := Call{
		Target:    fn,
		Arguments: make([]Typed, len(fn.Args)),
	}

	for i, arg := range node.Arguments {
		val := s.preEvaluate(arg)

		if !val.Type().AssignableTo(fn.Args[i].Type()) {
			s.error(node, "invalid argument type '%s'; expected '%s'", val.Type().Name(), fn.Args[i].Type().Name())
		}

		step.Arguments[i] = val
	}

	return step
}
