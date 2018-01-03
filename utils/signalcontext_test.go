package utils

import (
	"context"
	"syscall"
	"testing"
)

func TestCancelOnInterrupt(t *testing.T) {
	ctx := CancelOnInterrupt(context.Background())
	close := make(chan struct{})

	go func() {
		select {
		case <-ctx.Done():
			close <- struct{}{}
		}
	}()

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	// if ctx.Done() is never notified, this will be blocking and fail the test
	<-close
}
