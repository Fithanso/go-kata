package auth

import "fmt"


type AuthError struct {
	err       error
	timeout   bool
	temporary bool
	UserLogin string
}

func (e AuthError) Error() string {
	return fmt.Sprintf("AuthError. details: %v", e.err)
}

func (e AuthError) Unwrap() error {
	return e.err
}

func (e AuthError) Timeout() bool {
	return e.timeout
}

func (e AuthError) Temporary() bool {
	return e.temporary
}

