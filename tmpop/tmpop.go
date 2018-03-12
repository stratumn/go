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

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/evidences"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/merkle"
	abci "github.com/tendermint/abci/types"
)

// tmpopLastBlockKey is the database key where last block information are saved.
var tmpopLastBlockKey = []byte("tmpop:lastblock")

// LastBlock saves the information of the last block committed for Core/App Handshake on crash/restart.
type LastBlock struct {
	AppHash    *types.Bytes32
	Height     int64
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

// New creates a new instance of a TMPop.
func New(a store.Adapter, kv store.KeyValueStore, config *Config) (*TMPop, error) {
	initialized, err := kv.GetValue(tmpopLastBlockKey)
	if err != nil {
		return nil, err
	}
	if initialized == nil {
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
		return nil, errors.Wrap(err, "cannot read the last block")
	}

	s, err := NewState(a, config)
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
func (t *TMPop) SetOption(req abci.RequestSetOption) abci.ResponseSetOption {
	return abci.ResponseSetOption{
		Code: CodeTypeNotImplemented,
		Log:  "No options are supported yet",
	}
}

// BeginBlock implements github.com/tendermint/abci/types.Application.BeginBlock.
func (t *TMPop) BeginBlock(req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	t.currentHeader = &req.Header
	if t.currentHeader == nil {
		log.Error("Cannot begin block without header")
		return abci.ResponseBeginBlock{}
	}

	// If the AppHash of the previous block is present in this block's header,
	// consensus has been formed around it.
	// This AppHash will never be denied in a future block so we can add
	// evidence to the links that were added in the previous blocks.
	if t.lastBlock.AppHash.EqualsBytes(t.currentHeader.AppHash) {
		t.addTendermintEvidence(&req.Header)
	} else {
		log.Warnf("Unexpected AppHash in BeginBlock, got %x, expected %x",
			t.currentHeader.AppHash,
			*t.lastBlock.AppHash)
	}

	t.state.UpdateValidators()

	t.state.previousAppHash = types.NewBytes32FromBytes(t.currentHeader.AppHash)

	return abci.ResponseBeginBlock{}
}

// DeliverTx implements github.com/tendermint/abci/types.Application.DeliverTx.
func (t *TMPop) DeliverTx(tx []byte) abci.ResponseDeliverTx {
	err := t.doTx(t.state.Deliver, tx)
	if !err.IsOK() {
		return abci.ResponseDeliverTx{
			Code: err.Code,
			Log:  err.Log,
		}
	}

	return abci.ResponseDeliverTx{}
}

// CheckTx implements github.com/tendermint/abci/types.Application.CheckTx.
func (t *TMPop) CheckTx(tx []byte) abci.ResponseCheckTx {
	err := t.doTx(t.state.Check, tx)
	if !err.IsOK() {
		return abci.ResponseCheckTx{
			Code: err.Code,
			Log:  err.Log,
		}
	}

	return abci.ResponseCheckTx{}
}

// Commit implements github.com/tendermint/abci/types.Application.Commit.
// It actually commits the current state in the Store.
func (t *TMPop) Commit() abci.ResponseCommit {
	appHash, links, err := t.state.Commit()
	if err != nil {
		log.Errorf("Error while committing: %s", err)
		return abci.ResponseCommit{}
	}

	if err := t.saveValidatorHash(); err != nil {
		log.Errorf("Error while saving validator hash: %s", err)
		return abci.ResponseCommit{}
	}

	if err := t.saveCommitLinkHashes(links); err != nil {
		log.Errorf("Error while saving committed link hashes: %s", err)
		return abci.ResponseCommit{}
	}

	t.eventsManager.AddSavedLinks(links)

	t.lastBlock.AppHash = appHash
	t.lastBlock.Height = t.currentHeader.Height
	t.lastBlock.LastHeader = t.currentHeader
	saveLastBlock(t.kvDB, *t.lastBlock)

	return abci.ResponseCommit{
		Data: appHash[:],
	}
}

// Query implements github.com/tendermint/abci/types.Application.Query.
func (t *TMPop) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	if reqQuery.Height != 0 {
		resQuery.Code = CodeTypeInternalError
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
		resQuery.Code = CodeTypeNotImplemented
		resQuery.Log = fmt.Sprintf("Unexpected Query path: %v", reqQuery.Path)
	}

	if err != nil {
		resQuery.Code = CodeTypeInternalError
		resQuery.Log = err.Error()

		return
	}
	if result != nil {
		resBytes, err := json.Marshal(result)
		if err != nil {
			resQuery.Code = CodeTypeInternalError
			resQuery.Log = err.Error()
		}

		resQuery.Value = resBytes
	}

	return
}

func (t *TMPop) doTx(createLink func(*cs.Link) *ABCIError, txBytes []byte) *ABCIError {
	if len(txBytes) == 0 {
		return &ABCIError{
			CodeTypeValidation,
			"Tx length cannot be zero",
		}
	}

	tx, err := unmarshallTx(txBytes)
	if !err.IsOK() {
		return err
	}

	switch tx.TxType {
	case CreateLink:
		return createLink(tx.Link)
	default:
		return &ABCIError{
			CodeTypeNotImplemented,
			fmt.Sprintf("Unexpected Tx type byte %X", tx.TxType),
		}
	}
}

// addTendermintEvidence computes and stores new evidence
func (t *TMPop) addTendermintEvidence(header *abci.Header) {
	if t.tmClient == nil {
		log.Warn("TMPoP not connected to Tendermint Core. Evidence will not be generated.")
		return
	}

	// Evidence for block N can only be generated at the beginning of block N+3.
	// That is because we need signatures for both block N and block N+1
	// (since the data is always reflected in the next block's AppHash)
	// so we need block N+1 to be committed.
	// The signatures for block N+1 will only be included in block N+2 so
	// we need block N+2 to be committed.
	evidenceHeight := header.Height - 3
	if evidenceHeight <= 0 {
		return
	}

	linkHashes, err := t.getCommitLinkHashes(evidenceHeight)
	if err != nil {
		log.Warnf("Could not get link hashes for block %d. Evidence will not be generated.", header.Height)
		return
	}

	if len(linkHashes) == 0 {
		return
	}

	validatorHash, err := t.getValidatorHash(evidenceHeight)
	if err != nil {
		log.Warnf("Could not get validator hash for block %d. Evidence will not be generated.", header.Height)
		return
	}

	evidenceBlock, err := t.tmClient.Block(evidenceHeight)
	if err != nil {
		log.Warnf("Could not get block %d header: %v", header.Height, err)
		return
	}

	evidenceNextBlock, err := t.tmClient.Block(evidenceHeight + 1)
	if err != nil {
		log.Warnf("Could not get next block %d header: %v", header.Height, err)
		return
	}

	evidenceLastBlock, err := t.tmClient.Block(evidenceHeight + 2)
	if err != nil {
		log.Warnf("Could not get last block %d header: %v", header.Height, err)
		return
	}

	if len(evidenceNextBlock.Votes) == 0 || len(evidenceLastBlock.Votes) == 0 {
		log.Warnf("Block %d isn't signed by validator nodes. Evidence will not be generated.", header.Height)
		return
	}

	evidenceBlockAppHash := types.NewBytes32FromBytes(evidenceBlock.Header.AppHash)
	leaves := make([][]byte, len(linkHashes), len(linkHashes))
	for i, lh := range linkHashes {
		leaves[i] = make([]byte, len(lh), len(lh))
		copy(leaves[i], lh[:])
	}
	merkle, err := merkle.NewStaticTree(leaves)
	if err != nil {
		log.Warnf("Could not create merkle tree for block %d. Evidence will not be generated.", header.Height)
		return
	}

	merkleRoot := types.NewBytes32FromBytes(merkle.Root())

	appHash, err := ComputeAppHash(evidenceBlockAppHash, validatorHash, merkleRoot)
	if err != nil {
		log.Warnf("Could not compute app hash for block %d. Evidence will not be generated.", header.Height)
		return
	}
	if !appHash.EqualsBytes(evidenceNextBlock.Header.AppHash) {
		log.Warnf("App hash %x of block %d doesn't match the header's: %x. Evidence will not be generated.",
			*appHash,
			header.Height,
			header.AppHash)
		return
	}

	linksPositions := make(map[types.Bytes32]int)
	for i, lh := range linkHashes {
		linksPositions[lh] = i
	}

	newEvidences := make(map[*types.Bytes32]*cs.Evidence)
	for _, tx := range evidenceBlock.Txs {
		// We only create evidence for valid transactions
		linkHash, _ := tx.Link.Hash()
		position, valid := linksPositions[*linkHash]

		if valid {
			evidence := &cs.Evidence{
				Backend:  Name,
				Provider: header.ChainID,
				Proof: &evidences.TendermintProof{
					BlockHeight:     evidenceHeight,
					Root:            merkleRoot,
					Path:            merkle.Path(position),
					ValidationsHash: validatorHash,
					Header:          evidenceBlock.Header,
					HeaderVotes:     evidenceNextBlock.Votes,
					NextHeader:      evidenceNextBlock.Header,
					NextHeaderVotes: evidenceLastBlock.Votes,
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
