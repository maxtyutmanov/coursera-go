package main

import "fmt"

func main() {
	outChans := make([]chan interface{}, 2)
	outChans[0] = make(chan interface{})
	outChans[1] = make(chan interface{})

	go workerFirst(outChans[0])
	go workerSecond(outChans[1])

	res1, ok := (<-outChans[0]).(int)
	if !ok {
		panic("FUCK!")
	}
	res2, ok := (<-outChans[1]).(int)
	if !ok {
		panic("FUCK2!")
	}

	fmt.Println("res1", res1)
	fmt.Println("res2", res2)

	fmt.Scanln()
}

func workerFirst(out chan interface{}) {
	out <- 1
}

func workerSecond(out chan interface{}) {
	out <- 2
}