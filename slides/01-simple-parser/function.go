package main

import "fmt"

func main() {
	if _, err := NewPositiveInteger(-3); err != nil {
		fmt.Println(err.Error())
	}
	result, err := NewPositiveInteger(42)
	if err != nil {
		panic(err)
	}
	fmt.Println(result.Value)
}

type PositiveInteger struct {
	Value int
}

func NewPositiveInteger(i int) (*PositiveInteger, error) {
	if i < 0 {
		return nil, fmt.Errorf("%d is not a positive integer", i)
	} else {
		return &PositiveInteger{i}, nil
	}
}
