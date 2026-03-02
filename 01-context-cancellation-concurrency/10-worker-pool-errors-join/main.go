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
		return fmt.Errorf("gavno error!!!")
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
	var errList  error
	var firstErr error
	
	p.wg.Add(p.n_workers)
	for i := 0; i < p.n_workers; i++ {
		go p.Worker(ctx, jobs, resultsCh)
	}
	
	go func() {
		p.wg.Wait()
		close(resultsCh)
	}()
	
	
	for {
		select {
			
			case <- ctx.Done():
			cancel()
			return errors.Join(errList, ctx.Err())
			
			case r, ok := <- resultsCh:
			if !ok {
				if firstErr != nil {
					
					return firstErr
				}
				if errList != nil {
					return errList
				}
				
				if parent.Err() != nil {
					return parent.Err()
				}
				return nil
			}
			
			if r.Error != nil {
				// thr main idea of fail-fast is to give all workers time to finish
				if p.stopOnFirstError {
					if firstErr == nil {
						firstErr = r.Error
						cancel()
						continue
					}
				}
				errList = errors.Join(errList, r.Error)
			}
			
		}
	}
		
}

func (p *Pool) Worker(ctx context.Context, jobsCh <- chan Job, resultsCh chan <- Result) {
	
	defer p.wg.Done()
	
	for j := range jobsCh {
		
		fmt.Println("working on job #", j.ID)
		
		select {
			case <- ctx.Done():
			resultsCh <- Result{Success: false, Error: ctx.Err()}
			return 
			default:
		}
		
		err := j.JobFunc(ctx)
		
		if err != nil {
			fmt.Println("error in job #", j.ID)
			resultsCh <- Result{Success: false, Error: err}
		}else{
			resultsCh <- Result{Success: true, Error: nil}
		}
		
	}
}



func main() {
	jobs := 5
	
	jobCh := make(chan Job)
	wg := sync.WaitGroup{}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	pool := NewPool(3, true)
	
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
	fmt.Println(err)
	
}