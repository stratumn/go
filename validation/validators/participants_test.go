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
	"context"
	"testing"

	"github.com/stratumn/go-chainscript"
	"github.com/stratumn/go-chainscript/chainscripttest"
	"github.com/stratumn/go-core/dummystore"
	"github.com/stratumn/go-core/testutil"
	"github.com/stratumn/go-core/validation/validationtesting"
	"github.com/stratumn/go-core/validation/validators"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newParticipantLinkBuilder(t *testing.T, step string) *chainscripttest.LinkBuilder {
	return chainscripttest.NewLinkBuilder(t).
		WithProcess(validators.GovernanceProcess).
		WithMapID(validators.ParticipantsMap).
		WithStep(step)
}

func newParticipantAccept(t *testing.T) *chainscripttest.LinkBuilder {
	return newParticipantLinkBuilder(t, validators.ParticipantsAcceptStep).
		WithDegree(1)
}

func newParticipantUpdate(t *testing.T, accepted *chainscript.Link, p ...*validators.ParticipantUpdate) *chainscripttest.LinkBuilder {
	lb := newParticipantLinkBuilder(t, validators.ParticipantsUpdateStep)
	if accepted != nil {
		lb.WithRef(t, accepted)
	}
	if len(p) > 0 {
		lb.WithData(t, p)
	}

	return lb
}

func newParticipantVote(t *testing.T, update *chainscript.Link, key []byte) *chainscripttest.LinkBuilder {
	lb := newParticipantLinkBuilder(t, validators.ParticipantsVoteStep).WithDegree(0)
	if update != nil {
		lb.WithParent(t, update)
	}

	if len(key) > 0 {
		lb.WithSignatureFromKey(t, key, "")
	}

	return lb
}

