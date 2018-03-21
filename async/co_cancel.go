package main

import "context"
import "math/rand"
import "time"
import "fmt"

func main() {
	in := make(chan int)
	ctx, finish := context.WithCancel(context.Background())
	for	i := 1; i <= 10; i++ {
		go worker(ctx, i, in)		
	}
	fmt.Println("Started all workers!")
	finishedId := <-in
	fmt.Println("The worker has finished first", finishedId)
	finish()
	fmt.Scanln()
}

func worker(ctx context.Context, ix int, out chan<- int) {
	ttw := time.Duration(rand.Intn(5))

	select {
	case <-ctx.Done():
		fmt.Println("The worker stops due to being canceled", ix)
		return
	case <-time.After(ttw * time.Second):
		fmt.Println("The worker has finished", ix)
		out <- ix
	}
}