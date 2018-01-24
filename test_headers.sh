#!/bin/sh

set -o errexit # exit on error
set -o nounset # errors on unset variables

RET=0
for f in $@
do
    if ! $(grep "$(head -n 1 LICENSE_HEADER)" --quiet $f); then
        echo "Missing header for $f"
        RET=1
    fi
done
exit $RET