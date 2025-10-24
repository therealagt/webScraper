package scraper

import (
	"context"
	"fmt"
	"time"
)

/*  workers with routines and channels */
func worker(id int, jobs <-chan int, results chan<- int) {
    for j := range jobs {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		done := make(chan struct{})
		go func ()  {
			fmt.Println("worker", id, "started  job", j)
			Task(j)
			fmt.Println("worker", id, "finished job", j)
			results <- j * 2
			close(done)
		}()

		select {
		case <-done:

		case <-ctx.Done():
			fmt.Println("worker", id, "timeout for job", j)
		}
		cancel()
    }
}

/* start workers, deal jobs and collect results */
func StartWorkerPool(numWorkers, numJobs int) {
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	for w := 1; w <= numWorkers; w++ {
		go worker(w, jobs, results)
	}

	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	for a := 1; a <= numJobs; a++ {
		res := <-results
		fmt.Println("Results", res)
	}

	fmt.Println("All jobs done. Shutting down worker pool")
}
