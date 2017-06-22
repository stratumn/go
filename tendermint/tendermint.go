// // Copyright 2017 Stratumn SAS. All rights reserved.
// //
// // This Source Code Form is subject to the terms of the Mozilla Public
// // License, v. 2.0. If a copy of the MPL was not distributed with this
// // file, You can obtain one at http://mozilla.org/MPL/2.0/.

package tendermint

import (
	log "github.com/Sirupsen/logrus"
	abci "github.com/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/cli/flags"
	tmlog "github.com/tendermint/tmlibs/log"
)

// RunNodeForever runs a tendermint node with an in-proc ABCI app and waits for an exit signal
func RunNodeForever(config *cfg.Config, app abci.Application) {
	node := NewNode(config, app)
	node.Start()
	node.RunForever()
}

// NewNode creates a tendermint node with an in-proc ABCI app
func NewNode(config *cfg.Config, app abci.Application) *node.Node {
	logger := tmlog.NewTMLogger(log.StandardLogger().Out)
	logger, _ = flags.ParseLogLevel(config.BaseConfig.LogLevel, logger, "info")

	privValidator := types.LoadOrGenPrivValidator(config.PrivValidatorFile(), logger)
	return node.NewNode(config, privValidator, proxy.NewLocalClientCreator(app), logger)
}
