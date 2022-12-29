package factorio

import (
	"fmt"
	"io"
	"main/generator"
	"main/inspector"
	"main/lexer"
)

// type Connection struct {

// }

// type Component interface {
// 	hasOutput() bool
// }

var debugNet *Network

const debugEntityId = 27

type Network struct {
	// signalNum int
	members []*Connector
	isGreen bool
	cellID  Constant
}

func (n *Network) merge(o *Network) {
	n.members = append(n.members, o.members...)

	for _, mem := range o.members {
		mem.Network = n
	}
}

func (n *Network) nextCellID() Constant {
	n.cellID++
	return n.cellID
}

func (n *Network) shareSignal(other *Network, sig *Signal) {
	var (
		in  = &Arithmetic{Operator: ArithmeticOperationOr}
		out = &Arithmetic{Operator: ArithmeticOperationOr}
	)

	in.output = sig
	out.output = sig
	in.primaryInput = sig
	out.primaryInput = sig
	in.secondaryInput = Constant(0)
	out.secondaryInput = Constant(0)

	n.connectInput(out)
	n.connectOutput(in)

	other.connectInput(in)
	other.connectOutput(out)

	return
}

type Constant int32

func (Constant) isInput()          {}
func (Constant) isSecondaryInput() {}

type Input interface {
	isInput()
}

type InputComponent struct {
	primaryInput   *Signal
	secondaryInput Input
	inputGreen     *Connector
	inputRed       *Connector
}

func (i *InputComponent) setInput(isGreen bool, conn *Connector) {
	if isGreen {
		if i.inputGreen != nil {
			panic("green already set")
		}

		i.inputGreen = conn
		return
	}

	if i.inputRed != nil {
		panic("red already set")
	}

	i.inputRed = conn
	return
}

type OutputComponent struct {
	outputGreen *Connector
	outputRed   *Connector
}

func (o *OutputComponent) setOutput(isGreen bool, conn *Connector) {
	if isGreen {
		if o.outputGreen != nil {
			panic("green output already set")
		}

		o.outputGreen = conn
		return
	}

	if o.outputRed != nil {
		panic("red output already set")
	}

	o.outputRed = conn

}

type ComponentOutput interface {
	Component
	setOutput(isGreen bool, conn *Connector)
}

type ComponentInput interface {
	Component
	setInput(isGreen bool, conn *Connector)
}

type Component interface {
	Entity() ConnectorEntity
}

type ConstantCombinatorItem struct {
	signal *Signal
	count  int32
}

type Emitter struct {
	OutputComponent
	items []ConstantCombinatorItem
	ent   *ConstantCombinatorEntity
}

func (e *Emitter) Entity() ConnectorEntity {
	if e.ent == nil {
		ent := new(ConstantCombinatorEntity)

		ent.ControlBehavior.Filters = make([]ConstantFilter, len(e.items))

		for i, v := range e.items {
			ent.ControlBehavior.Filters[i] = ConstantFilter{
				Signal: *v.signal,
				Count:  int(v.count),
				Index:  i + 1,
			}
		}

		e.ent = ent
		e.items = nil

		if e.outputRed != nil {
			e.outputRed.isSecondary = false
		}

		if e.outputGreen != nil {
			e.outputGreen.isSecondary = false
		}

	}

	return e.ent
}

type Arithmetic struct {
	InputComponent
	OutputComponent
	Operator ArithmeticOperation
	output   *Signal
	ent      *ArithmeticCombinatorEntity
}

func (a *Arithmetic) Entity() ConnectorEntity {
	if a.ent == nil {
		var input ArithmeticInputs

		switch v := a.secondaryInput.(type) {
		case *Signal:
			input = ArithmeticInputs{
				FirstSignal:  a.primaryInput,
				SecondSignal: v,
			}
		case Constant:
			input = ArithmeticInputs{
				FirstSignal: a.primaryInput,
				Constant:    v,
			}
		default:
			panic(fmt.Errorf("invalid secondary input for arithmetic: %s", inspector.Inspect(a)))
		}

		a.ent = &ArithmeticCombinatorEntity{
			ControlBehavior: ArithmeticControlBehavior{
				ArithmeticConditions: ArithmeticConditions{
					ArithmeticInputs: input,
					Operation:        a.Operator,
					OutputSignal:     a.output,
				},
			},
		}

		if a.ent.Number() == debugEntityId {
			fmt.Println("debug ent", a.InputComponent.inputRed)
			debugNet = a.InputComponent.inputRed.Network
		}

		a.primaryInput = nil
		a.secondaryInput = nil
		a.output = nil
	}

	return a.ent
}

