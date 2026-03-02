package storage 

import (
	"context"
	"time"
	"fmt"
)

type StorageService struct {
	Fail bool
}

func (ss *StorageService) StoreBlob(ctx context.Context) error {
	fmt.Println("storage working...")

	time.Sleep(2 * time.Second)
	
	select {
		case  <- ctx.Done():
		return BlobError{
			err: ctx.Err(),
			timeout: true,
			temporary: false,
		}
		default:
	}
	
	if ss.Fail {
		return BlobError{
			err: fmt.Errorf("blob should have failed"),
			timeout: false,
			temporary: false,
		}
	}
	
	return nil
}
