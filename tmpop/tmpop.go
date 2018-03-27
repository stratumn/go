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
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/stratumn/go-indigocore/cs"
	"github.com/stratumn/go-indigocore/cs/evidences"
	"github.com/stratumn/go-indigocore/monitoring"
	"github.com/stratumn/go-indigocore/store"
	"github.com/stratumn/go-indigocore/types"
	"github.com/stratumn/merkle"
	abci "github.com/tendermint/abci/types"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
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
		log.Debug("No existing db, creating new db")
		saveLastBlock(ctx, kv, LastBlock{
			AppHash: types.NewBytes32FromBytes(nil),
			Height:  0,
		})
	} else {
		log.Debug("Loading existing db")
	}

	lastBlock, err := ReadLastBlock(ctx, kv)
	if err != nil {
		return nil, errors.Wrap(err, "cannot read the last block")
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
	log.Info("TMPoP connected to Tendermint Core")
}

// Info implements github.com/tendermint/abci/types.Application.Info.
func (t *TMPop) Info(req abci.RequestInfo) abci.ResponseInfo {
	_, span := trace.StartSpan(context.Background(), "tmpop/Info")
	defer span.End()

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
	ctx, span := trace.StartSpan(context.Background(), "tmpop/BeginBlock")
	defer span.End()

	t.currentHeader = &req.Header
	if t.currentHeader == nil {
		log.Error("Cannot begin block without header")
		span.SetStatus(trace.Status{Code: monitoring.InvalidArgument, Message: "Cannot begin block without header"})
		return abci.ResponseBeginBlock{}
	}

	stats.Record(ctx, blockCount.M(1), txPerBlock.M(int64(t.currentHeader.NumTxs)))

	// If the AppHash of the previous block is present in this block's header,
	// consensus has been formed around it.
	// This AppHash will never be denied in a future block so we can add
	// evidence to the links that were added in the previous blocks.
	if t.lastBlock.AppHash.EqualsBytes(t.currentHeader.AppHash) {
		t.addTendermintEvidence(ctx, &req.Header)
	} else {
		errorMessage := fmt.Sprintf(
			"Unexpected AppHash in BeginBlock, got %x, expected %x",
			t.currentHeader.AppHash,
			*t.lastBlock.AppHash,
		)
		log.Warn(errorMessage)
		span.Annotate(nil, errorMessage)
	}

	t.state.UpdateValidators(ctx)

	t.state.previousAppHash = types.NewBytes32FromBytes(t.currentHeader.AppHash)

	return abci.ResponseBeginBlock{}
}

// DeliverTx implements github.com/tendermint/abci/types.Application.DeliverTx.
func (t *TMPop) DeliverTx(tx []byte) abci.ResponseDeliverTx {
	ctx, span := trace.StartSpan(context.Background(), "tmpop/DeliverTx")
	defer span.End()

	err := t.doTx(ctx, t.state.Deliver, tx)
	if !err.IsOK() {
		ctx, _ = tag.New(ctx, tag.Upsert(txStatus, "invalid"))
		stats.Record(ctx, txCount.M(1))
		return abci.ResponseDeliverTx{
			Code: err.Code,
			Log:  err.Log,
		}
	}

	ctx, _ = tag.New(ctx, tag.Upsert(txStatus, "valid"))
	stats.Record(ctx, txCount.M(1))
	return abci.ResponseDeliverTx{}
}

