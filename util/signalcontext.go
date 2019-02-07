// Copyright 2016-2018 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/stratumn/go-core/monitoring"
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
			monitoring.LogEntry().WithField("signal", sig).Info("Got exit signal")
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx
}
