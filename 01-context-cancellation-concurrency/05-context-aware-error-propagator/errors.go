package main

import "fmt"

type InfernalError struct {
	Err       error
}

func (e InfernalError) Error() string {
	return fmt.Sprintf("Infernal error occured. Details: %v", e.Err)
} 

func (e InfernalError) Unwrap() error {
	return e.Err
}