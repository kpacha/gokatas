package main

import "fmt"

type Greeter struct {
	Name string
}

func (g Greeter) greet(name string) {
	fmt.Printf("%s: Hello, %s!\n", g.Name, name)
}

func main() {
	name := "World" // good enough for this example
	var greeter Greeter
	for i := 0; i < 10; i++ {
		greeter = Greeter{fmt.Sprintf("Greeter #%d", i)}
		greeter.greet(name)
	}
}
