package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fithanso/go-kata/01-context-cancellation-concurrency/05-context-aware-error-propagator/auth"
	"github.com/fithanso/go-kata/01-context-cancellation-concurrency/05-context-aware-error-propagator/meta"
	"github.com/fithanso/go-kata/01-context-cancellation-concurrency/05-context-aware-error-propagator/storage"
)

type FileGateway struct {
	as *auth.AuthService
	ms *meta.MetadataService
	ss *storage.StorageService
	uplAttempts int
	timeoutSec  int
}

func NewFileGateway(as *auth.AuthService, ms *meta.MetadataService, ss *storage.StorageService) *FileGateway {
	return &FileGateway{as, ms, ss, 3, 7}
}

func (fg *FileGateway) StartUpload() {
	
	for _ = range(fg.uplAttempts) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(fg.timeoutSec) * time.Second)
		err := fg.attemptUpload(ctx)
		
		if err != nil {
			cancel()
			aErr := &auth.AuthError{}
			if errors.As(err, aErr) {
				if !aErr.Temporary() {
					fmt.Println("auth not temp error")

					break
				}
				fmt.Println("auth error")
			}
			
			mErr := &meta.StorageQuotaError{}
			if errors.As(err, mErr) {
				if !mErr.Temporary() {
					fmt.Println("meta not temp error")
					break
				}
				fmt.Println("quota error")
			}
			
			bErr := &storage.BlobError{}
			if errors.As(err, bErr) {
				if !bErr.Temporary() {
					fmt.Println("blob not temp error")

					break
				}
				fmt.Println("blob error")
				
			}
			
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Println("deadline exceeded")
				break
			}
			
		} else {
			cancel()
			break
		}
		cancel()
	}
	

}

func (fg *FileGateway) attemptUpload(ctx context.Context) error {
	err := fg.as.Auth(ctx)
	
	if err != nil {
		return InfernalError{Err: err}
	}
	
	err = fg.ms.StoreMeta(ctx)
	if err != nil {
		return InfernalError{Err: err}
	}
	
	err = fg.ss.StoreBlob(ctx)
	if err != nil {
		return InfernalError{Err: err}
	}
	
	return nil
}

func main() {
	as := &auth.AuthService{Fail:false}
	ms := &meta.MetadataService{Fail:true}
	ss := &storage.StorageService{Fail:false}
	s := NewFileGateway(as, ms, ss)
	s.StartUpload()
}







