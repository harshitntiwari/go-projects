package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)


func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	numJobs := 20
	wp := NewWorkerPool(5, 20, 20)

	wp.Start(ctx)
	
	// pushing jobs into the jobs channel in a separate goroutine, so that it doesn't block the main go routine.
	go func() {
		for i := range numJobs {
			job := Job{ID: i, Payload: fmt.Sprintf("task-%d", i)}
			if err := wp.Submit(ctx, job); err != nil {
				fmt.Println("Error during job submission: ", err)
				return
			}
		}
		wp.Shutdown()
	}()

	// extracting the results from the results channel.
	// it is upto the caller on how it wants to process the result
	var errCount int
	for result := range wp.Result() {
		if result.Err != nil {
			fmt.Printf("ERROR job %d: %s\n", result.JobID, result.Err)
		} else {
			fmt.Printf("OK    job %d: %s\n", result.JobID, result.Output)
			errCount++;
		}
	}

	fmt.Printf("\nDone. %d errors out of %d jobs", errCount, numJobs)
}