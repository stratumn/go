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
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/cs/evidences"
	"github.com/stratumn/sdk/merkle"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
	"github.com/stratumn/sdk/validator"
	abci "github.com/tendermint/abci/types"
)

// tmpopLastBlockKey is the database key where last block information are saved.
var tmpopLastBlockKey = []byte("tmpop:lastblock")

// LastBlock saves the information of the last block committed for Core/App Handshake on crash/restart.
type LastBlock struct {
	AppHash    *types.Bytes32
	Height     uint64
	LastHeader *abci.Header
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

	// JSON schema rules definition
	ValidatorFilename string
}

// TMPop is the type of the application that implements github.com/tendermint/abci/types.Application,
// the tendermint socket protocol (ABCI).
type TMPop struct {
	abci.BaseApplication

	state         *State
	adapter       store.Adapter
	kvDB          store.KeyValueStore
	lastBlock     *LastBlock
	config        *Config
	currentHeader *abci.Header
	tmClient      TendermintClient
	eventsManager eventsManager
}

const (
	// Name of the Tendermint Application.
	Name = "TMPop"

	// Description of this Tendermint Application.
	Description = "Agent Store in a Blockchain"
)

const (
	// CodeTypeValidation is the ABCI error code for a validation error.
	CodeTypeValidation abci.CodeType = 400
)

// New creates a new instance of a TMPop.
func New(a store.Adapter, kv store.KeyValueStore, config *Config) (*TMPop, error) {
	initalized, err := kv.GetValue(tmpopLastBlockKey)
	if err != nil {
		return nil, err
	}
	if initalized == nil {
		log.Debug("No existing db, creating new db")
		saveLastBlock(kv, LastBlock{
			AppHash: types.NewBytes32FromBytes(nil),
			Height:  0,
		})
	} else {
		log.Debug("Loading existing db")
	}

	lastBlock, err := ReadLastBlock(kv)
	if err != nil {
		return nil, err
	}

	s, err := NewState(a)
	if err != nil {
		return nil, err
	}

	return &TMPop{
		state:         s,
		adapter:       a,
		kvDB:          kv,
		lastBlock:     lastBlock,
		config:        config,
		currentHeader: lastBlock.LastHeader,
	}, nil
}

// ConnectTendermint connects TMPoP to a Tendermint node
func (t *TMPop) ConnectTendermint(tmClient TendermintClient) {
	t.tmClient = tmClient
	log.Info("TMPoP connected to Tendermint Core")
}

// Info implements github.com/tendermint/abci/types.Application.Info.
func (t *TMPop) Info(req abci.RequestInfo) abci.ResponseInfo {
	// In case we don't have an app hash, Tendermint requires us to return
	// an empty byte slice (instead of a 32-byte array of 0)
	// Otherwise handshake will not work
	lastAppHash := []byte{}
	if !t.lastBlock.AppHash.Zero() {
		lastAppHash = t.lastBlock.AppHash[:]
	}

	return abci.ResponseInfo{
		Data:             Name,
		Version:          t.config.Version,
		LastBlockHeight:  t.lastBlock.Height,
		LastBlockAppHash: lastAppHash,
	}
}

// SetOption implements github.com/tendermint/abci/types.Application.SetOption.
func (t *TMPop) SetOption(key string, value string) (log string) {
	return "No options are supported yet"
}

// BeginBlock implements github.com/tendermint/abci/types.Application.BeginBlock.
func (t *TMPop) BeginBlock(req abci.RequestBeginBlock) {
	t.currentHeader = req.GetHeader()
	if t.currentHeader == nil {
		log.Error("Cannot begin block without header")
		return
	}

	// If the AppHash of the previous block is present in this block's header,
	// consensus has been formed around it.
	// This AppHash will never be denied in a future block so we can add
	// evidence to the links that were added in the previous blocks.
	if t.lastBlock.AppHash.EqualsBytes(t.currentHeader.AppHash) {
		t.addTendermintEvidence(req.Header)
	} else {
		log.Warnf("Unexpected AppHash in BeginBlock, got %x, expected %x",
			t.currentHeader.AppHash,
			*t.lastBlock.AppHash)
	}

	// TODO: we don't need to re-load the file for each block, it's expensive.
	// We should improve this and only reload when a config update was committed.
	if t.config.ValidatorFilename != "" {
		rootValidator := validator.NewRootValidator(t.config.ValidatorFilename, true)
		t.state.validator = &rootValidator
	}

	t.state.previousAppHash = types.NewBytes32FromBytes(t.currentHeader.AppHash)
}

// DeliverTx implements github.com/tendermint/abci/types.Application.DeliverTx.
func (t *TMPop) DeliverTx(tx []byte) abci.Result {
	return t.doTx(t.state.Deliver, tx)
}

// CheckTx implements github.com/tendermint/abci/types.Application.CheckTx.
func (t *TMPop) CheckTx(tx []byte) abci.Result {
	return t.doTx(t.state.Check, tx)
}

// Commit implements github.com/tendermint/abci/types.Application.Commit.
// It actually commits the current state in the Store.
func (t *TMPop) Commit() abci.Result {
	appHash, links, err := t.state.Commit()
	if err != nil {
		return abci.NewError(abci.CodeType_InternalError, err.Error())
	}

	if err := t.saveValidatorHash(); err != nil {
		return abci.NewError(abci.CodeType_InternalError, err.Error())
	}

	if err := t.saveCommitLinkHashes(links); err != nil {
		return abci.NewError(abci.CodeType_InternalError, err.Error())
	}

	t.eventsManager.AddSavedLinks(links)

	t.lastBlock.AppHash = appHash
	t.lastBlock.Height = t.currentHeader.Height
	t.lastBlock.LastHeader = t.currentHeader
	saveLastBlock(t.kvDB, *t.lastBlock)

	return abci.NewResultOK(appHash[:], "")
}

