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

package postgresstore

import (
	"context"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-core/monitoring/errorcode"
	"github.com/stratumn/go-core/store"
	"github.com/stratumn/go-core/types"
)

// AddEvidence implements github.com/stratumn/go-core/store.EvidenceWriter.AddEvidence.
func (s *scopedStore) AddEvidence(ctx context.Context, linkHash chainscript.LinkHash, evidence *chainscript.Evidence) error {
	data, err := chainscript.MarshalEvidence(evidence)
	if err != nil {
		return types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not add evidence")
	}

	_, err = s.stmts.AddEvidence.Exec(linkHash, evidence.Provider, data)
	if err != nil {
		return types.WrapError(err, errorcode.Unavailable, store.Component, "could not add evidence")
	}

	return nil
}

// GetEvidences implements github.com/stratumn/go-core/store.EvidenceReader.GetEvidences.
func (s *scopedStore) GetEvidences(ctx context.Context, linkHash chainscript.LinkHash) (types.EvidenceSlice, error) {
	var evidences types.EvidenceSlice

	rows, err := s.stmts.GetEvidences.Query(linkHash)
	if err != nil {
		return nil, types.WrapError(err, errorcode.Unavailable, store.Component, "could not get evidences")
	}

	for rows.Next() {
		var (
			data     []byte
			evidence *chainscript.Evidence
		)

		if err := rows.Scan(&data); err != nil {
			return nil, types.WrapError(err, errorcode.Internal, store.Component, "could not get evidence")
		}

		evidence, err = chainscript.UnmarshalEvidence(data)
		if err != nil {
			return nil, types.WrapError(err, errorcode.InvalidArgument, store.Component, "could not unmarshal evidence")
		}

		evidences = append(evidences, evidence)
	}

	return evidences, nil
}
