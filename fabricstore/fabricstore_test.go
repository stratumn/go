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

package fabricstore

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stratumn/sdk/cs/cstesting"
	"github.com/stratumn/sdk/store"
	"github.com/stratumn/sdk/types"
)

var (
	err             error
	fabricstore     *FabricStore
	mockFabricstore *FabricStore
	test            *testing.T
	linkHash        *types.Bytes32
	process         string
	integration     = flag.Bool("integration", false, "Run integration tests")
	channelID       = flag.String("channelID", "mychannel", "channelID")
	chaincodeID     = flag.String("chaincodeID", "pop", "chaincodeID")
	configFile      = flag.String("configFile", os.Getenv("GOPATH")+"/src/github.com/stratumn/sdk/fabricstore/integration/config-client.yaml", "Absolute path to network config file")
	version         = "0.1.0"
	commit          = "00000000000000000000000000000000"
)

func TestMain(m *testing.M) {
	flag.Parse()
	mockFabricstore = NewTestClient()
	if *integration {
		config := &Config{
			ChannelID:   *channelID,
			ChaincodeID: *chaincodeID,
			ConfigFile:  *configFile,
			Version:     version,
			Commit:      commit,
		}
		fabricstore, err = New(config)
		if err != nil {
			fmt.Println("Could not initiate client")
			os.Exit(0)
		}
	}
	result := m.Run()
	os.Exit(result)
}

// Unit tests
func Test_SaveSegment(t *testing.T) {
	segment := cstesting.RandomSegment()
	err := mockFabricstore.SaveSegment(segment)
	if err != nil {
		t.FailNow()
	}
}

func Test_GetSegment(t *testing.T) {
	segment := cstesting.RandomSegment()
	_, err := mockFabricstore.GetSegment(segment.GetLinkHash())
	if err != nil {
		t.FailNow()
	}
}

func Test_DeleteSegment(t *testing.T) {
	segment := cstesting.RandomSegment()
	_, err := mockFabricstore.DeleteSegment(segment.GetLinkHash())
	if err != nil {
		t.FailNow()
	}
}

func Test_FindSegments(t *testing.T) {
	segmentFilter := &store.SegmentFilter{}
	_, err := mockFabricstore.FindSegments(segmentFilter)
	if err != nil {
		t.FailNow()
	}
}

func Test_SaveValue(t *testing.T) {
	err := mockFabricstore.SaveValue([]byte("key"), []byte("value"))
	if err != nil {
		t.FailNow()
	}
}

func Test_GetValue(t *testing.T) {
	_, err := mockFabricstore.GetValue([]byte("key"))
	if err != nil {
		t.FailNow()
	}
}

func Test_DeleteValue(t *testing.T) {
	_, err := mockFabricstore.DeleteValue([]byte("key"))
	if err != nil {
		t.FailNow()
	}
}

// Integration tests (go test -integration)
func Test_SaveValueIntegration(t *testing.T) {
	if *integration {
		err := fabricstore.SaveValue([]byte("key"), []byte("value"))
		if err != nil {
			fmt.Println("Could not save value", err.Error())
			t.FailNow()
		}
	}
}

func Test_GetValueIntegration(t *testing.T) {
	if *integration {
		value, err := fabricstore.GetValue([]byte("key"))
		if err != nil {
			fmt.Println("Could not get saved value", err.Error())
			t.FailNow()
		}
		if string(value) != "value" {
			fmt.Println("Unexpected value in store")
			t.FailNow()
		}
	}
}

func Test_DeleteValueIntegration(t *testing.T) {
	if *integration {
		_, err := fabricstore.DeleteValue([]byte("key"))
		if err != nil {
			fmt.Println("Error while deleting value", err.Error())
			t.FailNow()
		}
	}
}

func Test_GetValueMissingIntegration(t *testing.T) {
	if *integration {
		value, err := fabricstore.GetValue([]byte("key"))
		if err != nil {
			fmt.Println("Did not expect error getting non existent key", err.Error())
			t.FailNow()
		}
		if value != nil {
			fmt.Println("Value:", value, "not deleted")
			t.FailNow()
		}
	}
}

func Test_SaveSegmentIntegration(t *testing.T) {
	if *integration {
		segment := cstesting.RandomSegment()
		linkHash = segment.GetLinkHash()
		err := fabricstore.SaveSegment(segment)
		if err != nil {
			fmt.Println("Could not save segment", err.Error())
			t.FailNow()
		}
	}
}

func Test_GetSegmentIntegration(t *testing.T) {
	if *integration {
		segment, err := fabricstore.GetSegment(linkHash)
		if err != nil {
			fmt.Println("Could not get segment", err.Error())
			t.FailNow()
		}

		if segment.GetLinkHashString() != linkHash.String() {
			fmt.Println("Did not retrieve expected segment")
			t.FailNow()
		}
	}
}

func Test_DeleteSegmentIntegration(t *testing.T) {
	if *integration {
		_, err = fabricstore.DeleteSegment(linkHash)
		if err != nil {
			fmt.Println("Could not delete segment", err.Error())
			t.FailNow()
		}
	}
}

func Test_GetSegmentMissingIntegration(t *testing.T) {
	if *integration {
		segment, err := fabricstore.GetSegment(linkHash)
		if err != nil {
			fmt.Println("Could not get segment", err.Error())
			t.FailNow()
		}

		if segment != nil {
			fmt.Println("Expected nil segment")
			t.FailNow()
		}
	}
}

func Test_FindSegmentsIntegration(t *testing.T) {
	if *integration {
		segment1 := cstesting.RandomSegment()
		segment2 := cstesting.RandomBranch(segment1)
		segment3 := cstesting.RandomSegment()

		delete(segment1.Link.Meta, "prevLinkHash")
		delete(segment3.Link.Meta, "prevLinkHash")

		err := fabricstore.SaveSegment(segment1)
		err = fabricstore.SaveSegment(segment2)
		err = fabricstore.SaveSegment(segment3)

		if err != nil {
			fmt.Println("Could not save segments", err.Error())
		}

		process = segment1.Link.GetProcess()

		segmentFilter := &store.SegmentFilter{
			MapIDs: []string{segment1.Link.GetMapID()},
			Pagination: store.Pagination{
				Limit: 20,
			},
		}

		segments, err := fabricstore.FindSegments(segmentFilter)
		if err != nil {
			fmt.Println("Could not find segments", err.Error())
			t.FailNow()
		}
		if len(segments) != 2 {
			fmt.Println("Expected 2 segments got", len(segments))
			t.FailNow()
		}

		segmentFilter.MapIDs = []string{segment1.Link.GetMapID(), segment3.Link.GetMapID()}
		segments, err = fabricstore.FindSegments(segmentFilter)
		if err != nil {
			fmt.Println("Could not find segments", err.Error())
			t.FailNow()
		}
		if len(segments) != 3 {
			fmt.Println("Expected 3 segments got", len(segments))
			t.FailNow()
		}
	}
}

func Test_GetMapIDsIntegration(t *testing.T) {
	if *integration {
		mapFilter := &store.MapFilter{
			Process: process,
			Pagination: store.Pagination{
				Limit: 20,
			},
		}

		mapIds, err := fabricstore.GetMapIDs(mapFilter)

		if err != nil {
			fmt.Println("Could not find mapIds", err.Error())
			t.FailNow()
		}
		if len(mapIds) == 0 {
			fmt.Println("Expected at least one mapId")
			t.FailNow()
		}
	}
}
