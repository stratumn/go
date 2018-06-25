#!/bin/sh

version_file=`dirname $0`/VERSION

usage()
{
            echo "Usage $0" 1>&2
            echo "Options:" 1>&2
            echo "  -h    print this help" 1>&2
            exit 1
}

while getopts "h" option; do
    case $option in
        *)
            usage
            ;;
    esac
done

if [ -f $version_file ]; then
    cat $version_file
    exit 0
fi

current_tag=`git describe --tags --exact-match 2> /dev/null`
if [ $? -eq 0 ]; then
    echo $current_tag
else
    git branch | grep '^\*' | awk '{ print $2}'
fi