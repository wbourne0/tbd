package main

// type FunctionStatement struct {
// 	BlockScope
// 	arguments []string
// }

// type IfStatement struct {
// 	initialization Step
// 	condition      Step
// 	root           *BlockScope
// 	block          *BlockScope
// 	elseifs        []*BlockScope
// 	// this is `else, it just can't be named as such as `else` is a reserved keyword.
// 	fallback *BlockScope
// }

// type Reference []string

// // func (a AggregateStep) append(steps ...Step) {
// // 	a.steps = append(a.steps, steps...)
// // }

// func parseIf(
// 	source []byte,
// 	index int,
// 	parent *BlockScope,
// ) {

// }

// func parseReference(source []byte, index int) (Reference, int, error) {
// 	var (
// 		words = make([]string, 0, 1)
// 		word  string
// 		err   error
// 		tmp   int
// 	)

// parseWord:
// 	index = skipWhitespace(source, index)

// 	if !isLetter(source[index]) && len(words) > 1 {
// 		return words, index, err
// 	}

// 	word, tmp, err = parseWord(source, index)
// 	words = append(words, word)

// 	if err != nil {
// 		return words, index, err
// 	}
// 	index = tmp

// 	index = skipWhitespace(source, index)

// 	if source[index+1] == '.' {
// 		index++
// 		goto parseWord
// 	}

// 	return words, index, err
// }
