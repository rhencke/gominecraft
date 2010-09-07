package main

import "os"

type Errorer interface {
	Error() <-chan *os.Error
}
