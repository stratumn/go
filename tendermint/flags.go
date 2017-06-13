package tendermint

import (
	"flag"
	"os"

	cfg "github.com/tendermint/tendermint/config"
)

var (
	config = cfg.DefaultConfig()

	rootDir = flag.String("home", os.ExpandEnv("$HOME/.tendermint"), "Root directory for config and data")

	chainID           = flag.String("chain_id", config.BaseConfig.ChainID, "The ID of the chain to join (should be signed with every transaction and vote)")
	privValidatorFile = flag.String("priv_validator_file", config.BaseConfig.PrivValidator, "Validator private key file")
	moniker           = flag.String("moniker", config.BaseConfig.Moniker, "Node name")
	genesisFile       = flag.String("genesis_file", config.BaseConfig.Genesis, "The location of the genesis file")
	logLevel          = flag.String("log_level", config.BaseConfig.LogLevel, "Log level")
	profLaddr         = flag.String("prof_laddr", config.BaseConfig.ProfListenAddress, "Profile listen address")
	fastSync          = flag.Bool("fast_sync", config.BaseConfig.FastSync, "Fast blockchain syncing")
	filterPeers       = flag.Bool("filter_peers", config.BaseConfig.FilterPeers, "If true, query the ABCI app on connecting to a new peer so the app can decide if we should keep the connection or not")
	txIndex           = flag.String("tx_index", config.BaseConfig.TxIndex, "What indexer to use for transactions")
	dbBackend         = flag.String("db_backend", config.BaseConfig.DBBackend, "Database backend for the blockchain and TendermintCore state (leveldb or memdb)")
	dbDir             = flag.String("db_dir", config.BaseConfig.DBPath, "Database directory")

	nodeLaddr      = flag.String("node_laddr", config.P2P.ListenAddress, "Node listen address (0.0.0.0:0 means any interface, any port)")
	seeds          = flag.String("seeds", config.P2P.Seeds, "Comma delimited host:port seed nodes")
	skipUPNP       = flag.Bool("skip_upnp", config.P2P.SkipUPNP, "Skip UPNP configuration")
	addrBook       = flag.String("addr_book_file", config.P2P.AddrBook, "")
	addrBookStrict = flag.Bool("addr_book_strict", config.P2P.AddrBookStrict, "")
	pexReactor     = flag.Bool("pex_reactor", config.P2P.PexReactor, "Enable Peer-Exchange (dev feature)")
	maxNumPeers    = flag.Int("max_num_peers", config.P2P.MaxNumPeers, "")

	rpcLaddr       = flag.String("rpc_laddr", config.RPC.ListenAddress, "RPC listen address (port required)")
	grpcListenAddr = flag.String("grpc_laddr", config.RPC.GRPCListenAddress, "TCP or UNIX socket address for the gRPC server to listen on. This server only supports /broadcast_tx_commit")
	unsafe         = flag.Bool("unsafe", config.RPC.Unsafe, "Activate unsafe RPC commands like /dial_seeds and /unsafe_flush_mempool")

	csWalFile  = flag.String("cs_wal_file", config.Consensus.WalPath, "Consensus state store directory")
	csWalLight = flag.Bool("cs_wal_light", config.Consensus.WalLight, "Whether to use light-mode for consensus state WAL")

	timeoutPropose        = flag.Int("timeout_propose", config.Consensus.TimeoutPropose, "")
	timeoutProposeDelta   = flag.Int("timeout_propose_delta", config.Consensus.TimeoutProposeDelta, "")
	timeoutPrevote        = flag.Int("timeout_prevote", config.Consensus.TimeoutPrevote, "")
	timeoutPrevoteDelta   = flag.Int("timeout_prevote_delta", config.Consensus.TimeoutPrevoteDelta, "")
	timeoutPrecommit      = flag.Int("timeout_precommit", config.Consensus.TimeoutPrecommit, "")
	timeoutPrecommitDelta = flag.Int("timeout_precommit_delta", config.Consensus.TimeoutPrecommitDelta, "")
	timeoutCommit         = flag.Int("timeout_commit", config.Consensus.TimeoutCommit, "")
	skipTimeoutCommit     = flag.Bool("skip_timeout_commit", config.Consensus.SkipTimeoutCommit, "Make progress as soon as we have all the precommits (as if TimeoutCommit = 0)")
	maxBlockSizeTxs       = flag.Int("max_block_size_txs", config.Consensus.MaxBlockSizeTxs, "Maximum number of block txs")
	maxBlockSizeBytes     = flag.Int("max_block_size_bytes", config.Consensus.MaxBlockSizeBytes, "Maximum block size")
)

// GetConfig returns a Tendermint config setup from flags
func GetConfig() *cfg.Config {
	config.SetRoot(*rootDir)

	config.BaseConfig.ChainID = *chainID
	config.BaseConfig.PrivValidator = *privValidatorFile
	config.BaseConfig.Moniker = *moniker
	config.BaseConfig.Genesis = *genesisFile
	config.BaseConfig.LogLevel = *logLevel
	config.BaseConfig.ProfListenAddress = *profLaddr
	config.BaseConfig.FastSync = *fastSync
	config.BaseConfig.FilterPeers = *filterPeers
	config.BaseConfig.TxIndex = *txIndex
	config.BaseConfig.DBBackend = *dbBackend
	config.BaseConfig.DBPath = *dbDir

	config.P2P.ListenAddress = *nodeLaddr
	config.P2P.Seeds = *seeds
	config.P2P.SkipUPNP = *skipUPNP
	config.P2P.AddrBook = *addrBook
	config.P2P.AddrBookStrict = *addrBookStrict
	config.P2P.PexReactor = *pexReactor
	config.P2P.MaxNumPeers = *maxNumPeers

	config.RPC.ListenAddress = *rpcLaddr
	config.RPC.GRPCListenAddress = *grpcListenAddr
	config.RPC.Unsafe = *unsafe

	config.Consensus.WalPath = *csWalFile
	config.Consensus.WalLight = *csWalLight
	config.Consensus.TimeoutPropose = *timeoutPropose
	config.Consensus.TimeoutProposeDelta = *timeoutProposeDelta
	config.Consensus.TimeoutPrevote = *timeoutPrevote
	config.Consensus.TimeoutPrevoteDelta = *timeoutPrevoteDelta
	config.Consensus.TimeoutPrecommit = *timeoutPrecommit
	config.Consensus.TimeoutPrecommitDelta = *timeoutPrecommitDelta
	config.Consensus.TimeoutCommit = *timeoutCommit
	config.Consensus.SkipTimeoutCommit = *skipTimeoutCommit
	config.Consensus.MaxBlockSizeTxs = *maxBlockSizeTxs
	config.Consensus.MaxBlockSizeBytes = *maxBlockSizeBytes

	return config
}
