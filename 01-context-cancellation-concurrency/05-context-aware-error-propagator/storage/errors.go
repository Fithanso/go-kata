package storage

import (
	"fmt"
)

type BlobError struct {
	err       error
	timeout   bool
	temporary bool
}

func (e BlobError) Error() string {
	return fmt.Sprintf("BlobError. details: %v", e.err)
}

func (e BlobError) Unwrap() error {
	return e.err
}

func (e BlobError) Timeout() bool {
	return e.timeout
}

func (e BlobError) Temporary() bool {
	return e.temporary
}