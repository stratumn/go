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

package tendermint

import (
	"flag"
	"os"

	cfg "github.com/tendermint/tendermint/config"
)

var (
	config = cfg.DefaultConfig()

	rootDir           string
	chainID           string
	privValidatorFile string
	moniker           string
	genesisFile       string
	proxyApp          string
	abciVal           string
	logLevel          string
	profLaddr         string
	fastSync          bool
	filterPeers       bool
	txIndex           string
	dbBackend         string
	dbDir             string

	nodeLaddr               string
	seeds                   string
	skipUPNP                bool
	addrBook                string
	addrBookStrict          bool
	pexReactor              bool
	maxNumPeers             int
	flushThrottleTimeout    int
	maxMsgPacketPayloadSize int
	sendRate                int64
	recvRate                int64

	poolRecheck      bool
	poolRecheckEmpty bool
	poolBroadcast    bool
	poolWalDir       string

	rpcLaddr       string
	grpcListenAddr string
	unsafe         bool

	csWalFile  string
	csWalLight bool

	timeoutPropose              int
	timeoutProposeDelta         int
	timeoutPrevote              int
	timeoutPrevoteDelta         int
	timeoutPrecommit            int
	timeoutPrecommitDelta       int
	timeoutCommit               int
	skipTimeoutCommit           bool
	maxBlockSizeTxs             int
	maxBlockSizeBytes           int
	createEmptyBlocks           bool
	createEmptyBlocksInterval   int
	peerGossipSleepDuration     int
	peerQueryMaj23SleepDuration int
)

// RegisterFlags registers the tendermint flags.
func RegisterFlags() {
	flag.StringVar(&rootDir, "home", os.ExpandEnv("$TMHOME"), "Root directory for config and data")

	flag.StringVar(&privValidatorFile, "priv_validator_file", config.BaseConfig.PrivValidator, "Validator private key file")
	flag.StringVar(&moniker, "moniker", config.BaseConfig.Moniker, "Node name")
	flag.StringVar(&genesisFile, "genesis_file", config.BaseConfig.Genesis, "The location of the genesis file")
	flag.StringVar(&proxyApp, "proxy_app", config.BaseConfig.ProxyApp, "TCP or UNIX socket address of the ABCI application, or the name of an ABCI application compiled in with the Tendermint binary")
	flag.StringVar(&abciVal, "abci", config.BaseConfig.ABCI, "Mechanism to connect to the ABCI application: socket | grpc")
	flag.StringVar(&logLevel, "log_level", config.BaseConfig.LogLevel, "Log level")
	flag.StringVar(&profLaddr, "prof_laddr", config.BaseConfig.ProfListenAddress, "Profile listen address")
	flag.BoolVar(&fastSync, "fast_sync", config.BaseConfig.FastSync, "Fast blockchain syncing")
	flag.BoolVar(&filterPeers, "filter_peers", config.BaseConfig.FilterPeers, "If true, query the ABCI app on connecting to a new peer so the app can decide if we should keep the connection or not")
	flag.StringVar(&dbBackend, "db_backend", config.BaseConfig.DBBackend, "Database backend for the blockchain and TendermintCore state (leveldb or memdb)")
	flag.StringVar(&dbDir, "db_dir", config.BaseConfig.DBPath, "Database directory")

	flag.StringVar(&txIndex, "tx_index", config.TxIndex.Indexer, "What indexer to use for transactions")

	flag.StringVar(&nodeLaddr, "node_laddr", config.P2P.ListenAddress, "Node listen address (0.0.0.0:0 means any interface, any port)")
	flag.StringVar(&seeds, "seeds", config.P2P.Seeds, "Comma delimited host:port seed nodes")
	flag.BoolVar(&skipUPNP, "skip_upnp", config.P2P.SkipUPNP, "Skip UPNP configuration")
	flag.StringVar(&addrBook, "addr_book_file", config.P2P.AddrBook, "")
	flag.BoolVar(&addrBookStrict, "addr_book_strict", config.P2P.AddrBookStrict, "")
	flag.BoolVar(&pexReactor, "pex_reactor", config.P2P.PexReactor, "Enable Peer-Exchange (dev feature)")
	flag.IntVar(&maxNumPeers, "max_num_peers", config.P2P.MaxNumPeers, "")
	flag.IntVar(&flushThrottleTimeout, "flush_throttle_timeout", config.P2P.FlushThrottleTimeout, "Time to wait before flushing messages out on the connection, in ms")
	flag.IntVar(&maxMsgPacketPayloadSize, "max_msg_packet_payload_size", config.P2P.MaxMsgPacketPayloadSize, "Maximum size of a message packet payload, in bytes")
	flag.Int64Var(&sendRate, "send_rate", config.P2P.SendRate, "Rate at which packets can be sent, in bytes/second")
	flag.Int64Var(&recvRate, "recv_rate", config.P2P.RecvRate, "Rate at which packets can be received, in bytes/second")

	flag.BoolVar(&poolRecheck, "pool_recheck", config.Mempool.Recheck, "")
	flag.BoolVar(&poolRecheckEmpty, "pool_recheck_empty", config.Mempool.RecheckEmpty, "")
	flag.BoolVar(&poolBroadcast, "pool_broadcast", config.Mempool.Broadcast, "")
	flag.StringVar(&poolWalDir, "pool_wal_dir", config.Mempool.WalPath, "")

	flag.StringVar(&rpcLaddr, "rpc_laddr", config.RPC.ListenAddress, "RPC listen address (port required)")
	flag.StringVar(&grpcListenAddr, "grpc_laddr", config.RPC.GRPCListenAddress, "TCP or UNIX socket address for the gRPC server to listen on. This server only supports /broadcast_tx_commit")
	flag.BoolVar(&unsafe, "unsafe", config.RPC.Unsafe, "Activate unsafe RPC commands like /dial_seeds and /unsafe_flush_mempool")

	flag.StringVar(&csWalFile, "cs_wal_file", config.Consensus.WalPath, "Consensus state store directory")
	flag.BoolVar(&csWalLight, "cs_wal_light", config.Consensus.WalLight, "Whether to use light-mode for consensus state WAL")

	flag.IntVar(&timeoutPropose, "timeout_propose", config.Consensus.TimeoutPropose, "")
	flag.IntVar(&timeoutProposeDelta, "timeout_propose_delta", config.Consensus.TimeoutProposeDelta, "")
	flag.IntVar(&timeoutPrevote, "timeout_prevote", config.Consensus.TimeoutPrevote, "")
	flag.IntVar(&timeoutPrevoteDelta, "timeout_prevote_delta", config.Consensus.TimeoutPrevoteDelta, "")
	flag.IntVar(&timeoutPrecommit, "timeout_precommit", config.Consensus.TimeoutPrecommit, "")
	flag.IntVar(&timeoutPrecommitDelta, "timeout_precommit_delta", config.Consensus.TimeoutPrecommitDelta, "")
	flag.IntVar(&timeoutCommit, "timeout_commit", config.Consensus.TimeoutCommit, "")
	flag.BoolVar(&skipTimeoutCommit, "skip_timeout_commit", config.Consensus.SkipTimeoutCommit, "Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)")
	flag.IntVar(&maxBlockSizeTxs, "max_block_size_txs", config.Consensus.MaxBlockSizeTxs, "Maximum number of block txs")
	flag.IntVar(&maxBlockSizeBytes, "max_block_size_bytes", config.Consensus.MaxBlockSizeBytes, "Maximum block size")
	flag.BoolVar(&createEmptyBlocks, "create_empty_blocks", config.Consensus.CreateEmptyBlocks, "EmptyBlocks mode")
	flag.IntVar(&createEmptyBlocksInterval, "create_empty_blocks_interval", config.Consensus.CreateEmptyBlocksInterval, "Possible interval between empty blocks in seconds")
	flag.IntVar(&peerGossipSleepDuration, "peer_gossip_sleep_duration", config.Consensus.PeerGossipSleepDuration, "Reactor sleep duration parameters, in ms")
	flag.IntVar(&peerQueryMaj23SleepDuration, "peer_query_maj23_sleep_duration", config.Consensus.PeerQueryMaj23SleepDuration, "Reactor sleep duration parameters, in ms")
}

