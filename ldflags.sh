#!/usr/bin/env bash

if git branch &> /dev/null ; then
    sha=$(git rev-parse HEAD)
    echo "-ldflags=\"-X 'main.commit=$sha'\""
fi
