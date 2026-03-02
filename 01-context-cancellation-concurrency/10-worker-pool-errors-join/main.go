package main

import (
	"fmt"
	"sync"
	"errors"
	"time"
	"context"
)

func huyachim(ctx context.Context) error {
	select {
	case <-time.After(1 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
type Job struct {
	ID        int
	JobFunc   func(context.Context) error
}

type Result struct {
	Success  bool
	Error    error
}


type Pool struct {
	n_workers        int
	stopOnFirstError bool // collect all errors or fail-fast
	wg               sync.WaitGroup
}

func NewPool(n int, stopOnErr bool) *Pool {
	return &Pool{
		n,
		stopOnErr,
		sync.WaitGroup{},
	}
}

func (p *Pool) Run(parent context.Context, jobs <- chan Job) error {
	
	ctx, cancel := context.WithCancel(parent)
	resultsCh := make(chan Result)
	var errList error
	
	select {
		case <- ctx.Done():
		cancel()
		return errors.Join(errList, ctx.Err())
		
		case j := <- jobs:
		p.wg.Add(1)
		go p.Worker(ctx, j, resultsCh)
		
		case r := <- resultsCh:
		
		if r.Error != nil {
			
			if p.stopOnFirstError {
				cancel()
				return r.Error
			}
			errList = errors.Join(errList, r.Error)
		}
		
	}
	
	p.wg.Wait()
	cancel()

	return errList
	
}

func (p *Pool) Worker(ctx context.Context, j Job, resultsCh chan <- Result) {
	
	defer p.wg.Done()
	
	fmt.Printf("working on job #%d", j.ID)
	
	select {
		case <- ctx.Done():
		resultsCh <- Result{Success: false, Error: ctx.Err()}
		return 
		default:
	}
	err := j.JobFunc(ctx)
	if err != nil {
		resultsCh <- Result{Success: false, Error: err}
	}
	
	resultsCh <- Result{Success: true, Error: nil}

}



func main() {
	jobs := 5
	
	jobCh := make(chan Job)
	wg := sync.WaitGroup{}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	pool := NewPool(3, false)
	
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(jobCh)
		
		for i := 0; i < jobs; i++ {
			
			select {
				case <- ctx.Done():
				return
				case jobCh <- Job{ID: i, JobFunc: huyachim}:
			}
			
		}
	}()
	
	err := pool.Run(ctx, jobCh)
	if err != nil {
		cancel()
	}
	
	wg.Wait()
	fmt.Println("finished")	
	
}