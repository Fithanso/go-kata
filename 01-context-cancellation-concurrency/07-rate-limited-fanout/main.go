package main

import (
	"context"
	"fmt"
	"time"
	// "http"
	"sync"

	"golang.org/x/time/rate"
	"golang.org/x/sync/semaphore"
)

const (
	MAXINFLIGHT = 8
	RPS = 10
	BURST = 20
)

type FanOutClient struct {
	limiter     *rate.Limiter
	sem         *semaphore.Weighted
}

func NewFanOutClient (l *rate.Limiter, s *semaphore.Weighted) *FanOutClient {
	
	return &FanOutClient{
		limiter:     l,
		sem:         s,
	}
}

type FetchRes struct {
	uid int
	res []byte
}


func (foc *FanOutClient) FetchAll(parent context.Context, userIDs []int) (map[int][]byte, error) {
	ctx, cancel := context.WithCancel(parent) 
	defer cancel()
	
	results := make(map[int][]byte)
	var (
		wg sync.WaitGroup
	)
	
	resChan := make(chan FetchRes)
	errChan := make(chan error)
	doneChan := make(chan error)
	
	go func() {
		
		for {
			select {
				case <- ctx.Done():
				return
				
				case res := <- resChan:
				results[res.uid] = res.res
				
				case err := <- errChan:
				// return to the main FetchAll function, pass the error
				doneChan <- err
				return
			}
		}
	}()
	
	
	
	for _, uid := range userIDs {
		uid := uid
		
		// check if context is not closed
		select {
			case <- ctx.Done():
			return nil, ctx.Err()
			
			case fetchErr := <- doneChan:
			// exit immediately if any error occurred
			return nil, fetchErr
			
			default:
		}
		
		waitErr := foc.limiter.Wait(ctx)
		if waitErr != nil {
			return nil, waitErr
		}
		
		semErr := foc.sem.Acquire(ctx, 1)
		if semErr != nil {
			return nil, semErr
		}
		
		wg.Add(1)
		go func() {
		defer foc.sem.Release(1)
		defer wg.Done()	
		
		select {
			case <- ctx.Done():
			return
			
			case <- doneChan:
			return
			
			default:
		}
		//time.Sleep(2 * time.Second)
		resChan <- FetchRes{uid: uid, res: []byte("success")}
		
		}()
		
		
	}
	
	wg.Wait()
	return results, nil
}

// в метод приходят айдишки. из каждого id мы далем PendingJob. они хранятся в мапе клиента. клиент имеет: мапу с джобами, лимитер, семафор.
// метод FetchAll ждет завершения всех горутин либо первой ошибки. в конце очищаем поле с джобами
// при создании клиента передаем ему лимитер и семафор. проходим по списку id, для каждого id ждем лимитер, далее ждем семафор, 

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	lim := rate.NewLimiter(rate.Every(RPS * time.Second), BURST)
	sem := semaphore.NewWeighted(MAXINFLIGHT)
	foCli := NewFanOutClient(lim, sem)
	
	var (
		wg  sync.WaitGroup
		fres map[int][]byte
	)
	
	uidsLen := 100
	// TODO: при десяти проблем нет. проблемы со 100 и если дропнуть ошибку. мб не возвращается в семафор или лимитер
	
	uids := make([]int, uidsLen)
	for i := 0; i < uidsLen; i++ {
		uids[i] = 100 + i
	}	

	// launch fetcher
	wg.Go(func() {
		res, err := foCli.FetchAll(ctx, uids)
		if err != nil {
			fmt.Println(err)
			wg.Done()
			return
		}
		fres = res
	})
	
	wg.Wait()
	for uid, result := range fres {
		fmt.Println(uid, string(result))
	}
	fmt.Println(len(fres))
		
}