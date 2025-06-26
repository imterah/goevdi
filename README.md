# goevdi: EVDI bindings for Go
[![Build Status](https://git.terah.dev/imterah/goevdi/actions/workflows/build.yml/badge.svg)](https://git.terah.dev/imterah/goevdi/actions)
[![GoDoc](https://godoc.org/git.terah.dev/imterah/goevdi/libevdi?status.svg)](https://godoc.org/git.terah.dev/imterah/goevdi/libevdi)
[![Go Report Card](https://goreportcard.com/badge/git.terah.dev/imterah/goevdi/libevdi)](https://goreportcard.com/report/git.terah.dev/imterah/goevdi/libevdi)
[![Examples](https://img.shields.io/badge/learn%20by-examples-0077b3.svg?style=flat-square)](https://git.terah.dev/imterah/goevdi/src/branch/main/example)
[![License](https://img.shields.io/badge/license-GNU_Lesser_General_Public_License_2.1-green)](https://git.terah.dev/imterah/goevdi/src/branch/main/LICENSE.md)

The [Extensible Virtual Display Interface (EVDI)](https://github.com/DisplayLink/evdi) is a Linux kernel module that enables management of multiple screens, allowing user-space programs to take control over what happens with the image. It is essentially a virtual display you can add, remove and receive screen updates for, in an application that uses the `libevdi` library.

This package provides a Go wrapper for the `libevdi` library, allowing you to easily manage Linux virtual displays.

## Warning

This package is still in early development and may contain bugs. Use at your own risk. Additionally, due to nature of this being in early development, the API is unstable and may change between versions.

This package can work in production environments, but it is not recommended for production use until the API is stable.

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
