package main

import (
	"fmt"
	"go/token"
	"io/ioutil"
	"main/factorio"
	"main/generator"
	"main/parser"
	"os"
)


const hm = 0x1.2p+1

func main() {
	content, _ := ioutil.ReadFile("./test2.tbd")
	file := token.NewFileSet().AddFile("blah", 1, len(content))

	p := parser.NewParser(content, file)

	mod := p.ParseModule()
	
	m := generator.ProcessModule(mod)
	// fmt.Println(inspector.Inspect(a))

	if len(m.Errors) > 0 {
		for _, err := range m.Errors {
			fmt.Println(err.Format(file))
		}

		return
	}

	// fmt.Println(stringify(m))
	// 
	

	f, err := os.Create("out.bp")

	if err != nil {
		panic(err)
	}
	
	fmt.Println("wat")



// pl.DoLargeBuild()
// 
	if err := factorio.CreateBlueprint(m, f); err != nil {
		panic(err)
	}

// pl.Save(f)

	// factorio.DoLargeBuild(f)	// if err := factorio.CreateBlueprint(m, f); err != nil {
	// 	panic(err)
	// }

	// 	for _, sample := range [...]string{
	// 		`func main() {t = --a << 1
	// 	&^ 32
	// 	^ b++
	// 	|( ~c
	// 	+ myVar)
	// 	- -3 / a + a
	// 		.bc(
	// 			32
	// 			& 3 | 15
	// 		).dc([...]int{3*15})
	// 	* 15

	// 	t &^= ~++t | 3
	// 	}`,
	// 		"func main(){d = [][...][5 * 2][]int{1,2}}",
	// 		"func main(){abc = [...]int{3 * 2: 1, 2 % 2: 2}}",
	// 		"func main(){*(abc)(def) = 3}",
	// 		"func main(){abc*=func1();def = func2();abc.def()}",
	// 		`func main(){a="abcd"; a+="efgh"}`,
	// 	} {
	// 		fmt.Println(inspector.Inspect(getParser(sample).ParseModule()))
	// 	}

	// 	fmt.Println(inspector.Inspect(getParser("func a(a, b int) {a *= b; otherFunc(a); a++ }").ParseModule()))
	// 	fmt.Println(inspector.Inspect(getParser(`
	// func a(b, c int, o myStruct) {
	// 	b |= c << 1 & ~b
	// 	o.doSmthn(&b)
	// 	return b ^ c
	// }

	// func b(d, e int, o string) {
	// 	d *= -e
	// 	var h int
	// 	var l int = a * 3 - 2 / 5 % 15

	// 	return o * d
	// }

	// func myStruct.c(i int) {
	// 	return i
	// }

	// func *myStruct.c(i int) {
	// 	return i
	// }

	// const i = 3 ^ 2

	// 	`).ParseModule()))

	// 	fmt.Println(inspector.Inspect(getParser(`

	// 	func a () {
	// 		d := 32
	// 		return a
	// 	.b(3 / (10 * 2))
	// 	.c(83 / ~15)
	// 	  * d[32 / (x + 5)]}`).ParseModule()))
}
