[![Logo](logo.png)](https://indigocore.org)

# IndigoCore

[Stratumn](https://stratumn.com)'s open-source SDK to create Indigo applications and networks.

[![GoDoc](https://godoc.org/github.com/stratumn/go-indigocore?status.svg)](https://godoc.org/github.com/stratumn/go-indigocore)
[![Build Status](https://semaphoreci.com/api/v1/stratumn/go-indigocore/branches/master/badge.svg)](https://semaphoreci.com/stratumn/go-indigocore)
[![codecov](https://codecov.io/gh/stratumn/go-indigocore/branch/master/graph/badge.svg)](https://codecov.io/gh/stratumn/go-indigocore)
[![Go Report Card](https://goreportcard.com/badge/github.com/stratumn/go-indigocore)](https://goreportcard.com/report/github.com/stratumn/go-indigocore)
[![Join the chat at https://gitter.im/stratumn/sdk](https://badges.gitter.im/stratumn/sdk.svg)](https://gitter.im/stratumn/sdk?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

---

This repository includes tools to build [Proof of Process Networks](https://proofofprocess.org) using [Indigo Core](https://indigocore.org).

To get started, visit the Indigo Core website:
https://indigocore.org

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

If you want to release a new version of IndigoCore, here is what you need to do.
You need to install:

- [Docker](https://www.docker.com/)
- [Keybase](https://keybase.io/)
- [github-release](https://github.com/aktau/github-release/releases/)

You'll need to add your PGP public key to the [Indigo CLI](https://github.com/stratumn/indigo-cli) in the pubkey.go file.

Then at the root of the repo:

- Update the CHANGELOG file
- Create a branch named vA.B.x (for example: 0.1.x) from master
- On this new branch, create a VERSION file that contains the version (for example: 0.1.0)
- Set the pre-release flag in PRERELEASE if needed
- Run _make release_ (this will create the tag, build the binaries and the docker images, push the docker images and publish a release on Github)

## License

Copyright 2017 Stratumn SAS. All rights reserved.

Unless otherwise noted, the source files are distributed under the Apache
License 2.0 found in the LICENSE file.

Third party dependencies included in the vendor directory are distributed under
their respective licenses.
