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
	"github.com/stratumn/sdk/cs"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
	"github.com/stratumn/sdk/validator"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire"
	merkle "github.com/tendermint/merkleeyes/iavl"
)

// tmpopLastBlockKey is the database key where last block information are saved.
var tmpopLastBlockKey = []byte("tmpop:lastblock")

// lastBlock saves the information of the last block commited for Core/App Handshake on crash/restart.
type lastBlock struct {
	AppHash    []byte
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

	// The DB cache size.
	CacheSize int

	// JSON schema rules definition
	ValidatorFilename string
}

// TMPop is the type of the application that implements github.com/tendermint/abci/types.Application,
// the tendermint socket protocol (ABCI).
type TMPop struct {
	abci.BaseApplication

	state                *State
	adapter              store.Adapter
	lastBlock            *lastBlock
	config               *Config
	header               *abci.Header
	currentBlockSegments []*types.Bytes32
}

const (
	// Name of the Tendermint Application.
	Name = "TMPop"

	// Description of this Tendermint Application.
	Description = "Agent Store in a Blockchain"

	// DefaultCacheSize is the default size of the DB cache.
	DefaultCacheSize = 0
)

const (
	// CodeTypeValidation is the ABCI error code for a validation error.
	CodeTypeValidation abci.CodeType = 400
)

// New creates a new instance of a TMPop.
func New(a store.Adapter, config *Config) (*TMPop, error) {
	db := NewDBAdapter(a)

	initalized, err := a.GetValue(tmpopLastBlockKey)
	if err != nil {
		return nil, err
	}

	// Load Tree
	tree := merkle.NewIAVLTree(config.CacheSize, db)

	if initalized == nil {
		log.Debug("No existing db, creating new db")
		saveLastBlock(a, lastBlock{
			AppHash: tree.Save(),
			Height:  0,
		})
	} else {
		log.Debug("Loading existing db")
	}

	lastBlock, err := readLastBlock(a)
	if err != nil {
		return nil, err
	}

	// TODO: At this point we might want to clean segments whose evidence is not complete
	// since we cannot be certain they will be delivered again, for instance if +2/3 nodes crashed
	// during the commit phase, and the transactions are not in the memory pool of anybody.
	tree.Load(lastBlock.AppHash)

	s, err := NewState(tree, a)
	if err != nil {
		return nil, err
	}

	return &TMPop{
		state:     s,
		adapter:   a,
		lastBlock: lastBlock,
		config:    config,
		header:    lastBlock.LastHeader,
	}, nil
}

// Info implements github.com/tendermint/abci/types.Application.Info.
func (t *TMPop) Info(req abci.RequestInfo) abci.ResponseInfo {
	return abci.ResponseInfo{
		Data:             Name,
		Version:          t.config.Version,
		LastBlockHeight:  t.lastBlock.Height,
		LastBlockAppHash: t.lastBlock.AppHash,
	}
}

// SetOption implements github.com/tendermint/abci/types.Application.SetOption.
func (t *TMPop) SetOption(key string, value string) (log string) {
	return "No options are supported yet"
}

// BeginBlock implements github.com/tendermint/abci/types.Application.BeginBlock.
func (t *TMPop) BeginBlock(req abci.RequestBeginBlock) {
	t.header = req.GetHeader()
	if t.header == nil {
		log.Error("Cannot begin block without header")
		return
	}

	// If the AppHash is present in this block, consensus has been formed around
	// it. Even though the current block might not get Committed in the end, that
	// would only be du to the transactions it contains. This AppHash will never be
	// denied in a future Block.
	if bytes.Compare(t.state.Committed().Hash(), t.header.AppHash) == 0 {
		for _, lh := range t.currentBlockSegments {
			err := t.addOriginalEvidence(lh)
			if err != nil {
				log.Warnf("Unexpected error while adding evidence to segment %x: %v", lh, err)
			}
		}
	} else {
		log.Warnf("Unexpected AppHash in BeginBlock, got %x, expected %x", t.header.AppHash, t.lastBlock.AppHash)
	}

	// We have been waiting for the BeginBlock callback to save the new LastBlockHeight and
	// LastBlockAppHeight to be absolutely sure that App has not saved a State it has
	// not communicated to Core. That would prevent the Handshake to succeed.
	t.lastBlock.Height = t.header.Height - 1
	t.lastBlock.AppHash = t.state.Committed().Hash()
	t.lastBlock.LastHeader = t.header

	saveLastBlock(t.adapter, *t.lastBlock)

	t.currentBlockSegments = nil
}

