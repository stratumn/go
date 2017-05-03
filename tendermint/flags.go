package tendermint

import (
	"flag"

	cfg "github.com/tendermint/go-config"
	tmconfig "github.com/tendermint/tendermint/config/tendermint"
)

var (
	config = tmconfig.GetConfig("")

	genesisFile       = flag.String("genesis_file", config.GetString("genesis_file"), "The location of the genesis file")
	moniker           = flag.String("moniker", config.GetString("moniker"), "Node name")
	nodeLaddr         = flag.String("node_laddr", config.GetString("node_laddr"), "Node listen address (0.0.0.0:0 means any interface, any port)")
	fastSync          = flag.Bool("fast_sync", config.GetBool("fast_sync"), "Fast blockchain syncing")
	seeds             = flag.String("seeds", config.GetString("seeds"), "Comma delimited host:port seed nodes")
	skipUPNP          = flag.Bool("skip_upnp", config.GetBool("skip_upnp"), "Skip UPNP configuration")
	privValidatorFile = flag.String("priv_validator_file", config.GetString("priv_validator_file"), "Validator private key file")

	dbBackend       = flag.String("db_backend", config.GetString("db_backend"), "Database backend for the blockchain and TendermintCore state (leveldb or memdb)")
	dbDir           = flag.String("db_dir", config.GetString("db_dir"), "Database directory")
	logLevel        = flag.String("log_level", config.GetString("log_level"), "Log level")
	rpcLaddr        = flag.String("rpc_laddr", config.GetString("rpc_laddr"), "RPC listen address (port required)")
	profLaddr       = flag.String("prof_laddr", config.GetString("prof_laddr"), "Profile listen address")
	csWalFile       = flag.String("cs_wal_file", config.GetString("cs_wal_file"), "Consensus state store directory")
	csWalLight      = flag.Bool("cs_wal_light", config.GetBool("cs_wal_light"), "Whether to use light-mode for consensus state WAL")
	blockSize       = flag.Int("block_size", config.GetInt("block_size"), "Maximum number of block txs")
	disableDataHash = flag.Bool("disable_data_hash", config.GetBool("disable_data_hash"), "Disable merklizing block txs")
	pexReactor      = flag.Bool("pex_reactor", config.GetBool("pex_reactor"), "Enable Peer-Exchange (dev feature)")
)

// GetConfig returns a Tendermint config setup from flags
func GetConfig() cfg.Config {
	config.Set("genesis_file", *genesisFile)
	config.Set("moniker", *moniker)
	config.Set("node_laddr", *nodeLaddr)
	config.Set("fast_sync", *fastSync)
	config.Set("seeds", *seeds)
	config.Set("skip_upnp", *skipUPNP)
	config.Set("priv_validator_file", *privValidatorFile)
	config.Set("db_backend", *dbBackend)
	config.Set("db_dir", *dbDir)
	config.Set("log_level", *logLevel)
	config.Set("rpc_laddr", *rpcLaddr)
	config.Set("prof_laddr", *profLaddr)
	config.Set("genesis_file", *genesisFile)
	config.Set("cs_wal_file", *csWalFile)
	config.Set("cs_wal_light", *csWalLight)
	config.Set("block_size", *blockSize)
	config.Set("disable_data_hash", *disableDataHash)
	config.Set("pex_reactor", *pexReactor)

	return config
}
