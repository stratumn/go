// Copyright 2017 Stratumn SAS. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// The command filetmpop starts a server.
//
// Usage:
//
//
//	$ filetmpop -h
// 		Usage of filetmpop:
//			-fast_sync
//				Fast blockchain syncing (default true)
//			-grpc_laddr string
//				GRPC listen address (BroadcastTx only). Port required
//			-log_level string
//				Log level (default "notice")
//			-moniker string
//				Node Name (default "anonymous")
//			-node_laddr string
//				Node listen address. (0.0.0.0:0 means any interface, any port) (default "tcp://0.0.0.0:46656")
//			-path string
//				path to directory where files are stored (default "/var/stratumn/filestore")
//			-pex
//				Enable Peer-Exchange (dev feature)
//			-rpc_laddr string
//				RPC listen address. Port required (default "tcp://0.0.0.0:46657")
//			-seeds string
//				Comma delimited host:port seed nodes
//			-skip_upnp
//				Skip UPNP configuration
//
// Docker:
//
// A Docker image is available. To create a container, run:
//
//	$ docker run -p 46656-46657:46656-46657 stratumn/filetmpop

package main
