// Copyright 2016 Stratumn SAS. All rights reserved.
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

// The command dummyfossilizer starts an HTTP server with a dummyfossilizer.
//
// Usage
//
// The following flags are available:
//
//	$ dummyfossilizer -h
//	Usage of dummyfossilizer:
//        -callbacktimeout duration
//          	callback requests timeout (default 10s)
//        -http string
//          	http address (default ":6000")
//        -maxdata int
//          	maximum data length (default 64)
//        -mindata int
//          	minimum data length (default 32)
//        -tlscert string
//          	TLS certificate file
//        -tlskey string
//          	TLS private key file
//        -workers int
//          	number of result workers (default 8)
//
// Docker
//
// A Docker image is available. To create a container, run:
//
//	$ docker run -p 6000:6000 stratumn/dummyfossilizer
package main
