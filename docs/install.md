Install
=======

Supported Platforms
-------------------

| Platform | Status                |
|----------|-----------------------|
| Linux    | Fully supported       |
| \*BSD    | Will probably work    |
| macOS    | Will probably work    |
| Windows  | Supported on WSL only |

Stable Releases
---------------

### Official Binaries

You can download the official binaries from [the aretext releases page](https://github.com/aretext/aretext/releases).

### Arch Linux

aretext is available as an [AUR Package](https://aur.archlinux.org/packages/aretext-bin/). If you use [yay](https://github.com/Jguer/yay), run this to install it:

```shell
yay -S aretext
```

Unstable Builds
---------------

### From source

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
