![](https://png.icons8.com/ios-glyphs/200/5ECCDD/octahedron.png)![](https://png.icons8.com/color/64/3498db/golang.png)

# fatd - Factom Asset Token Daemon v0.4.2 - Alpha

A daemon written in Golang that maintains the current state of Factom Asset
Tokens (FAT) tokens. Includes a CLI for interacting with the FAT Daemon from
the command line.

Provides a standard RPC API to access FAT data.

## Development Status

The FAT protocol and this implementation is still in Alpha. That means that we
are still testing and making changes to the protocol and the implementation.
The on-chain data protocol is relatively stable but this implementation and the
database schema is not.

So long as the major version is v0 everything is subject to potential change.

At times new v0 releases may require you to rebuild your fatd.db database from
scratch, but this will be minimized after the next major release due to an
improved migration framework and database validation on startup.

Please help us improve the code and the protocol by trying to break things and
reporting bugs! Thank you!


## Binaries

Pre-compiled binaries for Linux, Windows, and Mac x86\_64 systems can be found
on the [releases page.](https://github.com/Factom-Asset-Tokens/fatd/releases/)



## Building From Source

#### Build Dependencies
This project uses SQLite3 which uses [CGo](https://blog.golang.org/c-go-cgo) to
dynamically link to the SQLite3 C shared libraries to the `fatd` Golang binary.
CGo requires that GCC be available on your system.

The following dependencies are required to build `fatd` and `fat-cli`.
- [Golang](https://golang.org/) 1.11.4 or later. The latest official release of
  Golang is always recommended.
- [GNU GCC](https://gcc.gnu.org/) is used by
  [CGo](https://blog.golang.org/c-go-cgo) to link to the SQLite3 shared
libraries.
- [SQLite3](https://sqlite.org/index.html) is the database `fatd` uses to save
  state.
- [Git](https://git-scm.com/) is used to clone the project and is used by `go
  build` to pull some dependencies.
- [Bazaar VCS](https://bazaar.canonical.com/en/) is used by `go build` to pull
  some dependencies.
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
$ fat-cli -installcompletion
Install completion for fat-cli? y
Installing...
Done!
```
This only adds lines to your `.bash_profile` or Zsh equivalent. So you must
restart your shell or source your `.bash_profile` for the completion to take
effect.

This only needs to be performed once and does not need to be repeated again
when `fatd` or `fat-cli` is updated.



## Running
Start the daemon from the command line:
```
$ ./fatd
INFO Fatd Version: r155.c812dd1                    pkg=main
INFO State engine started.                         pkg=main
INFO JSON RPC API server started.                  pkg=main
INFO Factom Asset Token Daemon started.            pkg=main
INFO Syncing from block 183396 to 183520...        pkg=engine
INFO Synced.                                       pkg=engine
```

### Exiting
To tell `fatd` to safely exit send a `SIGINT`. From most shells this can be
done by simply pressing `CTRL`+`c`.
```
INFO Synced.                                       pkg=engine
^CINFO SIGINT: Shutting down now.                    pkg=main
INFO Factom Asset Token Daemon stopped.            pkg=main
INFO JSON RPC API server stopped.                  pkg=main
INFO State engine stopped.                         pkg=main
$
```



## Startup Flags & Options

Control how fatd runs using additional options at startup. For example:

```bash
./fatd -debug -dbpath /home/ubuntu/mycustomfolder
```



| Name              | Description                                                  | Validation               | Default                   |
| ----------------- | ------------------------------------------------------------ | ------------------------ | ------------------------- |
| `startscanheight` | The Factom block height to begin scanning for new FAT chains and transactions | Positive Integer         | 0                         |
| `debug`           | Enable debug mode for extra information during runtime. No value needed. | -                        | -                         |
| `dbpath`          | Specify the path to use as fatd's sqlite database.           | Valid system path        | Current working directory |
| `ecpub`           | The public Entry Credit address used to pay for submitting transactions | Valid EC address         | -                         |
| `apiaddress`      | What port string the FAT daemon RPC will be bound to         | String                   | `:8078`                   |
|                   |                                                              |                          |                           |
| `s`               | The URL of the Factom API host                               | Valid URL                | `localhost:8088`          |
| `factomdtimeout`  | The timeout in seconds to time out requests to factomd       | integer                  | 0                         |
| `factomduser`     | The username of the user for factomd API authentication      | string                   | -                         |
| `factomdpassword` | The password of the user for factomd API authentication      | string                   | -                         |
| `factomdcert`     | Path to the factomd connection TLS certificate file          | Valid system path string | -                         |
| `factomdtls`      | Whether to use TLS on connection to factomd                  | boolean                  | false                     |
|                   |                                                              |                          |                           |
| `w`               | The URL of the Factom Wallet Daemon API host                 | Valid URL                | `localhost:8089`          |
| `wallettimeout`   | The timeout in seconds to time out requests to factomd       | integer                  | 0                         |
| `walletuser`      | The username of the user for walletd API authentication      | string                   | -                         |
| `walletpassword`  | The username of the user for walletd API authentication      | string                   | -                         |
| `walletcert`      | Path to the walletd connection TLS certificate file          | Valid system path string | -                         |
| `wallettls`       | Whether to use TLS on connection to walletd                  | boolean                  | false                     |

For a complete up to date list of flags & options please see `flag/flag.go`



## [FAT CLI Documentation](CLI.md)

Interact with the FAT daemon RPC from the command line

[Token Initialization & Transaction Walk Through](docs/ISSUING.md)



## [RPC API Documentation](RPC.md)

Default `http://localhost:8078/v1`




## Contributing

All PRs should be rebased on the latest `develop` branch commit.

## Issues

Please attempt to reproduce the issue using the `-debug` flag. For `fatd`,
please provide the initial output which prints all current settings.
Intermediate `DEBUG Scanning block 187682 for FAT entries.` lines may be
omitted, but please provide the first and last of these lines.

