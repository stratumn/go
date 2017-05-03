// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// The command strat contains various subcommands for the Stratumn SDK.
package main

import "github.com/stratumn/sdk/strat/cmd"

var (
	version = "0.1.0"
	commit  = "00000000000000000000000000000000"
)

func main() {
	cmd.Execute(version, commit)
}
