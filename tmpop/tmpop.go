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

package tmpop

import (
	"bytes"
	"encoding/json"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/stratumn/go/cs"
	"github.com/stratumn/go/store"
	"github.com/stratumn/go/types"
	tmtypes "github.com/tendermint/abci/types"
	godb "github.com/tendermint/go-db"
	merkle "github.com/tendermint/go-merkle"
	wire "github.com/tendermint/go-wire"
)

// LastBlockInfo stores information about the last block
type LastBlockInfo struct {
	Height  uint64
	AppHash []byte
}

// Info is the info returned by GetInfo.
type Info struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Version     string      `json:"version"`
	Commit      string      `json:"commit"`
	AdapterInfo interface{} `json:"adapterInfo"`
}

// Config contains configuration options for the App.
type Config struct {
	// A version string that will be set in the store's information.
	Version string

	// A git commit hash that will be set in the store's information.
	Commit string

	// Where godb will be saved.
	DbDir string
}

// TMPop is the type of the application that implements github.com/tendermint/abci/types.Application,
// the tendermint socket protocol (ABCI)
type TMPop struct {
	objects     merkle.Tree
	db          godb.DB
	newSegments map[*types.Bytes32]*cs.Segment
	adapter     store.Adapter
	lastBlock   *LastBlockInfo
	blockHeader *tmtypes.Header
	config      *Config
}

const (
	// Name of the Tendermint Application
	Name = "TMPop"

	// Description of this Tendermint Application
	Description = "Agent Store in a Blockchain"
)

var (
	lastBlockInfoKeyName = []byte("info")
)

// New creates a new instance of a TMPop
func New(a store.Adapter, config *Config) *TMPop {
	db := godb.NewDB(Name, godb.GoLevelDBBackendStr, config.DbDir)
	objects := merkle.NewIAVLTree(0, db)

	h := &TMPop{
		objects:     objects,
		adapter:     a,
		db:          db,
		newSegments: make(map[*types.Bytes32]*cs.Segment),
		config:      config,
	}
	h.LoadLastBlock()

	return h
}

// DeliverTx implements github.com/tendermint/abci/types.Application.DeliverTx
func (t *TMPop) DeliverTx(tx []byte) tmtypes.Result {
	segment, res := unmarshallTx(tx)
	if res.IsErr() {
		return res
	}

	check := t.checkSegment(segment)
	if check.IsErr() {
		return check
	}

	t.newSegments[segment.GetLinkHash()] = segment

	return tmtypes.OK
}

// CheckTx implements github.com/tendermint/abci/types.Application.CheckTx
func (t *TMPop) CheckTx(tx []byte) tmtypes.Result {
	segment, res := unmarshallTx(tx)
	if res.IsErr() {
		return res
	}

	return t.checkSegment(segment)
}

func (t *TMPop) checkSegment(segment *cs.Segment) tmtypes.Result {
	err := segment.Validate()
	if err != nil {
		return tmtypes.ErrUnauthorized.SetLog(fmt.Sprintf("Invalid segment %v: %v", segment, err))
	}

	return tmtypes.OK
}

// Commit implements github.com/tendermint/abci/types.Application.Commit
func (t *TMPop) Commit() tmtypes.Result {
	for _, segment := range t.newSegments {
		evidence := make(map[string]interface{})
		evidence["state"] = "COMPLETE"
		evidence["transactions"] = map[string]string{"[tmpop]:[]": fmt.Sprintf("%v", t.blockHeader.Height)}

		segment.Meta["evidence"] = evidence

		s, err := json.Marshal(segment)
		if err != nil {
			return tmtypes.NewError(tmtypes.CodeType_InternalError, err.Error())
		}
		t.objects.Set([]byte(segment.GetLinkHashString()), s)
		if err := t.adapter.SaveSegment(segment); err != nil {
			return tmtypes.NewError(tmtypes.CodeType_InternalError, err.Error())
		}
	}

	hash := t.objects.Save()

	t.lastBlock = &LastBlockInfo{
		Height:  t.blockHeader.Height,
		AppHash: hash, // this hash will be in the next block header
	}
	t.saveLastBlock()

	return tmtypes.NewResultOK(hash, "")
}

