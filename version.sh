#!/bin/sh

version_file=`dirname $0`/VERSION
generator_version=0

usage()
{
            echo "Usage $0 [-g]" 1>&2
            echo "Options:" 1>&2
            echo "  -g    display generator branch" 1>&2
            echo "  -h    print this help" 1>&2
            exit 1
}

while getopts "gh" option; do
    case $option in
        g)
            generator_version=1
            ;;
        *)
            usage
            ;;
    esac
done


if [ -f $version_file ]; then
    cat $version_file
    exit 0
fi

if [ $generator_version -eq 1 ]; then
    echo "master"
    exit 0
fi

current_tag=`git describe --tags`
if git tag -l | grep "^$current_tag\$"; then
    echo $current_tag
else
    git branch | grep '^\*' | awk '{ print $2}'
fi
