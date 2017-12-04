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

package btc

import "testing"

func TestNetworkString(t *testing.T) {
	if got, want := NetworkTest3.String(), "bitcoin:test3"; got != want {
		t.Errorf("NetworkTest3.String() = %s want %s", got, want)
	}
}

func TestNetworkID(t *testing.T) {
	if got, want := NetworkTest3.ID(), byte(0x6F); got != want {
		t.Errorf(`NetworkTest3.String() = "%x" want "%x"`, got, want)
	}
	if got, want := NetworkMain.ID(), byte(0x00); got != want {
		t.Errorf(`NetworkTest3.String() = "%x" want "%x"`, got, want)
	}
}
