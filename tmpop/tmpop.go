// Copyright 2016-2018 Stratumn SAS. All rights reserved.
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
	"context"
	"encoding/json"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/tmpop/evidences"
	"github.com/stratumn/go-core/types"
	"github.com/stratumn/go-core/validation"
	"github.com/stratumn/merkle"
	abci "github.com/tendermint/abci/types"
)

// tmpopLastBlockKey is the database key where last block information are saved.
var tmpopLastBlockKey = []byte("tmpop:lastblock")

// LastBlock saves the information of the last block committed for Core/App Handshake on crash/restart.
type LastBlock struct {
	AppHash    []byte
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

	// path to the rules definition and validator plugins
	Validation *validation.Config

	// Monitoring configuration
	Monitoring *monitoring.Config
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
func New(ctx context.Context, a store.Adapter, kv store.KeyValueStore, config *Config) (*TMPop, error) {
	initialized, err := kv.GetValue(ctx, tmpopLastBlockKey)
	if err != nil {
		return nil, err
	}
	if initialized == nil {
		monitoring.LogEntry().Debug("No existing db, creating new db")
		saveLastBlock(ctx, kv, LastBlock{
			AppHash: nil,
			Height:  0,
		})
	} else {
		monitoring.LogEntry().Debug("Loading existing db")
	}

	lastBlock, err := ReadLastBlock(ctx, kv)
	if err != nil {
		return nil, err
	}

	s, err := NewState(ctx, a, config)
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
	monitoring.LogEntry().Info("TMPoP connected to Tendermint Core")
}

// Info implements github.com/tendermint/abci/types.Application.Info.
func (t *TMPop) Info(req abci.RequestInfo) abci.ResponseInfo {
	span, _ := monitoring.StartSpanIncomingRequest(context.Background(), "tmpop/Info")
	defer span.End()

	return abci.ResponseInfo{
		Data:             Name,
		Version:          t.config.Version,
		LastBlockHeight:  t.lastBlock.Height,
		LastBlockAppHash: t.lastBlock.AppHash,
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
	span, ctx := monitoring.StartSpanIncomingRequest(context.Background(), "tmpop/BeginBlock")
	defer span.End()

	t.currentHeader = &req.Header
	if t.currentHeader == nil {
		monitoring.LogEntry().Error("Cannot begin block without header")
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.InvalidArgument))
		span.Context.SetTag(monitoring.ErrorLabel, "Cannot begin block without header")
		return abci.ResponseBeginBlock{}
	}

	blockCount.Inc()
	txPerBlock.Observe(float64(t.currentHeader.NumTxs))

	// If the AppHash of the previous block is present in this block's header,
	// consensus has been formed around it.
	// This AppHash will never be denied in a future block so we can add
	// evidence to the links that were added in the previous blocks.
	if bytes.Equal(t.lastBlock.AppHash, t.currentHeader.AppHash) {
		t.addTendermintEvidence(ctx, &req.Header)
	} else {
		errorMessage := fmt.Sprintf(
			"Unexpected AppHash in BeginBlock, got %x, expected %x",
			t.currentHeader.AppHash,
			t.lastBlock.AppHash,
		)
		monitoring.LogEntry().Warn(errorMessage)
		span.Context.SetTag(monitoring.ErrorLabel, errorMessage)
	}

	t.state.UpdateValidators(ctx)

	t.state.previousAppHash = t.currentHeader.AppHash

	return abci.ResponseBeginBlock{}
}

// DeliverTx implements github.com/tendermint/abci/types.Application.DeliverTx.
func (t *TMPop) DeliverTx(tx []byte) abci.ResponseDeliverTx {
	span, ctx := monitoring.StartSpanIncomingRequest(context.Background(), "tmpop/DeliverTx")
	defer span.End()

	err := t.doTx(ctx, t.state.Deliver, tx)
	if !err.IsOK() {
		txCount.With(prometheus.Labels{txStatus: "invalid"}).Inc()
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.InvalidArgument))
		span.Context.SetTag(monitoring.ErrorLabel, err.Log)

		return abci.ResponseDeliverTx{
			Code: err.Code,
			Log:  err.Log,
		}
	}

	txCount.With(prometheus.Labels{txStatus: "valid"}).Inc()
	return abci.ResponseDeliverTx{}
}

