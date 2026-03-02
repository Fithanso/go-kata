package meta 

import (
	"context"
	"time"
	"fmt"
)


type MetadataService struct {
	Fail bool
}

func (ms *MetadataService) StoreMeta(ctx context.Context) error {
	fmt.Println("meta working...")

	time.Sleep(2 * time.Second)
	
	select {
		case  <- ctx.Done():
		return MetaDatabaseError{
			err: ctx.Err(),
			timeout: true,
			temporary: false,
		}
		default:
	}
	
	if ms.Fail {
		return MetaDatabaseError{
			err: StorageQuotaError{
				err: fmt.Errorf("quota exceeded"),
				timeout: false,
				temporary: false,
			},
			timeout: false,
			temporary: false,
		}
	}
	
	return nil
}