func (*Arithmetic) isComponent() {}

// func (n *Network) nextSignal() *Signal {
// 	n.signalNum++

// 	return signals[n.signalNum]
// }

func (n *Network) connectOutput(e ComponentOutput) {
	e.setOutput(n.isGreen, n.createConnector(e, false))
}

func (n *Network) connectInput(e ComponentInput) {
	e.setInput(n.isGreen, n.createConnector(e, true))
}

func (n *Network) createConnector(e Component, isInput bool) (conn *Connector) {
	conn = &Connector{n, e, !isInput}
	n.members = append(n.members, conn)
	return
}

// type Arithmetic struct {
// 	// Inputs
// }

func (Arithmetic) hasOutput() bool {
	return true
}

type Decider struct {
	InputComponent
	OutputComponent
	Operator    Comparator
	outputFixed bool
	output      *Signal
	ent         *DeciderCombinatorEntity
}

func (d *Decider) Entity() ConnectorEntity {
	if d.ent == nil {
		var cond BooleanCondition

		switch v := d.secondaryInput.(type) {
		case *Signal:
			cond = BooleanCondition{
				FirstSignal:  d.primaryInput,
				SecondSignal: v,
				Comparator:   d.Operator,
			}
		case Constant:
			cond = BooleanCondition{
				FirstSignal: d.primaryInput,
				Constant:    v,
				Comparator:  d.Operator,
			}
		default:
			panic(fmt.Sprintln("invalid secondary input for arithmetic: %s", inspector.Inspect(d)))
		}

		d.ent = &DeciderCombinatorEntity{
			ControlBehavior: DeciderControlBehavior{
				DeciderConditions: DeciderConditions{
					BooleanCondition:   cond,
					CopyCountFromInput: !d.outputFixed,
					OutputSignal:       d.output,
				},
			},
		}

		d.secondaryInput = nil
	}

	return d.ent
}

type Builder struct {
	networks []*Network
	cells    map[*generator.Variable]*Cell
	tickers  []Ticker
}

type block struct {
	*Builder
	scope *generator.Scope
}

// func (b *Builder) addEntity(e Entity) int {
// 	b.entities = append(b.entities, e)
// 	be := e.basic()
// 	be.EntityNumber = len(b.entities)

// 	return be.EntityNumber
// }

// type InitPulse struct {}

type Cell struct {
	id                 Constant
	get, set, sto, tmp *Decider
	net                *Network
}

// func (b *Builder) nextSig() *Signal {
// 	// b.signalNum++
// 	// b.signalNum %= 26
// }

func (b *Builder) createNet(isGreen bool) (net *Network) {
	net = &Network{
		isGreen: isGreen,
	}
	b.networks = append(b.networks, net)
	return
}

func (b *Builder) createSubnet(parent *Network) (subnet *Network) {
	subnet = b.createNet(!parent.isGreen)
	// subnet.signalNum = parent.signalNum
	return
}

// type Step struct {
// 	tick *Arithmetic
// }

func (d *Decider) addInputs(nets ...*Network) *Decider {
	for _, net := range nets {
		net.connectInput(d)
	}

	return d
}

func (d *Decider) addOutputs(nets ...*Network) *Decider {
	for _, net := range nets {
		net.connectOutput(d)
	}

	return d
}

func (b *Builder) createCell(net *Network, v *generator.Variable) (c *Cell) {
	var (
		setNet = b.createSubnet(net)
		stoNet = b.createNet(false)
		getNet = b.createSubnet(net)
	)

	id := net.nextCellID()

	c = &Cell{
		id:  id,
		net: net,

		get: (&Decider{
			InputComponent: InputComponent{
				primaryInput:   SignalG,
				secondaryInput: id,
			},
			output:   SignalV,
			Operator: ComparatorEq,
		}).addInputs(net, getNet).addOutputs(net),
		set: (&Decider{
			InputComponent: InputComponent{
				primaryInput:   SignalS,
				secondaryInput: id,
			},
			output:   signalEverything,
			Operator: ComparatorEq,
		}).addInputs(net).addOutputs(setNet),
		sto: (&Decider{
			InputComponent: InputComponent{
				primaryInput:   SignalS,
				secondaryInput: Constant(0),
			},
			output:   SignalV,
			Operator: ComparatorEq,
		}).addInputs(stoNet, setNet).addOutputs(stoNet, getNet),
		tmp: (&Decider{
			InputComponent: InputComponent{
				primaryInput:   SignalS,
				secondaryInput: Constant(0),
			},
			output:   SignalV,
			Operator: ComparatorNe,
		}).addInputs(setNet).addOutputs(setNet),
	}

	return
}