// CheckTx implements github.com/tendermint/abci/types.Application.CheckTx.
func (t *TMPop) CheckTx(tx []byte) abci.ResponseCheckTx {
	span, ctx := monitoring.StartSpanIncomingRequest(context.Background(), "tmpop/CheckTx")
	defer span.End()

	err := t.doTx(ctx, t.state.Check, tx)
	if !err.IsOK() {
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.InvalidArgument))
		span.Context.SetTag(monitoring.ErrorLabel, err.Log)
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
	span, ctx := monitoring.StartSpanIncomingRequest(context.Background(), "tmpop/Commit")
	defer span.End()

	appHash, links, err := t.state.Commit(ctx)
	if err != nil {
		monitoring.LogEntry().Errorf("Error while committing: %s", err)
		monitoring.SetSpanStatus(span, err)
		return abci.ResponseCommit{}
	}

	if err := t.saveValidatorHash(ctx); err != nil {
		monitoring.LogEntry().Errorf("Error while saving validator hash: %s", err)
		monitoring.SetSpanStatus(span, err)
		return abci.ResponseCommit{}
	}

	if err := t.saveCommitLinkHashes(ctx, links); err != nil {
		monitoring.LogEntry().Errorf("Error while saving committed link hashes: %s", err)
		monitoring.SetSpanStatus(span, err)
		return abci.ResponseCommit{}
	}

	t.eventsManager.AddSavedLinks(links)

	t.lastBlock.AppHash = appHash
	t.lastBlock.Height = t.currentHeader.Height
	t.lastBlock.LastHeader = t.currentHeader
	saveLastBlock(ctx, t.kvDB, *t.lastBlock)

	return abci.ResponseCommit{
		Data: appHash[:],
	}
}

// Query implements github.com/tendermint/abci/types.Application.Query.
func (t *TMPop) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	span, ctx := monitoring.StartSpanIncomingRequest(context.Background(), "tmpop/Query")
	span.Context.SetTag("path", reqQuery.Path)
	defer span.End()

	if reqQuery.Height != 0 {
		resQuery.Code = CodeTypeInternalError
		resQuery.Log = "tmpop only supports queries on latest commit"
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.InvalidArgument))
		span.Context.SetTag(monitoring.ErrorLabel, resQuery.Log)
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
		var linkHash chainscript.LinkHash
		if err = json.Unmarshal(reqQuery.Data, &linkHash); err != nil {
			break
		}

		result, err = t.adapter.GetSegment(ctx, linkHash)

	case GetEvidences:
		var linkHash chainscript.LinkHash
		if err = json.Unmarshal(reqQuery.Data, &linkHash); err != nil {
			break
		}

		result, err = t.adapter.GetEvidences(ctx, linkHash)

	case AddEvidence:
		evidence := &struct {
			LinkHash chainscript.LinkHash
			Evidence *chainscript.Evidence
		}{}
		if err = json.Unmarshal(reqQuery.Data, evidence); err != nil {
			break
		}

		if err = t.adapter.AddEvidence(ctx, evidence.LinkHash, evidence.Evidence); err != nil {
			break
		}

		result = evidence.LinkHash

	case FindSegments:
		filter := &store.SegmentFilter{}
		if err = json.Unmarshal(reqQuery.Data, filter); err != nil {
			break
		}

		result, err = t.adapter.FindSegments(ctx, filter)

	case GetMapIDs:
		filter := &store.MapFilter{}
		if err = json.Unmarshal(reqQuery.Data, filter); err != nil {
			break
		}

		result, err = t.adapter.GetMapIDs(ctx, filter)

	case PendingEvents:
		result = t.eventsManager.GetPendingEvents()

	default:
		resQuery.Code = CodeTypeNotImplemented
		resQuery.Log = fmt.Sprintf("Unexpected Query path: %v", reqQuery.Path)
	}

	if err != nil {
		resQuery.Code = CodeTypeInternalError
		resQuery.Log = err.Error()
		monitoring.SetSpanStatus(span, err)
		return
	}

	if result != nil {
		resBytes, err := json.Marshal(result)
		if err != nil {
			resQuery.Code = CodeTypeInternalError
			resQuery.Log = err.Error()
			monitoring.SetSpanStatus(span, err)
		}

		resQuery.Value = resBytes
	}

	return
}

func (t *TMPop) doTx(ctx context.Context, createLink func(context.Context, *chainscript.Link) *ABCIError, txBytes []byte) *ABCIError {
	if len(txBytes) == 0 {
		return &ABCIError{
			Code: CodeTypeValidation,
			Log:  "Tx length cannot be zero",
		}
	}

	tx, err := unmarshallTx(txBytes)
	if !err.IsOK() {
		return err
	}

	switch tx.TxType {
	case CreateLink:
		return createLink(ctx, tx.Link)
	default:
		return &ABCIError{
			Code: CodeTypeNotImplemented,
			Log:  fmt.Sprintf("Unexpected Tx type byte %X", tx.TxType),
		}
	}
}

