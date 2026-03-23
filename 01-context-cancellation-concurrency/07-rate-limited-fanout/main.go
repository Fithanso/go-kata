package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"sync"

	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)


func MockRequest(shouldFail bool) error {
	if shouldFail {
		return fmt.Errorf("example error")
	}
	return nil
}

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
	// TODO: передавать горутинам каналы и http клиента
	ctx, cancel := context.WithCancel(parent) 
	defer cancel()
	
	client := http.Client{
		Transport: http.DefaultTransport,
		Timeout: 2 * time.Second,
	}
		
	
	results := make(map[int][]byte)
	var (
		wg sync.WaitGroup
	)
	
	resChan := make(chan FetchRes) // for worker - aggregator communication. contains result
	errChan := make(chan error, 1) // for worker - FetchAll communication. contains possible errors
	doneChan := make(chan error) // for aggregator - FetchAll communication
	
	go func() {
		// results aggregator
		for {
			select {
				case <- ctx.Done():
				return
				
				case res, ok := <- resChan:
				if !ok {
					// no more results - wg was awaited and resChan is closed
					close(doneChan)
					return
				}
				results[res.uid] = res.res
				
			}
		}
	}()
	
	
	
	for _, uid := range userIDs {
		uid := uid
		
		// check if context is not closed
		select {
			case <- ctx.Done():
			// exits immediately if any error occurred, because ctx is closed
			return nil, ctx.Err()
			
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
		go func(httpCli http.Client) {
			
		_ = httpCli
		defer foc.sem.Release(1)
		defer wg.Done()	
		
		select {
			case <- ctx.Done():
			return
			
			default:
		}
		time.Sleep(1 * time.Second)
		
		err := MockRequest(false)
		
		if err != nil {
			errChan <- fmt.Errorf("example error")
			return
		}
		resChan <- FetchRes{uid: uid, res: []byte("success")}

		}(client)
		
		
	}
	
	go func() {
		wg.Wait()
		close(resChan)
	}()
	
	select {
		case <- doneChan:
		return results, nil
		
		case err := <- errChan:
		cancel()
		return nil, err
	}
	
	
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
	
	uidsLen := 20
	// TODO: при десяти проблем нет. проблемы со 100 и если дропнуть ошибку. мб не возвращается в семафор или лимитер
	
	uids := make([]int, uidsLen)
	for i := 0; i < uidsLen; i++ {
		uids[i] = 100 + i
	}	

	// launch fetcher
	n1 := time.Now()
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
	
	fmt.Printf("Program took %d secs to run \n", time.Since(n1) / time.Second)
		
}