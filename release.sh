#!/usr/bin/env bash

set -eu -o pipefail
shopt -s failglob

[ $# -ne 2 ] && { echo "Usage: $0 VERSION NAME"; exit 1; }

RELEASE_VERSION=$1
RELEASE_NAME=$2
RELEASE_DIR=$(pwd)/dist

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

build() {
    goos=$1
    goarch=$2
    dir=${RELEASE_DIR}/aretext_${RELEASE_VERSION}_${goos}_${goarch}
    echo "$dir"
    mkdir -p "$dir"
    make build GO_OS="$1" GO_ARCH="$2" GO_BUILD_FLAGS="-trimpath" GO_OUTPUT="$dir/aretext"

    cp LICENSE "$dir/"
    cp -r docs "$dir/"

    archive_cwd=$(dirname "$dir")
    archive_src=$(basename "$dir")
    archive_dst=$RELEASE_DIR/${archive_src}.tar.gz
    archive_ts=$(git log -1 --format=%ct)
    echo "$archive_dst"
    tar --mtime "@${archive_ts}" -czf "$archive_dst" -C "$archive_cwd" "$archive_src"
}

build linux amd64
build linux arm64
build darwin arm64
build freebsd amd64
build freebsd arm64

checksums=aretext_${RELEASE_VERSION}_checksums.txt
echo "$RELEASE_DIR/$checksums"
(cd "$RELEASE_DIR" && shasum -a 256 ./*.tar.gz > "$checksums")

echo "verifying checksums"
(cd "$RELEASE_DIR" && shasum -c "$checksums")

echo "publishing git tag"
git push origin "$release_tag"
