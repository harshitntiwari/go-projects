package main

import (
	"context"
	"fmt"
	"sync"
)

type Job struct {
	ID int
	Payload string
}

type Result struct {
	JobID int
	Output string
	Err error
}

type WorkerPool struct {
	maxWorkers int
	jobs chan Job
	results chan Result
	wg sync.WaitGroup
}

// NewWorkerPool allocates and returns a new [WorkerPool]
func NewWorkerPool(maxWorkers int, jobBuffSize int, resultBuffSize int) *WorkerPool {
	return &WorkerPool {
		maxWorkers: maxWorkers,
		jobs: make(chan Job, jobBuffSize),
		results: make(chan Result, resultBuffSize),
	}
}

// Starts wp.maxWorkers number of worker go routines
func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < wp.maxWorkers; i++ {
		wp.wg.Add(1)
		go func(id int) {
			defer wp.wg.Done()
			wp.runWorker(ctx, id)
		}(i)
	}

	go func() {
		wp.wg.Wait()
		// close the results channel when all the workers are done pushing results
		close(wp.results)
	}()
}

func (wp *WorkerPool) runWorker(ctx context.Context, id int) {
	// we run an infinite loop because we want to keep receiving from the jobs channel
	for {
		select {
		case <- ctx.Done() : {
			fmt.Printf("worker %d shutting down: %v\n", id, ctx.Err())
			return
		}
		case job, ok := <-wp.jobs : {
			// if jobs channel is closed, ok will be == false
			if !ok {
				return
			}
			// process the job
			result := wp.process(id, job)
			wp.results <- result
		}
		}
	}
}

func (wp * WorkerPool) process(workedId int, job Job) (result Result) {
	
	// this defer func will be called if process() panics (possibly because doWork() panics)
	defer func () {
		if r := recover(); r != nil {
			result = Result{
				JobID: job.ID,
				Err: fmt.Errorf("worker %d panic on job %d: %v", workedId, job.ID, r),
			}
		}
	}()

	// process the job
	if err := doWork(job); err != nil {
		return Result{JobID: job.ID, Err: err}
	}

	return Result{JobID: job.ID, Output: fmt.Sprintf("processed: %s", job.Payload)}
}

func doWork(job Job) error {

	// some jobs fail
	if job.ID % 7 == 0 {
		return fmt.Errorf("error processing job %d", job.ID)
	}
	return nil
}

// Returns read-only results channel
func (wp *WorkerPool) Result() <-chan Result{
	return wp.results
}

func (wp *WorkerPool) Submit(ctx context.Context, job Job) error {
	select {
	case <- ctx.Done(): {
		return fmt.Errorf("pool shutting down: %v\n", ctx.Err())
	}
	case wp.jobs <- job :{
		return nil
	}
	}
}

// Closes the jobs channel and waits for workers to finish
func (wp *WorkerPool) Shutdown() {
	close(wp.jobs)
	wp.wg.Wait()
}