package main

import "testing"

func sum(a []int, c chan int) {
	total := 0
	for _, v := range a {
		total += v
	}
	c <- total // send total to c
}

type Person struct {
	name string
	address string
	age int
}

func TestConcurrency(t *testing.T) {
	t.Error("Testing!")
	t.Errorf("%v", Person{"Adam", "Garden of Eden", 32})
}
