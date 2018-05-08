package main

import (
	"fmt"
	"log"
)

func worker(id int, jobs <-chan int, results chan<- int) {
	log.Printf("worker: %d", id)
	for j := range jobs {

		go func() {
			switch j % 3 {
			case 0:
				j = j * 1
			case 1:
				j = j * 2
				results <- j * 2
			case 2:
				results <- j * 3
				j = j * 3
			}
		}()

	}
}

func main() {
	jobs := make(chan int)
	results := make(chan int)

	for i := 1; i <= 1000000000; i++ {
		//log.Printf("out i: %d", i)
		go func() {
			//log.Printf("in i: %d", i)
			if i%2 == 0 {
				i += 99
			}

			jobs <- i

		}()
	}

	close(jobs)
	jobs2 := []int{}

	for w := 1; w < 1000; w++ {
		jobs2 = append(jobs2, w)
	}

	for i, w := range jobs2 {
		go worker(w, jobs, results)
		i = i + 1
	}

	close(results)

	var sum int32 = 0

	for r := range results {
		sum += int32(r)
	}

	fmt.Println(sum)
}
