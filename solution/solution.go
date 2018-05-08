package main

import "fmt"

func main() {

	jobs := make(chan int, 100000)
	results := make(chan int, 100000)
	sum := make(chan int)

	go Stage1(jobs)

	go Stage2(jobs, results)

	go Stage3(results, sum)

	fmt.Println(<-sum)

}

func Stage3(results chan int, sum chan int) {

	var total int

	for result := range results {
		total += result
	}

	sum <- total
}

func Stage2(jobs chan int, results chan int) {

	for job := range jobs {
		switch job % 3 {
		case 0:
			results <- job * 1
		case 1:
			results <- job * 2
		case 2:
			results <- job * 3
		}
	}

	close(results)
}

func Stage1(jobs chan int) {

	for i := 1; i <= 1000000000; i++ {

		if i%2 == 0 {
			i += 99
		}

		jobs <- i
	}

	close(jobs)
}
