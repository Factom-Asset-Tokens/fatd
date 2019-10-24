![](https://png.icons8.com/ios-glyphs/200/5ECCDD/octahedron.png)![](https://png.icons8.com/color/64/3498db/golang.png)

This repo contains the Golang reference implementation of the Factom Asset
Tokens protocol. This repo provides two executables for interacting with FAT
chains, as well as several Golang packages for use by external programs.

# fatd - Factom Asset Token Daemon - Alpha

A daemon written in Golang that discovers new Factom Asset Tokens chains and
maintains their current state. The daemon provides a JSON-RPC 2.0 API for
accessing data about token chains.

### fat-cli
A CLI for creating new FAT chains, as well as exploring and making transactions
on existing FAT chains. See the output from `fat-cli --help` for more
information.

## Development Status

This implementation is now at a v1.0 release, which means that the public APIs,
both Golang and the JSON-RPC endpoint are version locked. This also means that
we believe the code to be stable and secure. That being said, the FAT system
and really the entire blockchain industry is experimental.

Please help us improve the code and the protocol by trying to break things and
reporting bugs! Thank you!


## Binaries

Pre-compiled binaries for Linux, Windows, and Mac x86\_64 systems can be found
on the [releases page.](https://github.com/Factom-Asset-Tokens/fatd/releases/)
However, building from source is very easy on most platforms.

## Install with Docker 🐳


Build the Docker image:

```bash
$ docker build -t fatd github.com/Factom-Asset-Tokens/fatd
```

Create a volume for the fatd database:

```bash
$ docker volume create fatd_db
```

Run fatd:

```bash
$ docker run -d --name=fatd --network=host -v "fatd_db:/fatd.db" fatd [fatd options]
```

## Building From Source

#### Build Dependencies
This project uses SQLite3 which uses [CGo](https://blog.golang.org/c-go-cgo) to
compile and statically link the SQLite3 C libraries to the `fatd` Golang
binary. CGo requires that GCC be available on your system.

The following dependencies are required to build `fatd` and `fat-cli`.
- [Golang](https://golang.org/) 1.13 or later. The latest official release of
  Golang is always recommended.
- [GNU GCC](https://gcc.gnu.org/) is used by
  [CGo](https://blog.golang.org/c-go-cgo) to link to the SQLite3 shared
libraries.
- [Git](https://git-scm.com/) is used to clone the project and is used by `go
  build` to pull some dependencies.
- [GNU Bash](https://www.gnu.org/software/bash/) is used by a small script
  which determines a build version number.
- [GNU Make](https://www.gnu.org/software/make/) is used to execute build
  commands.

#### How To Build
Ensure that Go Modules are enabled by cloning this project *outside* of your
`GOPATH`.

```bash
$ git clone https://github.com/Factom-Asset-Tokens/fatd.git
$ cd fatd
$ make
```
You should now see the `fatd` and `fat-cli` binaries for your platform in the
current directory.

You can also build binaries for all platforms (Linux, Windows, Mac):

```bash
$ make distribution
```

## Installing

You can install and run the `fatd` and `fat-cli` binaries from anywhere, so you
may select where you wish to install these on your system.

However, be aware that `fatd` will look for a directory named `fatd.db` in the
current working directory and will create it if it does not exist. So if you
start `fatd` from a new location then you will likely want to point it to use
an existing database using the `-dbpath` flag.

### CLI completion
If you are using Bash or Zsh with completion you can install CLI completion for
`fatd` and `fat-cli` using the `-installcompletion` flag on each command.
```
$ fatd -installcompletion
Install completion for fatd? y
Installing...
Done!
$ fat-cli --installcompletion
Install completion for fat-cli? y
Installing...
Done!
```
This only adds lines to your `.bash_profile` or Zsh equivalent. So you must
restart your shell or source your `.bash_profile` for the completion to take
effect.

This only needs to be performed once and does not need to be repeated again
when `fatd` or `fat-cli` is updated.



## Getting started

The Daemon needs a connection to `factomd`'s API. This defaults to
`http://localhost:8088` and can be specified with `-s`.

Start the daemon from the command line:
```
INFO Fatd Version: v0.6.0.r110.g73bdb76            pkg=main
INFO Loading chain databases from /home/aslevy/.fatd/mainnet/...  pkg=engine
INFO State engine started.                         pkg=main
INFO Listening on :8078...                         pkg=srv
INFO JSON RPC API server started.                  pkg=main
INFO Factom Asset Token Daemon started.            pkg=main
INFO Searching for new FAT chains from block 163181 to 215642...  pkg=engine
INFO Tracking new FAT chain: b54c4310530dc4dd361101644fa55cb10aec561e7874a7b786ea3b66f2c6fdfb  pkg=engine
INFO Syncing...                                    chain=b54c4310530dc4dd361101644fa55cb10aec561e7874a7b786ea3b66f2c6fdfb
INFO Synced.                                       chain=b54c4310530dc4dd361101644fa55cb10aec561e7874a7b786ea3b66f2c6fdfb
```

At this point `fatd` has synced the first ever created FAT chain on mainnet,
which is for testing purposes. It is continuing to scan for all valid FAT
chains so it can track them.

If you know the chain id of a FAT chain you are interested in, you can fast
sync it by using the `-whitelist` flag.

The daemon can be stopped and restarted and syncing will resume from the latest
point. By default the database directory is `~/.fatd/`. It can be specified
with `-dbdir`.

Once the JSON RPC API is started, `fat-cli` can be used to query about synced
chains, transactions and balances.

For a complete an up-to-date list of flags & options please see `fatd -h` and
`fat-cli -h`.

### Create a chain, make transactions

Interact with the FAT daemon RPC from the command line

[Token Initialization & Transaction Walk Through](docs/ISSUING.md)



## [RPC API Documentation](RPC.md)

Default `http://localhost:8078/v1`



## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md)

