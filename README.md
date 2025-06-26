# goevdi: EVDI bindings for Go

The [Extensible Virtual Display Interface (EVDI)](https://github.com/DisplayLink/evdi) is a Linux kernel module that enables management of multiple screens, allowing user-space programs to take control over what happens with the image. It is essentially a virtual display you can add, remove and receive screen updates for, in an application that uses the `libevdi` library.

This package provides a Go wrapper for the `libevdi` library, allowing you to easily manage Linux virtual displays.

## How to Use

See the [Go documentation](https://pkg.go.dev/git.terah.dev/imterah/goevdi@v1.14.10/libevdi) for documentation on how to use this package. EVDI can be a bit confusing at first, so I'd recommend looking at the [example here](https://git.terah.dev/imterah/goevdi/src/branch/main/example/main.go).

## Installation

To install the package, run the following command:

```
go get git.terah.dev/imterah/goevdi/libevdi@latest
```

After that, you'd want to install the Linux kernel headers, and libdrm. Packages for Debian are:

```
sudo apt install linux-headers-$(uname -r) libdrm-dev
```

To install the EVDI kernel module, install this on Debian:

```
sudo apt install evdi-dkms
```