// CheckTx implements github.com/tendermint/abci/types.Application.CheckTx.
func (t *TMPop) CheckTx(tx []byte) abci.ResponseCheckTx {
	ctx, span := trace.StartSpan(context.Background(), "tmpop/CheckTx")
	defer span.End()

	err := t.doTx(ctx, t.state.Check, tx)
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
	ctx, span := trace.StartSpan(context.Background(), "tmpop/Commit")
	defer span.End()

	appHash, links, err := t.state.Commit(ctx)
	if err != nil {
		log.Errorf("Error while committing: %s", err)
		span.SetStatus(trace.Status{Code: monitoring.Internal, Message: err.Error()})
		return abci.ResponseCommit{}
	}

	if err := t.saveValidatorHash(ctx); err != nil {
		log.Errorf("Error while saving validator hash: %s", err)
		span.SetStatus(trace.Status{Code: monitoring.Internal, Message: err.Error()})
		return abci.ResponseCommit{}
	}

	if err := t.saveCommitLinkHashes(ctx, links); err != nil {
		log.Errorf("Error while saving committed link hashes: %s", err)
		span.SetStatus(trace.Status{Code: monitoring.Internal, Message: err.Error()})
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
	ctx, span := trace.StartSpan(context.Background(), "tmpop/Query")
	span.AddAttributes(trace.StringAttribute("Path", reqQuery.Path))
	defer span.End()

	if reqQuery.Height != 0 {
		resQuery.Code = CodeTypeInternalError
		resQuery.Log = "tmpop only supports queries on latest commit"
		span.SetStatus(trace.Status{Code: monitoring.InvalidArgument, Message: resQuery.Log})
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

		result, err = t.adapter.GetSegment(ctx, linkHash)

	case GetEvidences:
		linkHash := &types.Bytes32{}
		if err = linkHash.UnmarshalJSON(reqQuery.Data); err != nil {
			break
		}

		result, err = t.adapter.GetEvidences(ctx, linkHash)

	case AddEvidence:
		evidence := &struct {
			LinkHash *types.Bytes32
			Evidence *cs.Evidence
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
		span.SetStatus(trace.Status{Code: monitoring.Internal, Message: resQuery.Log})
		return
	}

	if result != nil {
		resBytes, err := json.Marshal(result)
		if err != nil {
			resQuery.Code = CodeTypeInternalError
			resQuery.Log = err.Error()
			span.SetStatus(trace.Status{Code: monitoring.Internal, Message: resQuery.Log})
		}

		resQuery.Value = resBytes
	}

	return
}

func (t *TMPop) doTx(ctx context.Context, createLink func(context.Context, *cs.Link) *ABCIError, txBytes []byte) *ABCIError {
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
		return createLink(ctx, tx.Link)
	default:
		return &ABCIError{
			CodeTypeNotImplemented,
			fmt.Sprintf("Unexpected Tx type byte %X", tx.TxType),
		}
	}
}

// addTendermintEvidence computes and stores new evidence
func (t *TMPop) addTendermintEvidence(ctx context.Context, header *abci.Header) {
	ctx, span := trace.StartSpan(ctx, "tmpop/addTendermintEvidence")
	span.AddAttributes(trace.Int64Attribute("Height", header.Height))
	defer span.End()

	if t.tmClient == nil {
		log.Warn("TMPoP not connected to Tendermint Core. Evidence will not be generated.")
		span.SetStatus(trace.Status{Code: monitoring.Unavailable, Message: "TMPoP not connected to Tendermint Core."})
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
		span.SetStatus(trace.Status{Code: monitoring.FailedPrecondition})
		return
	}

	linkHashes, err := t.getCommitLinkHashes(ctx, evidenceHeight)
	if err != nil {
		log.Warnf("Could not get link hashes for block %d. Evidence will not be generated.", header.Height)
		span.SetStatus(trace.Status{Code: monitoring.Unavailable, Message: "Could not get link hashes"})
		return
	}

	if len(linkHashes) == 0 {
		return
	}

	span.AddAttributes(trace.Int64Attribute("LinksCount", int64(len(linkHashes))))

	validatorHash, err := t.getValidatorHash(ctx, evidenceHeight)
	if err != nil {
		log.Warnf("Could not get validator hash for block %d. Evidence will not be generated.", header.Height)
		span.SetStatus(trace.Status{Code: monitoring.Internal, Message: "Could not get validator hash"})
		return
	}

	evidenceBlock, err := t.tmClient.Block(ctx, evidenceHeight)
	if err != nil {
		log.Warnf("Could not get block %d header: %v", header.Height, err)
		span.SetStatus(trace.Status{Code: monitoring.Unavailable, Message: "Could not get block"})
		return
	}

	evidenceNextBlock, err := t.tmClient.Block(ctx, evidenceHeight+1)
	if err != nil {
		log.Warnf("Could not get next block %d header: %v", header.Height, err)
		span.SetStatus(trace.Status{Code: monitoring.Unavailable, Message: "Could not get next block"})
		return
	}

	evidenceLastBlock, err := t.tmClient.Block(ctx, evidenceHeight+2)
	if err != nil {
		log.Warnf("Could not get last block %d header: %v", header.Height, err)
		span.SetStatus(trace.Status{Code: monitoring.Unavailable, Message: "Could not get last block"})
		return
	}

	if len(evidenceNextBlock.Votes) == 0 || len(evidenceLastBlock.Votes) == 0 {
		log.Warnf("Block %d isn't signed by validator nodes. Evidence will not be generated.", header.Height)
		span.SetStatus(trace.Status{Code: monitoring.FailedPrecondition, Message: "Votes are missing"})
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
		span.SetStatus(trace.Status{Code: monitoring.Internal, Message: "Could not create merkle tree"})
		return
	}

	merkleRoot := types.NewBytes32FromBytes(merkle.Root())

	appHash, err := ComputeAppHash(evidenceBlockAppHash, validatorHash, merkleRoot)
	if err != nil {
		log.Warnf("Could not compute app hash for block %d. Evidence will not be generated.", header.Height)
		span.SetStatus(trace.Status{Code: monitoring.Internal, Message: "Could not compute app hash"})
		return
	}
	if !appHash.EqualsBytes(evidenceNextBlock.Header.AppHash) {
		log.Warnf("App hash %x of block %d doesn't match the header's: %x. Evidence will not be generated.",
			*appHash,
			header.Height,
			header.AppHash)
		span.SetStatus(trace.Status{Code: monitoring.FailedPrecondition, Message: "AppHash mismatch"})
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
				},
			}

			if err := t.adapter.AddEvidence(ctx, linkHash, evidence); err != nil {
				log.Warnf("Evidence could not be added to local store: %v", err)
				span.Annotatef(nil, "Evidence for %x could not be added: %s", *linkHash, err.Error())
			}

			newEvidences[linkHash] = evidence
		}
	}

	t.eventsManager.AddSavedEvidences(newEvidences)
}
