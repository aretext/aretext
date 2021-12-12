Install
=======

Supported Platforms
-------------------

| Platform | Status                |
|----------|-----------------------|
| Linux    | Fully supported       |
| macOS    | Fully supported       |
| Windows  | Supported on WSL only |
| \*BSD    | Will probably work    |

Official Binaries
-----------------

You can download the official binaries for Linux and macOS from [the aretext releases page](https://github.com/aretext/aretext/releases).

### Linux x86 64-bit

```
VERSION=0.3.0
RELEASE=aretext_${VERSION}_linux_amd64
curl -LO https://github.com/aretext/aretext/releases/download/v$VERSION/$RELEASE.tar.gz
tar -zxvf $RELEASE.tar.gz
sudo cp $RELEASE/aretext /usr/local/bin/
```

### Linux ARM 64-bit

```
VERSION=0.3.0
RELEASE=aretext_${VERSION}_linux_arm64
curl -LO https://github.com/aretext/aretext/releases/download/v$VERSION/$RELEASE.tar.gz
tar -zxvf $RELEASE.tar.gz
sudo cp $RELEASE/aretext /usr/local/bin/
```

### macOS x86 64-bit

```
VERSION=0.3.0
RELEASE=aretext_${VERSION}_darwin_amd64
curl -LO https://github.com/aretext/aretext/releases/download/v$VERSION/$RELEASE.tar.gz
tar -zxvf $RELEASE.tar.gz
sudo cp $RELEASE/aretext /usr/local/bin/
```

### macOS ARM 64-bit

```
VERSION=0.3.0
RELEASE=aretext_${VERSION}_darwin_arm64
curl -LO https://github.com/aretext/aretext/releases/download/v$VERSION/$RELEASE.tar.gz
tar -zxvf $RELEASE.tar.gz
sudo cp $RELEASE/aretext /usr/local/bin/
```

Build From Source
-----------------

If you have [installed go](https://golang.org/doc/install), then you can build aretext from source:

```
git clone https://github.com/aretext/aretext.git
cd aretext
make install
```

This will install aretext in `$HOME/go/bin`, which you can add to your `$PATH`:

```
export PATH=$PATH:$HOME/go/bin
```

Packages
--------

If a package is not yet available for your platform, please consider creating one! We are looking for package maintainers on Debian, Fedora, Homebrew, Nix, and any other platform you may prefer!

### Arch Linux

aretext is available as an [AUR Package](https://aur.archlinux.org/packages/aretext-bin/). If you use [yay](https://github.com/Jguer/yay), run this to install it:

```shell
yay -S aretext
```
