package optimizer

import "main/generator"

type Assembler struct {
	mod   generator.Module
	steps []Step
}

type Step interface{ isStep() }

type Define struct {
	Size int
}

func (a Define) isStep() {}

// func (a Assembler) AssembleMain() {
// 	for _, dec := range a.mod.Declarations {

// 		dec.(generator.Declare)
// 	}
// }

func (a Assembler) Declare(stp generator.Declare) {

}

// Steps are:
// Define     target [value]
// Assign     target value
// Compare    left right
// While      cond

// Too tired for this :/
// The end result of the current test.tbd file should be something like:
// declare a 42
// while true:
// assign a a + 532

// TODO: Implement a way to provide an API for the target runtime.
// This can be done through standard libraries (for example, print could be translated to `console.log` in a js runtime or to a syscall for llvm)
// Or through runtime-specific items, say `nodejs.console.log` or similar.

// A mix of both is ideal IMO, standard functions such as slice and array manipulation should be runtime-defined but under a global namespace,
// but there should also be an api for runtime-dependent functions.
