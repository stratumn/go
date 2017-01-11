package tmstore

import (
	"github.com/tendermint/go-rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

// TMClient is the type that implements the Tendermint RPC Client interface
type TMClient struct {
	remote   string
	endpoint string
	rpc      *rpcclient.ClientJSONRPC
}

// NewTMClient creates a new HTTPClient that communicates with Tendermint
func NewTMClient(remote string) *TMClient {
	return &TMClient{
		rpc:    rpcclient.NewClientJSONRPC(remote),
		remote: remote,
	}
}

// Status returns the status of the HTTP Client
func (c *TMClient) Status() (*ctypes.ResultStatus, error) {
	tmResult := new(ctypes.TMResult)
	_, err := c.rpc.Call("status", []interface{}{}, tmResult)
	if err != nil {
		return nil, err
	}
	// note: panics if rpc doesn't match.  okay???
	return (*tmResult).(*ctypes.ResultStatus), nil
}

// ABCIInfo returns info about the Tendermint App
func (c *TMClient) ABCIInfo() (*ctypes.ResultABCIInfo, error) {
	tmResult := new(ctypes.TMResult)
	_, err := c.rpc.Call("abci_info", []interface{}{}, tmResult)
	if err != nil {
		return nil, err
	}
	return (*tmResult).(*ctypes.ResultABCIInfo), nil
}

// ABCIQuery sends a query to the Tendermint App
func (c *TMClient) ABCIQuery(query []byte) (*ctypes.ResultABCIQuery, error) {
	tmResult := new(ctypes.TMResult)
	_, err := c.rpc.Call("abci_query", []interface{}{query}, tmResult)
	if err != nil {
		return nil, err
	}
	return (*tmResult).(*ctypes.ResultABCIQuery), nil
}

// BroadcastTxCommit broadcasts a Tx to the Tendermint Core
func (c *TMClient) BroadcastTxCommit(tx types.Tx) (*ctypes.ResultBroadcastTxCommit, error) {
	tmResult := new(ctypes.TMResult)
	_, err := c.rpc.Call("broadcast_tx_commit", []interface{}{tx}, tmResult)
	if err != nil {
		return nil, err
	}
	return (*tmResult).(*ctypes.ResultBroadcastTxCommit), nil
}

// BroadcastTxAsync broadcasts a Tx to the Tendermint Core
func (c *TMClient) BroadcastTxAsync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTX("broadcast_tx_async", tx)
}

// BroadcastTxSync broadcasts a Tx synchronously to the Tendermint Core
func (c *TMClient) BroadcastTxSync(tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	return c.broadcastTX("broadcast_tx_sync", tx)
}

func (c *TMClient) broadcastTX(route string, tx types.Tx) (*ctypes.ResultBroadcastTx, error) {
	tmResult := new(ctypes.TMResult)
	_, err := c.rpc.Call(route, []interface{}{tx}, tmResult)
	if err != nil {
		return nil, err
	}
	return (*tmResult).(*ctypes.ResultBroadcastTx), nil
}
