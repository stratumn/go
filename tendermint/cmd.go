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
	logLevel          string
	profLaddr         string
	fastSync          bool
	filterPeers       bool
	txIndex           string
	dbBackend         string
	dbDir             string

	nodeLaddr      string
	seeds          string
	skipUPNP       bool
	addrBook       string
	addrBookStrict bool
	pexReactor     bool
	maxNumPeers    int

	rpcLaddr       string
	grpcListenAddr string
	unsafe         bool

	csWalFile  string
	csWalLight bool

	timeoutPropose        int
	timeoutProposeDelta   int
	timeoutPrevote        int
	timeoutPrevoteDelta   int
	timeoutPrecommit      int
	timeoutPrecommitDelta int
	timeoutCommit         int
	skipTimeoutCommit     bool
	maxBlockSizeTxs       int
	maxBlockSizeBytes     int
)

// RegisterFlags register the tendemrint flags.
func RegisterFlags() {
	flag.StringVar(&rootDir, "home", os.ExpandEnv("$HOME/.tendermint"), "Root directory for config and data")

	flag.StringVar(&chainID, "chain_id", config.BaseConfig.ChainID, "The ID of the chain to join (should be signed with every transaction and vote)")
	flag.StringVar(&privValidatorFile, "priv_validator_file", config.BaseConfig.PrivValidator, "Validator private key file")
	flag.StringVar(&moniker, "moniker", config.BaseConfig.Moniker, "Node name")
	flag.StringVar(&genesisFile, "genesis_file", config.BaseConfig.Genesis, "The location of the genesis file")
	flag.StringVar(&logLevel, "log_level", config.BaseConfig.LogLevel, "Log level")
	flag.StringVar(&profLaddr, "prof_laddr", config.BaseConfig.ProfListenAddress, "Profile listen address")
	flag.BoolVar(&fastSync, "fast_sync", config.BaseConfig.FastSync, "Fast blockchain syncing")
	flag.BoolVar(&filterPeers, "filter_peers", config.BaseConfig.FilterPeers, "If true, query the ABCI app on connecting to a new peer so the app can decide if we should keep the connection or not")
	flag.StringVar(&txIndex, "tx_index", config.BaseConfig.TxIndex, "What indexer to use for transactions")
	flag.StringVar(&dbBackend, "db_backend", config.BaseConfig.DBBackend, "Database backend for the blockchain and TendermintCore state (leveldb or memdb)")
	flag.StringVar(&dbDir, "db_dir", config.BaseConfig.DBPath, "Database directory")

	flag.StringVar(&nodeLaddr, "node_laddr", config.P2P.ListenAddress, "Node listen address (0.0.0.0:0 means any interface, any port)")
	flag.StringVar(&seeds, "seeds", config.P2P.Seeds, "Comma delimited host:port seed nodes")
	flag.BoolVar(&skipUPNP, "skip_upnp", config.P2P.SkipUPNP, "Skip UPNP configuration")
	flag.StringVar(&addrBook, "addr_book_file", config.P2P.AddrBook, "")
	flag.BoolVar(&addrBookStrict, "addr_book_strict", config.P2P.AddrBookStrict, "")
	flag.BoolVar(&pexReactor, "pex_reactor", config.P2P.PexReactor, "Enable Peer-Exchange (dev feature)")
	flag.IntVar(&maxNumPeers, "max_num_peers", config.P2P.MaxNumPeers, "")

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
}

// GetConfig returns a Tendermint config setup from flags
func GetConfig() *cfg.Config {
	config.SetRoot(rootDir)

	config.BaseConfig.ChainID = chainID
	config.BaseConfig.PrivValidator = privValidatorFile
	config.BaseConfig.Moniker = moniker
	config.BaseConfig.Genesis = genesisFile
	config.BaseConfig.LogLevel = logLevel
	config.BaseConfig.ProfListenAddress = profLaddr
	config.BaseConfig.FastSync = fastSync
	config.BaseConfig.FilterPeers = filterPeers
	config.BaseConfig.TxIndex = txIndex
	config.BaseConfig.DBBackend = dbBackend
	config.BaseConfig.DBPath = dbDir

	config.P2P.ListenAddress = nodeLaddr
	config.P2P.Seeds = seeds
	config.P2P.SkipUPNP = skipUPNP
	config.P2P.AddrBook = addrBook
	config.P2P.AddrBookStrict = addrBookStrict
	config.P2P.PexReactor = pexReactor
	config.P2P.MaxNumPeers = maxNumPeers

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

	return config
}
