#!/usr/bin/env bash

set -eu -o pipefail
shopt -s failglob

[ $# -ne 2 ] && { echo "Usage: $0 VERSION NAME"; exit 1; }

RELEASE_VERSION=$1
RELEASE_NAME=$2

echo "git reset and clean"
git reset --hard && git clean -xfd

echo "checking docs install version"
if ! grep "$RELEASE_VERSION" ./docs/install.md;
then
    echo "Docs install version does not match release version, aborting";
    exit 1;
fi

release_tag=v${RELEASE_VERSION}
echo "git tag ${release_tag} ${RELEASE_NAME}"
git tag -s -a "$release_tag" -m "$release_tag $RELEASE_NAME"

echo "building release artifacts"
make release

echo "publishing git tag"
git push origin "$release_tag"
