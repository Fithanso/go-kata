package meta

import "fmt"


type StorageQuotaError struct {
	err       error
	timeout   bool
	temporary bool
}

func (e StorageQuotaError) Error() string {
	return fmt.Sprintf("StorageQuotaError. details: %v", e.err)
}

func (e StorageQuotaError) Unwrap() error {
	return e.err
}

func (e StorageQuotaError) Timeout() bool {
	return e.timeout
}

func (e StorageQuotaError) Temporary() bool {
	return e.temporary
}


type MetaDatabaseError struct {
	err       error
	timeout   bool
	temporary bool
}

func (e MetaDatabaseError) Error() string {
	return fmt.Sprintf("MetaDatabaseError. details: %v", e.err)
}

func (e MetaDatabaseError) Unwrap() error {
	return e.err
}

func (e MetaDatabaseError) Timeout() bool {
	return e.timeout
}

func (e MetaDatabaseError) Temporary() bool {
	return e.temporary
}