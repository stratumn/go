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

package validationtesting

import "fmt"

// Sample test validation rules exported by this package.
var (
	AuctionJSONRules = CreateValidatorJSON("auction", AuctionJSONPKI, AuctionJSONStepsRules)
	ChatJSONRules    = CreateValidatorJSON("chat", ChatJSONPKI, ChatJSONStepsRules)
	TestJSONRules    = fmt.Sprintf(`{%s,%s}`, AuctionJSONRules, ChatJSONRules)
)

// CreateValidatorJSON formats a PKI and steps rules into a valid JSON
// configuration.
func CreateValidatorJSON(name, pki, stepsRules string) string {
	return fmt.Sprintf(`"%s": {"pki": %s,"steps": %s}`, name, pki, stepsRules)
}
