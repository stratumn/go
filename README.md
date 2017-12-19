[![Logo](logo.png)](https://indigoframework.com)

# Stratumn SDK

[Stratumn](https://stratumn.com)'s SDK to create Indigo applications and networks.

[![GoDoc](https://godoc.org/github.com/stratumn/sdk?status.svg)](https://godoc.org/github.com/stratumn/sdk)
[![build status](https://travis-ci.org/stratumn/sdk.svg?branch=master)](https://travis-ci.org/stratumn/sdk)
[![codecov](https://codecov.io/gh/stratumn/sdk/branch/master/graph/badge.svg)](https://codecov.io/gh/stratumn/sdk)
[![Go Report Card](https://goreportcard.com/badge/github.com/stratumn/sdk)](https://goreportcard.com/report/github.com/stratumn/sdk)

---

The SDK includes tools to build [Proof of Process Networks](https://proofofprocess.org) using the [Indigo Framework](https://indigoframework.com).

To get started, visit the Indigo Framework website:
https://indigoframework.com

## Run tests

You need Docker to be able to run the tests. The images `couchstore:latest`, `rethink:latest` and
`postgres:latest` will be run automatically (and pulled from the docker hub if
you don't already have them locally).

Install dependencies:

```bash
go get -u github.com/golang/dep/cmd/dep
dep ensure
```

To manage dependencies, see [dep](https://github.com/golang/dep).

Run all tests:

```bash
make test
```

See test coverage in the browser:

```bash
make coverhtml
```

Run the linter:

```bash
go get -u github.com/golang/lint/golint
make lint
```

Build tagged docker images:

```bash
make docker_images
```

## Releasing a new version

If you want to release a new version of the Stratumn SDK, here is what you need to do.
You need to install:

* [Docker](https://www.docker.com/)
* [Keybase](https://keybase.io/)
* [github-release](https://github.com/aktau/github-release/releases/)

You'll need to add your PGP public key to strat/cmd/pubkey.go

Then at the root of the repo:

* Update the CHANGELOG file
* Create a branch named vA.B.x (for example: 0.1.x) from master
* On this new branch, create a VERSION file that contains the version (for example: 0.1.0)
* Set the pre-release flag in PRERELEASE if needed
* Run _make release_ (this will create the tag, build the binaries and the docker images, push the docker images and publish a release on Github)

## License

Copyright 2017 Stratumn SAS. All rights reserved.

Unless otherwise noted, the source files are distributed under the Apache
License 2.0 found in the LICENSE file.

Third party dependencies included in the vendor directory are distributed under
their respective licenses.