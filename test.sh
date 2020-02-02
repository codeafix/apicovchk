#!/usr/bin/env bash

set -e
mkdir -p temp
echo "" > temp/coverage.txt

go test -coverprofile=temp/profile.out -covermode=atomic $d
if [ -f temp/profile.out ]; then
    cat temp/profile.out >> temp/coverage.txt
    rm temp/profile.out
fi