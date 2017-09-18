#!/bin/sh

version_file=`dirname $0`/VERSION

if [ -f $version_file ]; then
    cat $version_file
    exit 0
fi

current_tag=`git describe --tags`
if git tag -l | grep "^$current_tag\$"; then
    echo $current_tag
else
    git branch | grep '^\*' | awk '{ print $2}'
fi
