package auth

import (
	"context"
	"fmt"
	"time"
)


type AuthService struct {
	Fail bool
}

func (as *AuthService) Auth(ctx context.Context) error {
	fmt.Println("auth working...")
	time.Sleep(2 * time.Second)
	
	select {
		case  <- ctx.Done():
		return AuthError{
			err: ctx.Err(),
			timeout: true,
			temporary: false,
			UserLogin: "",
		}
		default:
	}
	
	if as.Fail {
		return AuthError{
			err: fmt.Errorf("auth should have failed"),
			timeout: false,
			temporary: false,
			UserLogin: "",
		}
	}
	
	return nil
}
