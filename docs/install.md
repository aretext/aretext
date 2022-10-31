Install
=======

Supported Platforms
-------------------

-	Linux
-	macOS
-	FreeBSD

Official Binaries
-----------------

You can download the official binaries for Linux, macOS, and FreeBSD from [the aretext releases page](https://github.com/aretext/aretext/releases).

### Linux x86 64-bit

```
VERSION=0.7.0
RELEASE=aretext_${VERSION}_linux_amd64
curl -LO https://github.com/aretext/aretext/releases/download/v$VERSION/$RELEASE.tar.gz
tar -zxvf $RELEASE.tar.gz
sudo cp $RELEASE/aretext /usr/local/bin/
```

### Linux ARM 64-bit

```
VERSION=0.7.0
RELEASE=aretext_${VERSION}_linux_arm64
curl -LO https://github.com/aretext/aretext/releases/download/v$VERSION/$RELEASE.tar.gz
tar -zxvf $RELEASE.tar.gz
sudo cp $RELEASE/aretext /usr/local/bin/
```

### macOS x86 64-bit

```
VERSION=0.7.0
RELEASE=aretext_${VERSION}_darwin_amd64
curl -LO https://github.com/aretext/aretext/releases/download/v$VERSION/$RELEASE.tar.gz
tar -zxvf $RELEASE.tar.gz
sudo cp $RELEASE/aretext /usr/local/bin/
```

### macOS ARM 64-bit

```
VERSION=0.7.0
RELEASE=aretext_${VERSION}_darwin_arm64
curl -LO https://github.com/aretext/aretext/releases/download/v$VERSION/$RELEASE.tar.gz
tar -zxvf $RELEASE.tar.gz
sudo cp $RELEASE/aretext /usr/local/bin/
```

Build From Source
-----------------

If you have [installed go](https://golang.org/doc/install), then you can build aretext from source:

```
mkdir -p $(go env GOPATH)/bin
git clone https://github.com/aretext/aretext.git
cd aretext
make install
```

This will install aretext in `$(go env GOPATH)/bin`, which you can add to your `$PATH` environment variable. If you use bash, put this line in your `~/.bashrc` or `~/.bash_profile`:

```
export PATH=$PATH:$(go env GOPATH)/bin
```

Packages
--------

If a package is not yet available for your platform, please consider creating one! We are looking for package maintainers on Debian, Fedora, Homebrew, Nix, and any other platform you may prefer!

### Arch Linux

aretext is available as an [AUR Package](https://aur.archlinux.org/packages/aretext-bin/). If you use [yay](https://github.com/Jguer/yay), run this to install it:

```shell
yay -S aretext
```