// GetConfig returns a Tendermint config setup from flags
func GetConfig() *cfg.Config {
	if rootDir != "" {
		config.SetRoot(rootDir)
	} else {
		config.SetRoot(os.ExpandEnv("$HOME/.tendermint"))
	}

	config.BaseConfig.PrivValidator = privValidatorFile
	config.BaseConfig.Moniker = moniker
	config.BaseConfig.Genesis = genesisFile
	config.BaseConfig.ProxyApp = proxyApp
	config.BaseConfig.ABCI = abciVal
	config.BaseConfig.LogLevel = logLevel
	config.BaseConfig.ProfListenAddress = profLaddr
	config.BaseConfig.FastSync = fastSync
	config.BaseConfig.FilterPeers = filterPeers
	config.BaseConfig.DBBackend = dbBackend
	config.BaseConfig.DBPath = dbDir

	config.TxIndex.Indexer = txIndex

	config.P2P.ListenAddress = nodeLaddr
	config.P2P.Seeds = seeds
	config.P2P.SkipUPNP = skipUPNP
	config.P2P.AddrBook = addrBook
	config.P2P.AddrBookStrict = addrBookStrict
	config.P2P.PexReactor = pexReactor
	config.P2P.MaxNumPeers = maxNumPeers
	config.P2P.FlushThrottleTimeout = flushThrottleTimeout
	config.P2P.MaxMsgPacketPayloadSize = maxMsgPacketPayloadSize
	config.P2P.SendRate = sendRate
	config.P2P.RecvRate = recvRate

	config.Mempool.Recheck = poolRecheck
	config.Mempool.RecheckEmpty = poolRecheckEmpty
	config.Mempool.Broadcast = poolBroadcast
	config.Mempool.WalPath = poolWalDir

	config.RPC.ListenAddress = rpcLaddr
	config.RPC.GRPCListenAddress = grpcListenAddr
	config.RPC.Unsafe = unsafe

	config.Consensus.WalPath = csWalFile
	config.Consensus.WalLight = csWalLight
	config.Consensus.TimeoutPropose = timeoutPropose
	config.Consensus.TimeoutProposeDelta = timeoutProposeDelta
	config.Consensus.TimeoutPrevote = timeoutPrevote
	config.Consensus.TimeoutPrevoteDelta = timeoutPrevoteDelta
	config.Consensus.TimeoutPrecommit = timeoutPrecommit
	config.Consensus.TimeoutPrecommitDelta = timeoutPrecommitDelta
	config.Consensus.TimeoutCommit = timeoutCommit
	config.Consensus.SkipTimeoutCommit = skipTimeoutCommit
	config.Consensus.MaxBlockSizeTxs = maxBlockSizeTxs
	config.Consensus.MaxBlockSizeBytes = maxBlockSizeBytes
	config.Consensus.CreateEmptyBlocks = createEmptyBlocks
	config.Consensus.CreateEmptyBlocksInterval = createEmptyBlocksInterval
	config.Consensus.PeerGossipSleepDuration = peerGossipSleepDuration
	config.Consensus.PeerQueryMaj23SleepDuration = peerQueryMaj23SleepDuration

	return config
}