// type Step{}
//
// func (b *Builder) getTickArithmetic(in *Network,
type Ticker struct {
	*Arithmetic
	isMemOp bool
}

func (b *Builder) getTickerNetwork(tick int, isGreen, excludeMemOp bool) *Network {
	t := b.getTicker(tick, excludeMemOp)

	if t == nil {
		return nil
	}

	if isGreen {
		if t.outputGreen == nil {
			b.createNet(true).connectOutput(t)
		}

		return t.outputGreen.Network
	}

	if t.outputRed == nil {
		b.createNet(false).connectOutput(t)
	}

	return t.outputRed.Network
}

// func (b *Builder) connectToTicker(i ComponentInput, tick int, isGreen bool) {
// 	t := b.getTicker(tick)

// 	b.getTickerNetwork(t, isGreen).
// }

func (b *Builder) addControlTick(isGreen bool) (newtick int, end, next *Network) {
	end = b.getTickerNetwork(len(b.tickers)-1, isGreen, false)

	ticker := Ticker{
		Arithmetic: &Arithmetic{
			InputComponent: createInputs(SignalCheck, Constant(0)),
			output:         SignalCheck,
			Operator:       ArithmeticOperationOr,
		},
		isMemOp: true,
	}
	b.tickers = append(b.tickers, ticker)
	newtick = len(b.tickers) - 1
	// don't connect input
	next = b.createNet(isGreen)
	next.connectInput(ticker)
	return
}

func (b *Builder) getTicker(tick int, excludeMemOp bool) (t *Ticker) {
	if tick < len(b.tickers) {
		if excludeMemOp {
			if b.tickers[tick].isMemOp {
				return nil
			}
		}

		return &b.tickers[tick]
	}

	if tick >= len(b.tickers) {
		new := make([]Ticker, tick-len(b.tickers)+1)

		oldLen := len(b.tickers)
		b.tickers = append(b.tickers, new...)

		for i := range new {
			n := i + oldLen
			t = &b.tickers[n]

			t.Arithmetic = &Arithmetic{
				InputComponent: createInputs(SignalCheck, Constant(0)),
				output:         SignalCheck,
				Operator:       ArithmeticOperationOr,
			}

			if n > 0 {
				b.getTickerNetwork(n-1, false, false).connectInput(t)
			}
		}

	}

	if excludeMemOp {
		t.isMemOp = true
	}

	return
}

func (b *Builder) nextSafeMemOpTick(tick int, isRead bool) int {
	for ; ; tick++ {
		t := b.getTicker(tick, true)

		if t != nil {
			t2 := b.getTicker(tick+1, true)

			if t2 != nil {
				if !isRead {
					t3 := b.getTicker(tick+2, true)

					if t3 == nil {

						continue
					}

					t3.isMemOp = true
				}

				t.isMemOp = true
				t2.isMemOp = true

				return tick
			}

		}
	}
}

// takes 3 ticks
func (b *Builder) readCell(tick int, c *Cell, tr *Signal, onet *Network) int {
	tick = b.nextSafeMemOpTick(tick, true)
	dnet := b.getTickerNetwork(tick, false, false)
	rnet := b.getTickerNetwork(tick+2, !c.net.isGreen, false)

	disp := &Arithmetic{
		InputComponent: InputComponent{
			primaryInput:   SignalCheck,
			secondaryInput: Constant(c.id),
		},
		output:   SignalG,
		Operator: ArithmeticOperationMultiply,
	}
	dnet.connectInput(disp)

	c.net.connectOutput(disp)

	read := &Arithmetic{
		InputComponent: InputComponent{
			primaryInput:   SignalCheck,
			secondaryInput: SignalV,
		},
		output:   tr,
		Operator: ArithmeticOperationMultiply,
	}

	c.net.connectInput(read)
	rnet.connectInput(read)
	onet.connectOutput(read)

	return tick + 3
	// return b.getTicker(t2.net, pt.net.isGreen, )
}

func createInputs(primary *Signal, secondary Input) InputComponent {
	return InputComponent{primaryInput: primary, secondaryInput: secondary}
}

func (b *Builder) setCell(originTick int, net *Network, sig *Signal, c *Cell) int {

	tick := b.nextSafeMemOpTick(originTick, false)
	tn := b.getTickerNetwork(tick, !net.isGreen, false)

	if originTick != tick {
		cnet := b.createNet(net.isGreen)
		b.preserveUntil(net, cnet, sig, originTick, tick)
		net = cnet
	}

	disp := &Arithmetic{
		InputComponent: createInputs(SignalCheck, c.id),
		output:         SignalS,
		Operator:       ArithmeticOperationMultiply,
	}

	tn.connectInput(disp)
	c.net.connectOutput(disp)

	wr := &Arithmetic{
		InputComponent: createInputs(SignalCheck, sig),
		output:         SignalV,
		Operator:       ArithmeticOperationMultiply,
	}

	net.connectInput(wr)
	tn.connectInput(wr)
	c.net.connectOutput(wr)

	return tick + 1
}