// addTendermintEvidence computes and stores new evidence
func (t *TMPop) addTendermintEvidence(ctx context.Context, header *abci.Header) {
	span, ctx := monitoring.StartSpanProcessing(ctx, "tmpop/addTendermintEvidence")
	span.Context.SetTag("height", fmt.Sprintf("%d", header.Height))
	defer span.End()

	if t.tmClient == nil {
		monitoring.LogEntry().Warn("TMPoP not connected to Tendermint Core. Evidence will not be generated.")
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.Unavailable))
		span.Context.SetTag(monitoring.ErrorLabel, "TMPoP not connected to Tendermint Core.")
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
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.FailedPrecondition))
		return
	}

	linkHashes, err := t.getCommitLinkHashes(ctx, evidenceHeight)
	if err != nil {
		monitoring.LogEntry().Warnf("Could not get link hashes for block %d. Evidence will not be generated.", header.Height)
		monitoring.SetSpanStatus(span, err)
		return
	}

	if len(linkHashes) == 0 {
		return
	}

	span.Context.SetTag("links_count", fmt.Sprintf("%d", len(linkHashes)))

	validatorHash, err := t.getValidatorHash(ctx, evidenceHeight)
	if err != nil {
		monitoring.LogEntry().Warnf("Could not get validator hash for block %d. Evidence will not be generated.", header.Height)
		monitoring.SetSpanStatus(span, err)
		return
	}

	evidenceBlock, err := t.tmClient.Block(ctx, evidenceHeight)
	if err != nil {
		monitoring.LogEntry().Warnf("Could not get block %d header: %v", header.Height, err)
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.Unavailable))
		span.Context.SetTag(monitoring.ErrorLabel, "Could not get block")
		return
	}

	evidenceNextBlock, err := t.tmClient.Block(ctx, evidenceHeight+1)
	if err != nil {
		monitoring.LogEntry().Warnf("Could not get next block %d header: %v", header.Height, err)
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.Unavailable))
		span.Context.SetTag(monitoring.ErrorLabel, "Could not get next block")
		return
	}

	evidenceLastBlock, err := t.tmClient.Block(ctx, evidenceHeight+2)
	if err != nil {
		monitoring.LogEntry().Warnf("Could not get last block %d header: %v", header.Height, err)
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.Unavailable))
		span.Context.SetTag(monitoring.ErrorLabel, "Could not get last block")
		return
	}

	if len(evidenceNextBlock.Votes) == 0 || len(evidenceLastBlock.Votes) == 0 {
		monitoring.LogEntry().Warnf("Block %d isn't signed by validator nodes. Evidence will not be generated.", header.Height)
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.FailedPrecondition))
		span.Context.SetTag(monitoring.ErrorLabel, "Votes are missing")
		return
	}

	evidenceBlockAppHash := evidenceBlock.Header.AppHash
	leaves := make([][]byte, len(linkHashes))
	for i, lh := range linkHashes {
		leaves[i] = make([]byte, len(lh))
		copy(leaves[i], lh[:])
	}
	merkle, err := merkle.NewStaticTree(leaves)
	if err != nil {
		monitoring.LogEntry().Warnf("Could not create merkle tree for block %d. Evidence will not be generated.", header.Height)
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.InvalidArgument))
		span.Context.SetTag(monitoring.ErrorLabel, "Could not create merkle tree")
		return
	}

	merkleRoot := merkle.Root()
	appHash := ComputeAppHash(evidenceBlockAppHash, validatorHash, merkleRoot)

	if !bytes.Equal(appHash, evidenceNextBlock.Header.AppHash) {
		monitoring.LogEntry().Warnf("App hash %x of block %d doesn't match the header's: %x. Evidence will not be generated.",
			appHash,
			header.Height,
			header.AppHash)
		span.Context.SetTag(monitoring.ErrorCodeLabel, errorcode.Text(errorcode.FailedPrecondition))
		span.Context.SetTag(monitoring.ErrorLabel, "AppHash mismatch")
		return
	}

	linksPositions := make(map[types.Bytes32]int)
	for i, lh := range linkHashes {
		linksPositions[*types.NewBytes32FromBytes(lh)] = i
	}

	newEvidences := make(map[*chainscript.LinkHash]*chainscript.Evidence)
	for _, tx := range evidenceBlock.Txs {
		// We only create evidence for valid transactions
		linkHash, _ := tx.Link.Hash()
		position, valid := linksPositions[*types.NewBytes32FromBytes(linkHash)]

		if valid {
			proof := &evidences.TendermintProof{
				BlockHeight:            evidenceHeight,
				Root:                   merkleRoot,
				Path:                   merkle.Path(position),
				ValidationsHash:        validatorHash,
				Header:                 evidenceBlock.Header,
				HeaderVotes:            evidenceNextBlock.Votes,
				HeaderValidatorSet:     evidenceNextBlock.Validators,
				NextHeader:             evidenceNextBlock.Header,
				NextHeaderVotes:        evidenceLastBlock.Votes,
				NextHeaderValidatorSet: evidenceLastBlock.Validators,
			}

			evidence, err := proof.Evidence(header.ChainID)
			if err != nil {
				monitoring.LogEntry().Warnf("Evidence could not be created: %v", err)
				span.Context.SetTag(linkHash.String(), err.Error())
				continue
			}

			if err := t.adapter.AddEvidence(ctx, linkHash, evidence); err != nil {
				monitoring.LogEntry().Warnf("Evidence could not be added to local store: %v", err)
				span.Context.SetTag(linkHash.String(), err.Error())
				continue
			}

			if evidence != nil {
				newEvidences[&linkHash] = evidence
			}
		}
	}

	t.eventsManager.AddSavedEvidences(newEvidences)
}
