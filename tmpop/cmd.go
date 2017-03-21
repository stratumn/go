// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tmpop

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	log "github.com/Sirupsen/logrus"

	"github.com/stratumn/sdk/store"
	"github.com/tendermint/abci/server"
)

// Run launches a TMPop Tendermint App
func Run(a store.Adapter, config *Config, addrPtr, abciPtr *string) {
	adapterInfo, err := a.GetInfo()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("TMPop v%s@%s", config.Version, config.Commit[:7])
	log.Infof("Adapter %v", adapterInfo)
	log.Info("Copyright (c) 2017 Stratumn SAS")
	log.Info("Mozilla Public License 2.0")
	log.Infof("Runtime %s %s %s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	tmpop, err := New(a, config)
	if err != nil {
		log.Fatal(err)
	}

	srv, err := server.NewServer(*addrPtr, *abciPtr, tmpop)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		sigc := make(chan os.Signal)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigc
		log.WithField("signal", sig).Info("Got exit signal")
		srv.Stop()
		log.Info("Stopped")
		os.Exit(0)
	}()
	select {}
}
