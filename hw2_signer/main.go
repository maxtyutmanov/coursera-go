package main

import "fmt"

func main() {
	println("run as\n\ngo test -v -race")

	//inputData := []int{0, 1, 2, 3}
	 inputData := []int{0,1}
	var testResult int

	hashSignJobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(squareNumbers),
		job(sumNumbers),
		job(func(in, out chan interface{}) {
			data := <-in
			res, ok := data.(int)
			if !ok {
				panic("FUCK3!")
			}
			testResult = res
		}),
	}

	ExecutePipeline(hashSignJobs...)
	fmt.Println(testResult)
}

func squareNumbers(in, out chan interface{}) {
	for num := range in {
		numAs, ok := num.(int)
		if !ok {
			panic("FUCK!")
		}
		out <- numAs * numAs
	}
}

func sumNumbers(in, out chan interface{}) {
	var curSum int = 0
	for data := range in {
		num, ok := data.(int)
		if !ok {
			panic("FUCK2!")
		}
		curSum += num
	}
	out <- curSum
}