func TestParticipantsValidator(t *testing.T) {
	alice := &validators.Participant{
		Name:      "alice",
		Power:     3,
		PublicKey: []byte(validationtesting.AlicePublicKey),
	}
	bob := &validators.Participant{
		Name:      "bob",
		Power:     2,
		PublicKey: []byte(validationtesting.BobPublicKey),
	}
	carol := &validators.Participant{
		Name:      "carol",
		Power:     2,
		PublicKey: []byte(validationtesting.CarolPublicKey),
	}

	v := validators.NewParticipantsValidator()

	t.Run("Validate()", func(t *testing.T) {
		t.Run("rejects unknown step", func(t *testing.T) {
			l := newParticipantLinkBuilder(t, "pwn").Build()
			err := v.Validate(context.Background(), nil, l)
			testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidParticipantStep)
		})

		t.Run("accept", func(t *testing.T) {
			t.Run("missing participants", func(t *testing.T) {
				l := newParticipantAccept(t).Build()

				err := v.Validate(context.Background(), nil, l)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidParticipantData)
			})

			t.Run("invalid participant", func(t *testing.T) {
				l := newParticipantAccept(t).
					WithData(t, []*validators.Participant{
						&validators.Participant{
							Name:  "alice",
							Power: 3,
							// Missing public key
						},
					}).
					Build()

				err := v.Validate(context.Background(), nil, l)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidParticipantData)
			})

			t.Run("data is not a participants list", func(t *testing.T) {
				l := newParticipantAccept(t).
					WithData(t, map[string]string{
						"name": "alice",
						"role": "admin",
					}).
					Build()

				err := v.Validate(context.Background(), nil, l)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidParticipantData)
			})

			t.Run("out degree should be 1", func(t *testing.T) {
				l := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice}).
					WithDegree(3).
					Build()

				err := v.Validate(context.Background(), nil, l)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidAcceptParticipant)
			})

			t.Run("participants already initialized", func(t *testing.T) {
				init := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice}).
					Build()

				store := dummystore.New(&dummystore.Config{})
				_, err := store.CreateLink(context.Background(), init)
				require.NoError(t, err)

				// We can't add a new accept link with no parent if there's
				// already one.
				err = v.Validate(context.Background(), store, init)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrParticipantsAlreadyInitialized)
			})

			t.Run("initialize participants", func(t *testing.T) {
				init := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice, bob}).
					Build()

				err := v.Validate(context.Background(), dummystore.New(&dummystore.Config{}), init)
				require.NoError(t, err)
			})

			t.Run("update votes not on latest", func(t *testing.T) {
				ctx := context.Background()
				store := dummystore.New(&dummystore.Config{})

				// The link will be invalid because it tries to bypass the
				// latest accepted link (l2) by referencing valid votes on
				// the previous accepted link (l1) to try to bypass Bob's
				// signature.
				//
				// l1 <------ l2 <------- invalid
				//	\                       |
				//	 `----- u1 <----- v1 <--'

				l1 := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice}).
					Build()
				_, err := store.CreateLink(ctx, l1)
				require.NoError(t, err)

				l2 := newParticipantAccept(t).WithData(t, []*validators.Participant{alice, bob}).
					WithParent(t, l1).
					WithPriority(1).
					Build()
				_, err = store.CreateLink(ctx, l2)
				require.NoError(t, err)

				u1 := newParticipantUpdate(t, l1, &validators.ParticipantUpdate{
					Type:        validators.ParticipantUpsert,
					Participant: *carol,
				}).Build()
				_, err = store.CreateLink(ctx, u1)
				require.NoError(t, err)

				v1 := newParticipantVote(t, u1, []byte(validationtesting.AlicePrivateKey)).Build()
				_, err = store.CreateLink(ctx, v1)
				require.NoError(t, err)

				invalid := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice, carol}).
					WithParent(t, l2).
					WithPriority(2).
					WithRef(t, v1).
					Build()

				err = v.Validate(ctx, store, invalid)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidAcceptParticipant)
				assert.Contains(t, err.Error(), "latest accepted")
			})

			t.Run("invalid priority", func(t *testing.T) {
				ctx := context.Background()
				store := dummystore.New(&dummystore.Config{})

				a1 := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice}).
					Build()
				_, err := store.CreateLink(ctx, a1)
				require.NoError(t, err)

				u1 := newParticipantUpdate(t, a1, &validators.ParticipantUpdate{
					Type:        validators.ParticipantUpsert,
					Participant: *bob,
				}).Build()
				_, err = store.CreateLink(ctx, u1)
				require.NoError(t, err)

				v1 := newParticipantVote(t, u1, []byte(validationtesting.AlicePrivateKey)).Build()
				_, err = store.CreateLink(ctx, v1)
				require.NoError(t, err)

				a2 := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice, bob}).
					WithParent(t, a1).
					WithPriority(3).
					WithRef(t, v1).
					Build()

				err = v.Validate(ctx, store, a2)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidAcceptParticipant)
				assert.Contains(t, err.Error(), "priority")
			})

			t.Run("invalid new participants list", func(t *testing.T) {
				ctx := context.Background()
				store := dummystore.New(&dummystore.Config{})

				a1 := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice}).
					Build()
				_, err := store.CreateLink(ctx, a1)
				require.NoError(t, err)

				u1 := newParticipantUpdate(t, a1, &validators.ParticipantUpdate{
					Type:        validators.ParticipantUpsert,
					Participant: *bob,
				}).Build()
				_, err = store.CreateLink(ctx, u1)
				require.NoError(t, err)

				v1 := newParticipantVote(t, u1, []byte(validationtesting.AlicePrivateKey)).Build()
				_, err = store.CreateLink(ctx, v1)
				require.NoError(t, err)

				a2 := newParticipantAccept(t).
					// The update should only add Bob, not Carol.
					WithData(t, []*validators.Participant{alice, bob, carol}).
					WithParent(t, a1).
					WithPriority(1).
					WithRef(t, v1).
					Build()

				err = v.Validate(ctx, store, a2)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidAcceptParticipant)
				assert.Contains(t, err.Error(), "invalid participants list")
			})

			t.Run("update not enough votes", func(t *testing.T) {
				ctx := context.Background()
				store := dummystore.New(&dummystore.Config{})

				a1 := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice, bob}).
					Build()
				_, err := store.CreateLink(ctx, a1)
				require.NoError(t, err)

				u1 := newParticipantUpdate(t, a1, &validators.ParticipantUpdate{
					Type:        validators.ParticipantUpsert,
					Participant: *carol,
				}).Build()
				_, err = store.CreateLink(ctx, u1)
				require.NoError(t, err)

				v1 := newParticipantVote(t, u1, []byte(validationtesting.AlicePrivateKey)).Build()
				_, err = store.CreateLink(ctx, v1)
				require.NoError(t, err)

				a2 := newParticipantAccept(t).
					// The update should only add Bob, not Carol.
					WithData(t, []*validators.Participant{alice, bob, carol}).
					WithParent(t, a1).
					WithPriority(1).
					WithRef(t, v1).
					Build()

				err = v.Validate(ctx, store, a2)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidAcceptParticipant)
				assert.Contains(t, err.Error(), "voting power")
			})

			t.Run("update participants", func(t *testing.T) {
				ctx := context.Background()
				store := dummystore.New(&dummystore.Config{})

				dave := validators.Participant{
					Name:      "dave",
					Power:     1,
					PublicKey: []byte("this is a public key"),
				}

				a1 := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice, bob, carol}).
					Build()
				_, err := store.CreateLink(ctx, a1)
				require.NoError(t, err)

				u1 := newParticipantUpdate(t, a1,
					&validators.ParticipantUpdate{
						Type:        validators.ParticipantRemove,
						Participant: validators.Participant{Name: "carol"},
					},
					&validators.ParticipantUpdate{
						Type:        validators.ParticipantUpsert,
						Participant: dave,
					}).Build()
				_, err = store.CreateLink(ctx, u1)
				require.NoError(t, err)

				aliceVote := newParticipantVote(t, u1, []byte(validationtesting.AlicePrivateKey)).Build()
				_, err = store.CreateLink(ctx, aliceVote)
				require.NoError(t, err)

				bobVote := newParticipantVote(t, u1, []byte(validationtesting.BobPrivateKey)).Build()
				_, err = store.CreateLink(ctx, bobVote)
				require.NoError(t, err)

				a2 := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice, bob, &dave}).
					WithParent(t, a1).
					WithPriority(1).
					WithRef(t, aliceVote).
					WithRef(t, bobVote).
					Build()

				err = v.Validate(ctx, store, a2)
				assert.NoError(t, err)
			})
		})

		t.Run("update", func(t *testing.T) {
			ctx := context.Background()
			store := dummystore.New(&dummystore.Config{})
			accepted := newParticipantAccept(t).
				WithData(t, []*validators.Participant{alice, bob}).
				Build()
			_, err := store.CreateLink(ctx, accepted)
			require.NoError(t, err)

			t.Run("link data should contain participants updates", func(t *testing.T) {
				invalid := newParticipantUpdate(t, accepted).WithData(t, "not a participants list").Build()

				err := v.Validate(ctx, store, invalid)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidParticipantData)
			})

			t.Run("add invalid participant", func(t *testing.T) {
				missingKey := newParticipantUpdate(t, accepted, &validators.ParticipantUpdate{
					Type: validators.ParticipantUpsert,
					Participant: validators.Participant{
						Name:  "carol",
						Power: 2,
					},
				}).Build()

				err := v.Validate(ctx, store, missingKey)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidParticipantData)
			})

			t.Run("invalid update operation", func(t *testing.T) {
				invalidType := newParticipantUpdate(t, accepted, &validators.ParticipantUpdate{
					Type: "removeAll",
				}).Build()

				err := v.Validate(ctx, store, invalidType)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidParticipantData)
			})

			t.Run("remove missing participant", func(t *testing.T) {
				missingParticipant := newParticipantUpdate(t, accepted, &validators.ParticipantUpdate{
					Type: validators.ParticipantRemove,
					Participant: validators.Participant{
						Name: "carol",
					},
				}).Build()

				err := v.Validate(ctx, store, missingParticipant)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidParticipantData)
			})

			t.Run("should have no parent", func(t *testing.T) {
				addCarol := newParticipantUpdate(t, accepted, &validators.ParticipantUpdate{
					Type:        validators.ParticipantUpsert,
					Participant: *carol,
				}).WithParent(t, accepted).Build()

				err := v.Validate(ctx, store, addCarol)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidUpdateParticipant)
			})

			t.Run("should have unlimited children", func(t *testing.T) {
				addCarol := newParticipantUpdate(t, accepted, &validators.ParticipantUpdate{
					Type:        validators.ParticipantUpsert,
					Participant: *carol,
				}).WithDegree(3).Build()

				err := v.Validate(ctx, store, addCarol)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidUpdateParticipant)
			})

			t.Run("should reference an accept link", func(t *testing.T) {
				addCarol := newParticipantUpdate(t, nil, &validators.ParticipantUpdate{
					Type:        validators.ParticipantUpsert,
					Participant: *carol,
				}).Build()

				err := v.Validate(ctx, store, addCarol)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidUpdateParticipant)
			})

			t.Run("should not reference multiple links", func(t *testing.T) {
				addCarol := newParticipantUpdate(t, accepted, &validators.ParticipantUpdate{
					Type:        validators.ParticipantUpsert,
					Participant: *carol,
				}).WithRef(t, chainscripttest.RandomLink(t)).Build()

				err := v.Validate(ctx, store, addCarol)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidUpdateParticipant)
			})

			t.Run("should reference latest accept link", func(t *testing.T) {
				s2 := dummystore.New(&dummystore.Config{})
				a1 := newParticipantAccept(t).
					WithData(t, []*validators.Participant{alice}).
					Build()
				_, err := s2.CreateLink(ctx, a1)
				require.NoError(t, err)

				a2 := newParticipantAccept(t).
					WithParent(t, a1).
					WithPriority(a1.Meta.Priority+1).
					WithData(t, []*validators.Participant{alice, bob}).
					Build()
				_, err = s2.CreateLink(ctx, a2)
				require.NoError(t, err)

				addCarol := newParticipantUpdate(t, a1, &validators.ParticipantUpdate{
					Type:        validators.ParticipantUpsert,
					Participant: *carol,
				}).Build()

				err = v.Validate(ctx, s2, addCarol)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidUpdateParticipant)
			})

			t.Run("add and remove participants", func(t *testing.T) {
				updates := newParticipantUpdate(t, accepted,
					// Add Carol.
					&validators.ParticipantUpdate{
						Type:        validators.ParticipantUpsert,
						Participant: *carol,
					},
					// Update Alice's public key.
					&validators.ParticipantUpdate{
						Type: validators.ParticipantUpsert,
						Participant: validators.Participant{
							Name:      alice.Name,
							Power:     alice.Power,
							PublicKey: carol.PublicKey,
						},
					},
					// Remove Bob.
					&validators.ParticipantUpdate{
						Type: validators.ParticipantRemove,
						Participant: validators.Participant{
							Name: bob.Name,
						},
					},
				).Build()

				err := v.Validate(ctx, store, updates)
				assert.NoError(t, err)
			})
		})

		t.Run("vote", func(t *testing.T) {
			ctx := context.Background()
			store := dummystore.New(&dummystore.Config{})

			accepted := newParticipantAccept(t).
				WithData(t, []*validators.Participant{alice, bob}).
				Build()
			_, err := store.CreateLink(ctx, accepted)
			require.NoError(t, err)

			proposal := newParticipantUpdate(t, accepted, &validators.ParticipantUpdate{
				Type:        validators.ParticipantUpsert,
				Participant: *carol,
			}).Build()
			_, err = store.CreateLink(ctx, proposal)
			require.NoError(t, err)

			t.Run("missing parent", func(t *testing.T) {
				vote := newParticipantVote(t, nil, []byte(validationtesting.AlicePrivateKey)).Build()

				err := v.Validate(ctx, store, vote)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidVoteParticipant)
			})

			t.Run("parent is not an update proposal", func(t *testing.T) {
				vote := newParticipantVote(t, accepted, []byte(validationtesting.AlicePrivateKey)).Build()

				err := v.Validate(ctx, store, vote)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidVoteParticipant)
			})

			t.Run("missing signature", func(t *testing.T) {
				vote := newParticipantVote(t, proposal, nil).Build()

				err := v.Validate(ctx, store, vote)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidVoteParticipant)
			})

			t.Run("invalid out degree", func(t *testing.T) {
				vote := newParticipantVote(t, proposal, []byte(validationtesting.AlicePrivateKey)).
					WithDegree(3).
					Build()

				err := v.Validate(ctx, store, vote)
				testutil.AssertWrappedErrorEqual(t, err, validators.ErrInvalidVoteParticipant)
			})

			t.Run("provides valid signature", func(t *testing.T) {
				vote := newParticipantVote(t, proposal, []byte(validationtesting.AlicePrivateKey)).
					Build()

				err := v.Validate(ctx, store, vote)
				assert.NoError(t, err)
			})
		})
	})

	t.Run("ShouldValidate()", func(t *testing.T) {
		t.Run("ignores non-governance process", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess("not-governance").
				WithMapID(validators.ParticipantsMap).
				Build()
			assert.False(t, v.ShouldValidate(l))
		})

		t.Run("ignores non-participants map", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess(validators.GovernanceProcess).
				WithMapID("not-participants").
				Build()
			assert.False(t, v.ShouldValidate(l))
		})

		t.Run("validates governance participants", func(t *testing.T) {
			l := chainscripttest.NewLinkBuilder(t).
				WithProcess(validators.GovernanceProcess).
				WithMapID(validators.ParticipantsMap).
				Build()
			assert.True(t, v.ShouldValidate(l))
		})
	})

	t.Run("Hash()", func(t *testing.T) {
		h, err := v.Hash()
		require.NoError(t, err)
		assert.Nil(t, h)
	})
}
