package tmstore

import (
	"encoding/json"

	"github.com/tendermint/go-rpc/client"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
)

// TMClient is the type that implements the Tendermint RPC Client interface
type TMClient struct {
	remote     string
	wsEndpoint string
	rpc        *rpcclient.ClientJSONRPC
	ws         *rpcclient.WSClient
}

// NewTMClient creates a new HTTPClient that communicates with Tendermint
func NewTMClient(remote string) *TMClient {
	return &TMClient{
		rpc:        rpcclient.NewClientJSONRPC(remote),
		remote:     remote,
		wsEndpoint: "/websocket",
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

/** websocket event stuff here... **/

// StartWebsocket starts up a websocket and a listener goroutine
// if already started, do nothing
func (c *TMClient) StartWebsocket() error {
	var err error
	if c.ws == nil {
		ws := rpcclient.NewWSClient(c.remote, c.wsEndpoint)
		_, err = ws.Start()
		if err == nil {
			c.ws = ws
		}
	}
	return err
}

// StopWebsocket stops the websocket connection
func (c *TMClient) StopWebsocket() {
	if c.ws != nil {
		c.ws.Stop()
		c.ws = nil
	}
}

// GetEventChannels returns the results and error channel from the websocket
func (c *TMClient) GetEventChannels() (chan json.RawMessage, chan error, chan struct{}) {
	if c.ws == nil {
		return nil, nil, nil
	}
	return c.ws.ResultsCh, c.ws.ErrorsCh, c.ws.Quit
}

// Subscribe to websocket channel
func (c *TMClient) Subscribe(event string) error {
	return c.ws.Subscribe(event)
}

// Unsubscribe from websocket channel
func (c *TMClient) Unsubscribe(event string) error {
	return c.ws.Unsubscribe(event)
}