func (b *Builder) addArithmetic(cond ArithmeticConditions) (ent *ArithmeticCombinatorEntity) {
	ent = &ArithmeticCombinatorEntity{
		ControlBehavior: ArithmeticControlBehavior{cond},
	}

	// b.addEntity(ent)

	return
}

// func (b *Builder) addDecider(cond DeciderConditions) (ent *DeciderCombinatorEntity) {
// 	ent = &DeciderCombinatorEntity{
// 		ControlBehavior: DeciderControlBehavior{cond},
// 	}

// 	b.addEntity(ent)

// 	return
// }

// func (b *Builder) declareVariable(decl *generator.Declare) {
// 	cell := b.createCell()
// 	// decl.InitialValue.(generator.ConstantValue).(int32)
// }

func (b *Builder) preserveUntil(net *Network, into *Network, sig *Signal, start, end int) {
	var ar *Arithmetic
	for tick := start; tick < end; tick++ {
		ar = &Arithmetic{
			InputComponent: createInputs(sig, SignalCheck),
			output:         sig,
			Operator:       ArithmeticOperationMultiply,
		}

		tnet := b.getTickerNetwork(tick, !net.isGreen, false)

		tnet.connectInput(ar)
		net.connectInput(ar)

		if tick+1 < end {
			net = b.createNet(false)
			net.connectOutput(ar)
		}
	}

	into.connectOutput(ar)
}

func (b *block) extractValue(
	typ generator.Typed,
	tick int,
	net *Network,
	sig *Signal,
) int {

	switch v := typ.(type) {
	case generator.UnaryOperation:
		ar := &Arithmetic{}
		switch v.Operator {
		case lexer.ADD:
			return b.extractValue(v.Operand, tick, net, sig)
		case lexer.SUB:
			subnet := b.createSubnet(net)
			subnet.connectInput(ar)
			ar.primaryInput = SignalS
			ar.secondaryInput = Constant(-1)
			ar.output = sig
			net.connectOutput(ar)
			ar.Operator = ArithmeticOperationMultiply

			return b.extractValue(v.Operand, tick, subnet, SignalS) + 1
		case lexer.TILDE:
			subnet := b.createSubnet(net)
			subnet.connectInput(ar)
			ar.primaryInput = SignalT
			ar.secondaryInput = Constant(-1)
			ar.Operator = ArithmeticOperationXor

			return b.extractValue(v.Operand, tick, subnet, SignalT) + 1
		}
	case generator.BinaryOperation:
		// sn := b.createSubnet(net)
		var sn *Network
		snl := b.createSubnet(net)
		snr := b.createSubnet(net)

		left := b.extractValue(v.Left, tick, snl, SignalL)
		right := b.extractValue(v.Right, tick, snr, SignalR)

		fmt.Println("binary", left, right)

		if left < right {
			tick = right
			sn = snr
			b.preserveUntil(snl, sn, SignalL, left, tick)
			// sn.merge(snl)
		} else if right < left {
			tick = left
			sn = snl
			b.preserveUntil(snr, sn, SignalR, right, tick)
			// sn.merge(snl)
		} else {
			tick = right
			sn = snr
			sn.merge(snl)
		}

		ic := createInputs(SignalL, SignalR)

		switch v.Operator {
		case lexer.ADD:
			ar := &Arithmetic{
				InputComponent: ic,
				output:         sig,
				Operator:       ArithmeticOperationAdd,
			}
			net.connectOutput(ar)
			sn.connectInput(ar)
			return tick + 1
		case lexer.SUB:
			ar := &Arithmetic{
				InputComponent: ic,
				output:         sig,
				Operator:       ArithmeticOperationSubtract,
			}
			net.connectOutput(ar)
			sn.connectInput(ar)
			return tick + 1
		case lexer.MUL:
			ar := &Arithmetic{
				InputComponent: ic,
				output:         sig,
				Operator:       ArithmeticOperationMultiply,
			}
			net.connectOutput(ar)
			sn.connectInput(ar)

			return tick + 1
		}
	case generator.ConstantValue:
		b := &Emitter{}
		b.items = []ConstantCombinatorItem{{
			signal: sig,
			count:  int32(v.Value().(uint64)),
		}}

		net.connectOutput(b)

		return tick
	case *generator.Variable:
		cell := b.cells[v]

		// TODO: optimize for multiple reads in the same expression 
		return b.readCell(tick, cell, sig, net)
	}

	panic(fmt.Errorf("unsupported value: %s", inspector.Inspect(typ)))
}

