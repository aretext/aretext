#!/usr/bin/env bash

set -eu -o pipefail
shopt -s failglob

[ $# -ne 1 ] && { echo "Usage: $0 TEST"; exit 1; }

TEST=$1
OUT=$(pwd)/test

mkdir -p $OUT
cat > $OUT/test.txt << EOF
This is a test! $TEST
EOF

if ! grep "test" ${OUT}/test.txt;
    echo "could not find test"
then
    echo "found test";
    exit 1;
fi
