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

---

## Development

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

## License

Copyright 2017 Stratumn SAS. All rights reserved.

Unless otherwise noted, the source files are distributed under the Apache
License 2.0 found in the LICENSE file.

Third party dependencies included in the vendor directory are distributed under
their respective licenses.