type CellAssignment struct {
	cell *Cell
	val  generator.Typed
}

func (b *Builder) createBlock(scope *generator.Scope) *block {
	return &block{
		Builder: b,
		scope:   scope,
	}
}

func (b *block) execStep(st generator.Step, net *Network, tick int) int {

	switch s := st.(type) {
	case *generator.Declare:
		cell := b.createCell(net, s.Variable)

		if s.Variable.InitialValue != nil {
			subnet := b.createNet(true)
			tick = b.extractValue(s.InitialValue, tick, subnet, SignalI)
			tick = b.setCell(tick, net, SignalI, cell)
		}
	case generator.Assign:
		v := b.scope.Lookup(s.Target).(*generator.Variable)
		cell := b.cells[v]

		subnet := b.createNet(true)
		tick = b.extractValue(s.Value, tick, subnet, SignalA)
		tick = b.setCell(tick, subnet, SignalA, cell)
	case generator.Block: // anonymous block
		cb := b.createBlock(s.Scope)
		for _, v := range s.Steps {
			fmt.Println("cstep")
			tick = cb.execStep(v, net, tick)
		}
	case generator.If:
		var blockStart, afterBlock *Network
		condNet := b.createNet(true)
		tick = b.extractValue(s.Condition, tick, condNet, SignalC)
		// ev := b.addArithmetic(ArithmeticConditions{
		// 	ArithmeticInputs: ArithmeticInputs{
		// 		FirstSignal: c,
		// 	},
		// })
		tnet := b.getTickerNetwork(tick, !condNet.isGreen, false)

		tick, _, blockStart = b.addControlTick(!condNet.isGreen)
		ev := &Decider{
			InputComponent: createInputs(SignalC, Constant(0)),
			Operator:       ComparatorNe,
			output:         SignalCheck,
		}
		evo := &Decider{
			InputComponent: createInputs(SignalC, Constant(0)),
			Operator:       ComparatorEq,
			output:         SignalCheck,
		}
		condNet.connectInput(ev)
		condNet.connectInput(evo)
		tnet.connectInput(ev)
		tnet.connectInput(evo)
		blockStart.connectOutput(ev)
		tick = b.execStep(s.Then, net, tick)

		tick, _, afterBlock = b.addControlTick(!condNet.isGreen)

		afterBlock.connectOutput(evo)
		b.getTickerNetwork(tick-1, condNet.isGreen, false).connectInput(b.tickers[tick])

		// tick, ec, sb := b.addControlTick(true)
	// case *generator.Return

	default:
		inspector.Println(st)
		panic("unhandled step")
	}

	return tick
}

func CreateBlueprint(m generator.Module, out io.Writer) error {
	var (
		b = Builder{
			cells: map[*generator.Variable]*Cell{},
		}

		initSteps = make([]CellAssignment, 0, len(m.Declarations))
	)

	bl := b.createBlock(m.Scope)

	bl.getTickerNetwork(1, false, false)

	fmt.Println(len(b.networks))
	pnet := b.createNet(false)

	for _, dec := range m.Declarations {
		if dec.Type().Kind() != generator.KindInt32 {
			return fmt.Errorf("invalid type: %s", dec.Type().Name())
		}

		cell := b.createCell(pnet, dec.Variable)

		bl.cells[dec.Variable] = cell

		if dec.InitialValue != nil {
			initSteps = append(initSteps, CellAssignment{cell, dec.InitialValue})
		}
	}

	var tick int

	for _, s := range initSteps {
		net := b.createNet(true)
		// might have issues if we read before setting...
		tick = bl.extractValue(s.val, tick, net, SignalI)

		tick = bl.setCell(tick, net, SignalI, s.cell)
	}

	// // inspector.Println(m.Lookup("main"))
	// //
	// for _, s := range initSteps {
	// 	bl.execStep(s, pnet, tick)
	// }

	// bl.execStep(generator.Call{
	// 	Target: ,
	// }, pnet, tick)
	//
	main := m.Lookup("main").(*generator.Function)
	cbl := b.createBlock(main.Scope)
	for _, v := range main.Steps {
		tick = cbl.execStep(v, pnet, tick)
	}

	p := Planner{}

	p.plan(b.networks)

	p.Save(out)

	return nil
}
