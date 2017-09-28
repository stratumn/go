// // Copyright 2017 Stratumn SAS. All rights reserved.
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

	privValidator := types.LoadOrGenPrivValidatorFS(config.PrivValidatorFile())
	ret, err := node.NewNode(config,
		privValidator,
		proxy.NewLocalClientCreator(app),
		node.DefaultGenesisDocProviderFunc(config),
		node.DefaultDBProvider,
		logger)
	if err != nil {
		log.Errorf("Error on new node creation: %s", err.Error())
	}
	return ret
}
