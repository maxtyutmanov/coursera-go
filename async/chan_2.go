package main

import (
	"fmt"
)

func main() {
	in := make(chan int)

	go func(out chan<- int) {
		for i := 0; i <= 10; i++ {
			fmt.Println("before", i)
			out <- i
			fmt.Println("after", i)
		}
		close(out)
		out <- 100
		fmt.Println("generator finish")
	}(in)

	for i := range in {
		fmt.Println("\tget", i)
	}
	val, closed := <-in
	fmt.Println("\tanother val", val)
	fmt.Println("\tis closed", closed)

	fmt.Scanln()
}