// Query implements github.com/tendermint/abci/types.Application.Query
// it unmarshalls the read queries and forward them to the adapter
func (t *TMPop) Query(q []byte) tmtypes.Result {
	query := &Query{}
	if err := json.Unmarshal(q, query); err != nil {
		return tmtypes.NewError(tmtypes.CodeType_InternalError, err.Error())
	}
	var result interface{}
	var err error

	switch query.Name {
	case "GetInfo":
		var adapterInfo interface{}
		adapterInfo, err = t.adapter.GetInfo()
		result = &Info{
			Name:        Name,
			Description: Description,
			Version:     t.config.Version,
			Commit:      t.config.Commit,
			AdapterInfo: adapterInfo,
		}
	case "GetSegment":
		linkHash := &types.Bytes32{}
		if err := linkHash.UnmarshalJSON(query.Args); err != nil {
			return tmtypes.NewError(tmtypes.CodeType_InternalError, err.Error())
		}
		result, err = t.adapter.GetSegment(linkHash)
	case "FindSegments":
		filter := &store.Filter{}
		if err := json.Unmarshal(query.Args, filter); err != nil {
			return tmtypes.NewError(tmtypes.CodeType_InternalError, err.Error())
		}
		result, err = t.adapter.FindSegments(filter)
	case "GetMapIDs":
		pagination := &store.Pagination{}
		if err := json.Unmarshal(query.Args, pagination); err != nil {
			return tmtypes.NewError(tmtypes.CodeType_InternalError, err.Error())
		}
		result, err = t.adapter.GetMapIDs(pagination)
	case "DeleteSegment":
		linkHash := &types.Bytes32{}
		if err := linkHash.UnmarshalJSON(query.Args); err != nil {
			return tmtypes.NewError(tmtypes.CodeType_InternalError, err.Error())
		}
		result, err = t.adapter.DeleteSegment(linkHash)
	}

	if err != nil {
		return tmtypes.NewError(tmtypes.CodeType_InternalError, err.Error())
	}

	resBytes, err := json.Marshal(result)

	if err != nil {
		return tmtypes.NewError(tmtypes.CodeType_InternalError, err.Error())
	}

	return tmtypes.NewResultOK(resBytes, "OK")
}

// Info implements github.com/tendermint/abci/types.Application.Info
func (t *TMPop) Info() tmtypes.ResponseInfo {
	return tmtypes.ResponseInfo{
		Data:             Name,
		LastBlockHeight:  t.lastBlock.Height,
		LastBlockAppHash: t.lastBlock.AppHash,
	}
}

// SetOption implements github.com/tendermint/abci/types.Application.SetOption
func (t *TMPop) SetOption(key string, value string) (log string) {
	return ""
}

// InitChain implements github.com/tendermint/abci/types.BlockchainAware.InitChain
func (t *TMPop) InitChain(validators []*tmtypes.Validator) {
	log.WithField("validators", validators).Debug("Init Chain")
}

// BeginBlock implements github.com/tendermint/abci/types.BlockchainAware.BeginBlock
func (t *TMPop) BeginBlock(hash []byte, header *tmtypes.Header) {
	log.WithField("header", header).Debug("Begin Block")

	t.blockHeader = header
	t.newSegments = make(map[*types.Bytes32]*cs.Segment)
}

// EndBlock implements github.com/tendermint/abci/types.BlockchainAware.EndBlock
func (t *TMPop) EndBlock(height uint64) tmtypes.ResponseEndBlock {
	log.WithField("height", height).Debug("End Block")

	return tmtypes.ResponseEndBlock{}
}

// LoadLastBlock gets the last block from the db
func (t *TMPop) LoadLastBlock() (lastBlock LastBlockInfo) {
	buf := t.db.Get(lastBlockInfoKeyName)
	if len(buf) != 0 {
		r, n, err := bytes.NewReader(buf), new(int), new(error)
		wire.ReadBinaryPtr(&lastBlock, r, 0, n, err)
		if *err != nil {
			// DATA HAS BEEN CORRUPTED OR THE SPEC HAS CHANGED
			log.Fatalf("Data has been corrupted or its spec has changed: %v\n", *err)
		}
		// TODO: ensure that buf is completely read.

		log.WithFields(log.Fields{
			"Height":  lastBlock.Height,
			"AppHash": lastBlock.AppHash,
			"buf":     buf,
		}).Debug("Loading block")
	}
	t.lastBlock = &lastBlock
	return lastBlock
}

func (t *TMPop) saveLastBlock() {
	log.WithFields(log.Fields{
		"Height":  t.lastBlock.Height,
		"AppHash": t.lastBlock.AppHash,
	}).Debug("Saving block")

	buf, n, err := new(bytes.Buffer), new(int), new(error)
	wire.WriteBinary(*t.lastBlock, buf, n, err)
	if *err != nil {
		log.Fatal(*err)
	}

	t.db.Set(lastBlockInfoKeyName, buf.Bytes())
}

func unmarshallTx(tx []byte) (*cs.Segment, tmtypes.Result) {
	segment := &cs.Segment{}

	if err := json.Unmarshal(tx, segment); err != nil {
		return nil, tmtypes.NewError(tmtypes.CodeType_InternalError, err.Error())
	}

	return segment, tmtypes.NewResultOK([]byte{}, "ok")
}

// SetAdapter sets the adapter
func (t *TMPop) SetAdapter(a store.Adapter) {
	t.adapter = a
}
