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

package validators

import (
	"github.com/pkg/errors"
)

// Participant errors.
var (
	ErrInvalidVotingPower     = errors.New("participant voting power is missing")
	ErrMissingParticipantName = errors.New("participant name is missing")
	ErrMissingParticipantKey  = errors.New("participant public key is missing")
	ErrUnknownUpdateType      = errors.New("unknown participant update type")
	ErrParticipantNotFound    = errors.New("participant not found")
)

// Participant in a decentralized network.
// Participants have the responsibility of voting for validation rules updates.
type Participant struct {
	Name      string `json:"name"`
	Power     uint   `json:"votingPower"`
	PublicKey []byte `json:"publicKey"`
}

// Validate participant data.
func (p Participant) Validate() error {
	if p.Power == 0 {
		return ErrInvalidVotingPower
	}

	if len(p.Name) == 0 {
		return ErrMissingParticipantName
	}

	if len(p.PublicKey) == 0 {
		return ErrMissingParticipantKey
	}

	return nil
}

// Available update types.
const (
	ParticipantUpsert = "upsert"
	ParticipantRemove = "remove"
)

// ParticipantUpdate operation to remove/update network participants.
type ParticipantUpdate struct {
	Type string `json:"updateType"`
	Participant
}

// Validate participant update data.
func (p ParticipantUpdate) Validate(current []*Participant) error {
	switch p.Type {
	case ParticipantUpsert:
		return p.Participant.Validate()
	case ParticipantRemove:
		for _, cur := range current {
			if cur.Name == p.Name {
				return nil
			}
		}

		return ErrParticipantNotFound
	default:
		return ErrUnknownUpdateType
	}
}
