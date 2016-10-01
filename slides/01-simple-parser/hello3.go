package main

import (
	"fmt"
	"os"
)

func main() {
	name := "World"
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) > 0 {
		name = argsWithoutProg[0]
	}
	done := make(chan interface{})
	for i := 0; i < 10; i++ {
		go greeter(name, i, done)
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	fmt.Println("We're done!")
}

func greeter(name string, i int, done chan<- interface{}) {
	fmt.Printf("#%d: Hello, %s!\n", i, name)
	done <- nil
}
