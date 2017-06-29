// Copyright 2017 Stratumn SAS. All rights reserved.
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

package fossilizerhttp

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/stratumn/sdk/fossilizer"
	"github.com/stratumn/sdk/jsonhttp"
)

var (
	addr             string
	certFile         string
	keyFile          string
	numResultWorkers int
	minDataLen       int
	maxDataLen       int
	callbackTimeout  time.Duration
	readTimeout      time.Duration
	writeTimeout     time.Duration
	maxHeaderBytes   int
	shutdownTimeout  time.Duration
)

// Run launches a fossilizerhttp server.
func Run(
	a fossilizer.Adapter,
	config *Config,
	httpConfig *jsonhttp.Config,
	shutdownTimeout time.Duration,
) {
	log.Info("Copyright (c) 2017 Stratumn SAS")
	log.Info("Apache License 2.0")
	log.Infof("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	h := New(a, config, httpConfig)

	go func() {
		sigc := make(chan os.Signal)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigc
		log.WithField("signal", sig).Info("Got exit signal")
		log.Info("Cleaning up")
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := h.Shutdown(ctx); err != nil {
			log.WithField("error", err).Fatal("Failed to shutdown server")
		}
		log.Info("Stopped")
		os.Exit(0)
	}()

	log.WithField("http", httpConfig.Address).Info("Listening")
	if err := h.ListenAndServe(); err != nil {
		log.WithField("error", err).Fatal("Server stopped")
	}
}

// RegisterFlags register the flags used by RunWithFlags.
func RegisterFlags() {
	flag.StringVar(&addr, "http", DefaultAddress, "HTTP address")
	flag.StringVar(&certFile, "tls_cert", "", "TLS certificate file")
	flag.StringVar(&keyFile, "tls_key", "", "TLS private key file")
	flag.IntVar(&numResultWorkers, "workers", DefaultNumResultWorkers, "Number of result workers")
	flag.IntVar(&minDataLen, "mindata", DefaultMinDataLen, "Minimum data length")
	flag.IntVar(&maxDataLen, "maxdata", DefaultMaxDataLen, "Maximum data length")
	flag.DurationVar(&callbackTimeout, "callbacktimeout", DefaultCallbackTimeout, "Callback request timeout")
	flag.DurationVar(&readTimeout, "read_timeout", jsonhttp.DefaultReadTimeout, "Read timeout")
	flag.DurationVar(&writeTimeout, "write_timeout", jsonhttp.DefaultWriteTimeout, "Write timeout")
	flag.IntVar(&maxHeaderBytes, "max_header_bytes", jsonhttp.DefaultMaxHeaderBytes, "Maximum header bytes")
	flag.DurationVar(&shutdownTimeout, "shutdown_timeout", 10*time.Second, "Shutdown timeout")
}

// RunWithFlags should be called after RegisterFlags and flag.Parse to launch
// a fossilizerhttp server configured using flag values.
func RunWithFlags(a fossilizer.Adapter) {
	config := &Config{
		NumResultWorkers: numResultWorkers,
		MinDataLen:       minDataLen,
		MaxDataLen:       maxDataLen,
		CallbackTimeout:  callbackTimeout,
	}
	httpConfig := &jsonhttp.Config{
		Address:        addr,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
		CertFile:       certFile,
		KeyFile:        keyFile,
	}

	Run(
		a,
		config,
		httpConfig,
		shutdownTimeout,
	)
}
