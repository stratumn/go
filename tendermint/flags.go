package tendermint

import (
	"flag"

	cfg "github.com/tendermint/go-config"
	tmconfig "github.com/tendermint/tendermint/config/tendermint"
)

var (
	config    = tmconfig.GetConfig("")
	nodeLaddr = flag.String("node_laddr", config.GetString("node_laddr"), "Node listen address. (0.0.0.0:0 means any interface, any port)")
	moniker   = flag.String("moniker", config.GetString("moniker"), "Node Name")
	seeds     = flag.String("seeds", config.GetString("seeds"), "Comma delimited host:port seed nodes")
	fastSync  = flag.Bool("fast_sync", config.GetBool("fast_sync"), "Fast blockchain syncing")
	skipUPNP  = flag.Bool("skip_upnp", config.GetBool("skip_upnp"), "Skip UPNP configuration")
	rpcLaddr  = flag.String("rpc_laddr", config.GetString("rpc_laddr"), "RPC listen address. Port required")
	grpcLaddr = flag.String("grpc_laddr", config.GetString("grpc_laddr"), "GRPC listen address (BroadcastTx only). Port required")
	logLevel  = flag.String("log_level", config.GetString("log_level"), "Log level")

	// feature flags
	pex = flag.Bool("pex", config.GetBool("pex_reactor"), "Enable Peer-Exchange (dev feature)")
)

// GetConfig returns a Tendermint config setup from flags
func GetConfig() cfg.Config {
	config.Set("moniker", *moniker)
	config.Set("node_laddr", *nodeLaddr)
	config.Set("seeds", *seeds)
	config.Set("fast_sync", *fastSync)
	config.Set("skip_upnp", *skipUPNP)
	config.Set("rpc_laddr", *rpcLaddr)
	config.Set("grpc_laddr", *grpcLaddr)
	config.Set("log_level", *logLevel)

	return config
}