// Query implements github.com/tendermint/abci/types.Application.Query.
func (t *TMPop) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	if reqQuery.Height != 0 {
		resQuery.Code = abci.CodeType_InternalError
		resQuery.Log = "tmpop only supports queries on latest commit"
		return
	}

	resQuery.Height = t.lastBlock.Height

	var err error
	var result interface{}

	switch reqQuery.Path {
	case GetInfo:
		result = &Info{
			Name:        Name,
			Description: Description,
			Version:     t.config.Version,
			Commit:      t.config.Commit,
		}
	case GetSegment:
		linkHash := &types.Bytes32{}
		if err = linkHash.UnmarshalJSON(reqQuery.Data); err != nil {
			break
		}

		result, err = t.adapter.GetSegment(linkHash)

	case GetEvidences:
		linkHash := &types.Bytes32{}
		if err = linkHash.UnmarshalJSON(reqQuery.Data); err != nil {
			break
		}

		result, err = t.adapter.GetEvidences(linkHash)

	case AddEvidence:
		evidence := &struct {
			LinkHash *types.Bytes32
			Evidence *cs.Evidence
		}{}
		if err = json.Unmarshal(reqQuery.Data, evidence); err != nil {
			break
		}

		if err = t.adapter.AddEvidence(evidence.LinkHash, evidence.Evidence); err != nil {
			break
		}

		result = evidence.LinkHash

	case FindSegments:
		filter := &store.SegmentFilter{}
		if err = json.Unmarshal(reqQuery.Data, filter); err != nil {
			break
		}

		result, err = t.adapter.FindSegments(filter)

	case GetMapIDs:
		filter := &store.MapFilter{}
		if err = json.Unmarshal(reqQuery.Data, filter); err != nil {
			break
		}

		result, err = t.adapter.GetMapIDs(filter)

	case PendingEvents:
		result = t.eventsManager.GetPendingEvents()

	default:
		resQuery.Code = abci.CodeType_UnknownRequest
		resQuery.Log = fmt.Sprintf("Unexpected Query path: %v", reqQuery.Path)
	}

	if err != nil {
		resQuery.Code = abci.CodeType_InternalError
		resQuery.Log = err.Error()

		return
	}
	if result != nil {
		resBytes, err := json.Marshal(result)
		if err != nil {
			resQuery.Code = abci.CodeType_InternalError
			resQuery.Log = err.Error()
		}

		resQuery.Value = resBytes
	}

	return
}

func (t *TMPop) doTx(createLink func(*cs.Link) abci.Result, txBytes []byte) abci.Result {
	if len(txBytes) == 0 {
		return abci.ErrEncodingError.SetLog("Tx length cannot be zero")
	}

	tx, res := unmarshallTx(txBytes)
	if res.IsErr() {
		return res
	}

	switch tx.TxType {
	case CreateLink:
		return createLink(tx.Link)
	default:
		return abci.ErrUnknownRequest.SetLog(fmt.Sprintf("Unexpected Tx type byte %X", tx.TxType))
	}
}

// addTendermintEvidence computes and stores new evidence
func (t *TMPop) addTendermintEvidence(header *abci.Header) {
	if t.tmClient == nil {
		log.Warn("TMPoP not connected to Tendermint Core.\nEvidence will not be generated.")
		return
	}

	height := header.Height - 1
	if height <= 0 {
		return
	}

	block := t.tmClient.Block(int(height))
	if block.Header == nil {
		log.Warn("Could not get block header.\nEvidence will not be generated.")
		return
	}

	validatorHash, err := t.getValidatorHash(height)
	if err != nil {
		log.Warn("Could not get validator hash for this block.\nEvidence will not be generated.")
		return
	}

	previousAppHash := types.NewBytes32FromBytes(block.Header.AppHash)
	linkHashes, err := t.getCommitLinkHashes(height)
	if err != nil {
		log.Warn("Could not get link hashes for this block.\nEvidence will not be generated.")
		return
	}

	if len(linkHashes) == 0 {
		return
	}

	merkle, err := merkle.NewStaticTree(linkHashes)
	if err != nil {
		log.Warn("Could not create merkle tree for this block.\nEvidence will not be generated.")
		return
	}

	merkleRoot := merkle.Root()

	appHash, err := ComputeAppHash(previousAppHash, validatorHash, merkleRoot)
	if !appHash.EqualsBytes(header.AppHash) {
		log.Warnf("App hash %x doesn't match the header's: %x.\nEvidence will not be generated.",
			*appHash,
			header.AppHash)
		return
	}

	linksPositions := make(map[types.Bytes32]int)
	for i, lh := range linkHashes {
		linksPositions[lh] = i
	}

	newEvidences := make(map[*types.Bytes32]*cs.Evidence)
	for _, tx := range block.Txs {
		// We only create evidence for valid transactions
		linkHash, _ := tx.Link.Hash()
		position, valid := linksPositions[*linkHash]

		if valid {
			evidence := &cs.Evidence{
				Backend:  Name,
				Provider: header.ChainId,
				Proof: &evidences.TendermintProof{
					BlockHeight:     height,
					Root:            merkleRoot,
					Path:            merkle.Path(position),
					ValidationsHash: validatorHash,
					Header:          *block.Header,
					NextHeader:      *header,
				},
			}

			if err := t.adapter.AddEvidence(linkHash, evidence); err != nil {
				log.Warnf("Evidence could not be added to local store: %v", err)
			}

			newEvidences[linkHash] = evidence
		}
	}

	t.eventsManager.AddSavedEvidences(newEvidences)
}
