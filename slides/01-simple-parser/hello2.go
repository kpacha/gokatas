package main

import (
	"fmt"
	"os"
)

func main() {
	var name string                // declare a variable without instantiation
	argsWithoutProg := os.Args[1:] // declare a var but let the compiler infer the type
	if len(argsWithoutProg) > 0 {
		name = argsWithoutProg[0]
	} else {
		name = "World"
	}
	for i := 0; i < 10; i++ {
		greeter(name, i)
	}
}

func greeter(name string, i int) {
	fmt.Printf("#%d: Hello, %s!\n", i, name)
}
