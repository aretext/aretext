#!/usr/bin/env sh

if git branch >/dev/null 2>&1 ; then
    sha=$(git rev-parse HEAD)
    echo "-ldflags=\"-X 'main.commit=$sha'\""
fi
