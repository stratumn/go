#!/bin/bash

set -e

update-license-header() {
    perl -i -0pe 's/\/\/ Copyright .* Stratumn.*\n(\/\/.*\n)*/`cat LICENSE_HEADER`/ge' $1
}

for f in $(find . -name "*.go" | grep -v vendor); do
    echo Updating $f...
    update-license-header $f
done
