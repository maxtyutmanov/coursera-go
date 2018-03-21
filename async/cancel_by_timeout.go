package main

import "context"
import "math/rand"
import "time"
import "fmt"

func main() {
	in := make(chan int)
	ctx, _ := context.WithTimeout(context.Background(), time.Second * 4)

	for i := 1; i <= 10; i++ {
		go worker(ctx, i, in)
	}

	var totalFound int
LOOP:
	for {
		select {
		case foundBy := <-in:
			fmt.Println("Result was found by", foundBy)
			totalFound++
		case <- ctx.Done():
			break LOOP
		}
	}

	fmt.Println("Number of results: ", totalFound)
	fmt.Scanln()
}

func worker(ctx context.Context, ix int, out chan<- int) {
	ttw := time.Duration(rand.Intn(5)) * time.Second

	select {
	case <-ctx.Done():
		fmt.Println("The worker stops due to being canceled", ix)
		return
	case <-time.After(ttw):
		fmt.Println("The worker has finished", ix)
		out <- ix
	}
}