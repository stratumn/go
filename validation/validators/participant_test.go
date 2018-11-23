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

package validators_test

import (
	"testing"

	"github.com/stratumn/go-core/validation/validationtesting"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
)

func TestParticipant(t *testing.T) {
	t.Run("Validate()", func(t *testing.T) {
		t.Run("missing name", func(t *testing.T) {
			p := &validators.Participant{
				Power:     3,
				PublicKey: []byte(validationtesting.AlicePublicKey),
			}

			err := p.Validate()
			assert.EqualError(t, err, validators.ErrMissingParticipantName.Error())
		})

		t.Run("missing public key", func(t *testing.T) {
			p := &validators.Participant{
				Name:  "alice",
				Power: 3,
			}

			err := p.Validate()
			assert.EqualError(t, err, validators.ErrMissingParticipantKey.Error())
		})

		t.Run("missing voting power", func(t *testing.T) {
			p := &validators.Participant{
				Name:      "alice",
				PublicKey: []byte(validationtesting.AlicePublicKey),
			}

			err := p.Validate()
			assert.EqualError(t, err, validators.ErrInvalidVotingPower.Error())
		})

		t.Run("valid participant", func(t *testing.T) {
			p := &validators.Participant{
				Name:      "alice",
				Power:     3,
				PublicKey: []byte(validationtesting.AlicePublicKey),
			}

			err := p.Validate()
			assert.NoError(t, err)
		})
	})
}

func TestParticipantUpdate(t *testing.T) {
	t.Run("Validate()", func(t *testing.T) {
		t.Run("unknown update type", func(t *testing.T) {
			p := &validators.ParticipantUpdate{
				Type: "unknown",
			}

			err := p.Validate(nil)
			assert.EqualError(t, err, validators.ErrUnknownUpdateType.Error())
		})

		t.Run("remove missing participant", func(t *testing.T) {
			p := &validators.ParticipantUpdate{
				Type: validators.ParticipantRemove,
				Participant: validators.Participant{
					Name: "alice",
				},
			}

			err := p.Validate(nil)
			assert.EqualError(t, err, validators.ErrParticipantNotFound.Error())
		})

		t.Run("valid remove update", func(t *testing.T) {
			p := &validators.ParticipantUpdate{
				Type: validators.ParticipantRemove,
				Participant: validators.Participant{
					Name: "alice",
				},
			}

			err := p.Validate([]*validators.Participant{&validators.Participant{Name: "alice"}})
			assert.NoError(t, err)
		})

		t.Run("add invalid participant", func(t *testing.T) {
			p := &validators.ParticipantUpdate{
				Type: validators.ParticipantUpsert,
				Participant: validators.Participant{
					Name:  "alice",
					Power: 3,
					// Missing public key
				},
			}

			err := p.Validate(nil)
			assert.EqualError(t, err, validators.ErrMissingParticipantKey.Error())
		})

		t.Run("valid add update", func(t *testing.T) {
			p := &validators.ParticipantUpdate{
				Type: validators.ParticipantUpsert,
				Participant: validators.Participant{
					Name:      "alice",
					Power:     3,
					PublicKey: []byte(validationtesting.AlicePublicKey),
				},
			}

			err := p.Validate(nil)
			assert.NoError(t, err)
		})
	})
}
