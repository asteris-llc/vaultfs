# VaultFS

[![Build Status](https://travis-ci.org/asteris-llc/vaultfs.svg?branch=master)](https://travis-ci.org/asteris-llc/vaultfs)

VaultFS mounts arbitrary [Vault](https://vaultproject.io/) prefixes in a FUSE
filesystem. It also provides a Docker volume plugin to the do the same for your
containers.

<!-- markdown-toc start - Don't edit this section. Run M-x markdown-toc-generate-toc again -->
**Table of Contents**

- [VaultFS](#vaultfs)
- [Mounting](#mounting)
- [Docker](#docker)
- [License](#license)

<!-- markdown-toc end -->

# Installation

This project is in early development and has not reached 1.0. You will have to
build the binary yourself:

```shell
go get github.com/asteris-llc/vaultfs
env GOOS=linux go build github.com/asteris-llc/vaultfs
```

# Usage

VaultFS is one binary that can mount keys or run a Docker volume plugin to do so
for containers. Run `vaultfs --help` to see options not documented here.

## Mounting

```
Usage:
  vaultfs mount {mountpoint} [flags]

Flags:
  -a, --address="https://localhost:8200": vault address
  -i, --insecure[=false]: skip SSL certificate verification
  -r, --root="secret": root path for reads
  -t, --token="": vault token
```

To mount secrets, first create a mountpoint (`mkdir test`), then use `vaultfs`
to mount:

```shell
vaultfs mount --address=http://localhost:8200 -t 3a749a17-528e-e4b1-c28a-62e54f0098ae test
```

## Docker

```
Usage:
  vaultfs docker {mountpoint} [flags]

Flags:
  -a, --address="https://localhost:8200": vault address
  -i, --insecure[=false]: skip SSL certificate verification
  -s, --socket="/run/docker/plugins/vault.sock": socket address to communicate with docker
  -t, --token="": vault token
```

To start the Docker plugin, create a directory to hold mountpoints (`mkdir
test`), then use `vaultfs` to start the server. When Docker volumes request a
volume (`docker run --volume-driver vault --volume
{prefix}:/container/secret/path`), the plugin will create mountpoints and manage
FUSE servers automatically.

```shell
vaultfs docker --address=http://localhost:8200 -t 3a749a17-528e-e4b1-c28a-62e54f0098ae test
```

# License

VaultFS is licensed under an
[Apache 2.0 License](http://www.apache.org/licenses/LICENSE-2.0.html) (see also:
[LICENSE](LICENSE))