// DeliverTx implements github.com/tendermint/abci/types.Application.DeliverTx.
func (t *TMPop) DeliverTx(tx []byte) abci.Result {
	snapshot := t.state.Append()
	return t.doTx(snapshot, tx)
}

// CheckTx implements github.com/tendermint/abci/types.Application.CheckTx.
func (t *TMPop) CheckTx(tx []byte) abci.Result {
	snapshot := t.state.Check()
	return t.doTx(snapshot, tx)
}

// Commit implements github.com/tendermint/abci/types.Application.Commit.
// It actually commits the current state in the Store.
func (t *TMPop) Commit() abci.Result {
	appHash, err := t.state.Commit()
	if err != nil {
		return abci.NewError(abci.CodeType_InternalError, err.Error())
	}

	if t.state.Committed().Size() == 0 {
		return abci.NewResultOK(nil, "Empty hash for empty tree")
	}
	return abci.NewResultOK(appHash, "")
}

// Query implements github.com/tendermint/abci/types.Application.Query.
func (t *TMPop) Query(reqQuery abci.RequestQuery) (resQuery abci.ResponseQuery) {
	commit := t.state.Committed()

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
		var value *cs.Segment
		var proof []byte
		value, proof, err = commit.GetSegment(linkHash)

		t.addCurrentProof(value, proof)

		var valueByte []byte
		valueByte, err = json.Marshal(value)
		if err != nil {
			break
		}

		resQuery.Value = valueByte
		resQuery.Proof = proof

	case FindSegments:
		filter := &store.SegmentFilter{}
		if err := json.Unmarshal(reqQuery.Data, filter); err != nil {
			break
		}
		var values cs.SegmentSlice
		var proofs [][]byte
		values, proofs, err = commit.FindSegments(filter)

		for i, s := range values {
			t.addCurrentProof(s, proofs[i])
		}

		var valuesByte []byte
		valuesByte, err = json.Marshal(values)
		if err != nil {
			break
		}
		resQuery.Value = valuesByte

		var proofsByte []byte
		proofsByte, err = json.Marshal(proofs)
		if err != nil {
			break
		}

		resQuery.Proof = proofsByte

	case GetMapIDs:
		filter := &store.MapFilter{}
		if err := json.Unmarshal(reqQuery.Data, filter); err != nil {
			break
		}
		result, err = commit.GetMapIDs(filter)

	case GetValue:
		var key []byte
		if err := json.Unmarshal(reqQuery.Data, &key); err != nil {
			break
		}

		value, proof, _ := commit.Proof(key)

		resQuery.Value = value
		resQuery.Proof = proof
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

func (t *TMPop) doTx(snapshot *Snapshot, txBytes []byte) (result abci.Result) {
	if len(txBytes) == 0 {
		return abci.ErrEncodingError.SetLog("Tx length cannot be zero")
	}
	tx, res := unmarshallTx(txBytes)
	var err error

	switch tx.TxType {
	case SaveSegment:
		if res.IsErr() {
			return res
		}
		if res = t.checkSegment(snapshot, tx.Segment); res.IsErr() {
			return res
		}
		t.currentBlockSegments = append(t.currentBlockSegments, tx.Segment.GetLinkHash())

		if t.header != nil {
			tx.Segment.SetEvidence(
				map[string]interface{}{
					"state":        cs.PendingEvidence,
					"transactions": map[string]string{fmt.Sprintf("[%s]:[%s]", Name, t.header.ChainId): fmt.Sprintf("%v", t.header.Height)},
				})
		}

		err = snapshot.SetSegment(tx.Segment)
	case DeleteSegment:
		var segment *cs.Segment
		var found bool
		segment, found, err = snapshot.DeleteSegment(tx.LinkHash)
		var valueByte []byte
		valueByte, err = json.Marshal(segment)
		if err != nil {
			break
		}
		if found == true {
			result.Data = valueByte
		}
	case SaveValue:
		snapshot.SaveValue(tx.Key, tx.Value)
	case DeleteValue:
		value, found := snapshot.DeleteValue(tx.Key)
		if found == true {
			result.Data = value
		}
	default:
		return abci.ErrUnknownRequest.SetLog(fmt.Sprintf("Unexpected Tx type byte %X", tx.TxType))
	}

	if err != nil {
		result.Code = abci.CodeType_InternalError
		result.Log = err.Error()
		return
	}

	result.Code = abci.CodeType_OK
	return

}

func (t *TMPop) checkSegment(snapshot *Snapshot, segment *cs.Segment) abci.Result {
	err := segment.Validate()
	if err != nil {
		return abci.NewError(
			CodeTypeValidation,
			fmt.Sprintf("Segment validation failed %v: %v", segment, err),
		)
	}

	// TODO: in production do not reload validation rules each time a new segment is created
	// Use instead notification mechanisms
	if t.config.ValidatorFilename != "" {
		rootValidator := validator.NewRootValidator(t.config.ValidatorFilename, true)
		err = rootValidator.Validate(snapshot.segments, segment)

		if err != nil {
			return abci.NewError(
				CodeTypeValidation,
				fmt.Sprintf("Segment validation failed %v: %v", segment, err),
			)
		}
	} else {
		log.Debug("No custom validation configured")
	}

	return abci.OK
}

func unmarshallTx(txBytes []byte) (*Tx, abci.Result) {
	tx := &Tx{}

	if err := json.Unmarshal(txBytes, tx); err != nil {
		return nil, abci.NewError(abci.CodeType_InternalError, err.Error())
	}

	return tx, abci.NewResultOK([]byte{}, "ok")
}

func readLastBlock(a store.Adapter) (*lastBlock, error) {
	lBytes, err := a.GetValue(tmpopLastBlockKey)
	if err != nil {
		return nil, err
	}

	var l lastBlock
	if lBytes == nil {
		return &l, nil
	}
	err = wire.ReadBinaryBytes(lBytes, &l)
	if err != nil {
		return nil, err
	}

	return &l, nil
}

func saveLastBlock(a store.Adapter, l lastBlock) {
	a.SaveValue(tmpopLastBlockKey, wire.BinaryBytes(l))
}

// addOriginalEvidence adds the Evidence to the segment.
// It should only be called when the header with the signed AppHash that includes
// this segment is available.
func (t *TMPop) addOriginalEvidence(lh *types.Bytes32) error {
	s, err := t.adapter.GetSegment(lh)
	if err != nil {
		return err
	}
	if s == nil {
		log.Debug("No segment found with linkHash %v", lh)
		return nil
	}
	_, proof, err := t.state.Committed().GetSegment(s.GetLinkHash())
	if err != nil {
		return err
	}

	iavlProof, err := merkle.ReadProof(proof)
	if err != nil {
		return err
	}
	e := s.GetEvidence()

	e["originalProof"] = iavlProof
	e["state"] = cs.CompleteEvidence
	e["originalHeader"] = *t.header

	return t.adapter.SaveSegment(s)
}

func (t *TMPop) addCurrentProof(s *cs.Segment, proof []byte) error {
	iavlProof, err := merkle.ReadProof(proof)
	if err != nil {
		return err
	}
	e := s.GetEvidence()

	e["currentHeader"] = *t.header
	e["currentProof"] = iavlProof

	return nil
}
