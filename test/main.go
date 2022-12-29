package main

import (
	"main/factorio"
	"os"
)

var pl factorio.Planner

func main() {

	f, err := os.Create("./out.bp")
	defer f.Close()
	if err != nil {
		panic(err)
	}
	pl.DoLargeBuild()
	//
	//

	pl.Save(f)
	//
	// factorio.DoLargeBuild(f)
}
