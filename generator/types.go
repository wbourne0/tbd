package generator

type Step interface{ isStep() }

type Assign struct {
	Target string
	Value  Typed
}

func (Assign) isStep() {}

type Call struct {
	Target    *Function
	Arguments []Typed
}

func (c Call) Type() Type {
	return c.Target.Returns
}

func (Call) isStep() {}

type Declare struct {
	Name string
	*Variable
}

func (Declare) isStep() {}

type Block struct {
	*Scope
	Steps []Step
}

func (Block) isStep() {}

type If struct {
	Condition Typed
	Then      Block
	ElseIf    []If
	Else      *Block
}

func (If) isStep() {}

// A function argument.
type Argument struct {
	typ Type
	// The name of the argument.
	Name string
}

func (a Argument) Type() Type {
	return a.typ
}

func (a Argument) isWriteable() {}

type Return struct {
	Value Typed
}

func (r Return) isStep() {}
