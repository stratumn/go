package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// CancelOnInterrupt creates a context and calls the context cancel function when an interrupt signal is caught
func CancelOnInterrupt(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		defer func() {
			signal.Stop(c)
			cancel()
		}()
		select {
		case sig := <-c:
			log.WithField("signal", sig).Info("Got exit signal")
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx
}
