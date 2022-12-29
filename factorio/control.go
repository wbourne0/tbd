package factorio

import (
	"encoding/json"
	"errors"
)

type Comparator uint8

const (
	ComparatorEq = iota
	ComparatorNe

	ComparatorLt
	ComparatorLte
	ComparatorGt
	ComparatorGte
)

var comparatorMap = [...][]byte{
	[]byte(`"="`),
	[]byte(`"≠"`),
	[]byte(`"<"`),
	[]byte(`"≤"`),
	[]byte(`">"`),
	[]byte(`"≥"`),
}

func (c Comparator) MarshalJSON() ([]byte, error) {
	buf := comparatorMap[c]

	if buf != nil {
		return buf, nil
	}

	return nil, errors.New("invalid comparator")
}

type ArithmeticOperation uint8

const (
	ArithmeticOperationAdd = iota
	ArithmeticOperationSubtract
	ArithmeticOperationMultiply
	ArithmeticOperationDivide
	ArithmeticOperationModulus
	ArithmeticOperationPow
	ArithmeticOperationLeftShift
	ArithmeticOperationRightShift
	ArithmeticOperationAnd
	ArithmeticOperationOr
	ArithmeticOperationXor
)

var arithmeticOperationMap = [...][]byte{
	[]byte(`"+"`),
	[]byte(`"-"`),
	[]byte(`"*"`),
	[]byte(`"/"`),
	[]byte(`"%"`),
	[]byte(`"^"`),
	[]byte(`"<<"`),
	[]byte(`">>"`),
	[]byte(`"AND"`),
	[]byte(`"OR"`),
	[]byte(`"XOR"`),
}

func (a ArithmeticOperation) MarshalJSON() ([]byte, error) {
	buf := arithmeticOperationMap[a]

	if buf != nil {
		return buf, nil
	}

	return nil, errors.New("invalid comparator")
}

type BooleanConditionSignal struct {
	FirstSignal  *Signal    `json:"first_signal"`
	SecondSignal *Signal    `json:"second_signal"`
	Comparator   Comparator `json:"comparator"`
}

func (BooleanConditionSignal) isBooleanCondition() {}

type BooleanCondition struct {
	FirstSignal  *Signal    `json:"first_signal"`
	Constant     Constant   `json:"constant,omitempty"`
	SecondSignal *Signal    `json:"second_signal,omitempty"`
	Comparator   Comparator `json:"comparator"`
}

// func (BooleanConditionConstant) isBooleanCondition() {}

type DeciderConditions struct {
	BooleanCondition
	OutputSignal       *Signal `json:"output_signal"`
	CopyCountFromInput bool    `json:"copy_count_from_input"`
}

type DeciderControlBehavior struct {
	DeciderConditions DeciderConditions `json:"decider_conditions"`
}

type ArithmeticInputsSignals struct {
	FirstSignal  *Signal `json:"first_signal"`
	SecondSignal *Signal `json:"second_signal"`
}

func (ArithmeticInputsSignals) isArithmeticSignals() {}

type SecondaryInput interface {
	isSecondaryInput()
}

type ArithmeticInputs struct {
	FirstSignal  *Signal `json:"first_signal"`
	SecondSignal *Signal `json:"second_signal,omitempty"`

	Constant Constant `json:"constant,omitempty"`
}

// func (ArithmeticInputsConstant) isArithmeticSignals() {}

// type ArithmeticInputs struct {
// 	Primary *Signal `json:"first_signal"`
// }

type ArithmeticConditions struct {
	ArithmeticInputs
	Operation    ArithmeticOperation `json:"operation"`
	OutputSignal *Signal             `json:"output_signal"`
}

type ArithmeticControlBehavior struct {
	ArithmeticConditions ArithmeticConditions `json:"arithmetic_conditions"`
}

type ConstantFilter struct {
	Signal Signal `json:"signal"`
	Count  int    `json:"count"`
	Index  int    `json:"index"`
}

type ConstantControlBehavior struct {
	Filters []ConstantFilter `json:"filters"`
}

type LampControlBehavior struct {
	CircuitCondition BooleanCondition `json:"circuit_condition"`
	UseColors        bool             `json:"use_colors"`
}

type ConnectionEntity struct{ Entity }

func (c ConnectionEntity) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Number())
}

type Connection struct {
	EntityID  int `json:"entity_id"`
	CircuitID int `json:"circuit_id,omitempty"`
}

type Connector struct {
	*Network
	Component   Component
	isSecondary bool
}

type BlueprintIcon struct {
	Signal *Signal `json:"signal"`
	Index  int     `json:"index"`
}

type BlueprintData struct {
	Icons    []BlueprintIcon `json:"icons"`
	Entities []Entity        `json:"entities"`
	Item     string          `json:"item"`
	Version  int             `json:"version"`
}

type Blueprint struct {
	BlueprintData `json:"blueprint"`
}
