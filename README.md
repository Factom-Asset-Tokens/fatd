![](https://png.icons8.com/ios-glyphs/200/5ECCDD/octahedron.png)![](https://png.icons8.com/color/64/3498db/golang.png)

# fatd - Factom Asset Token Daemon

A daemon written in Golang that maintains the current state of Factom Asset
Tokens (FAT) tokens.

Provides a standard RPC API to access FAT data.



## Building

#### Build Dependencies
This project uses SQLite3 which uses [CGo](https://blog.golang.org/c-go-cgo) to
dynamically link to the SQLite3 C shared libraries to the `fatd` Golang binary.
CGo requires that GCC be available on your system.

The following dependencies are required to build `fatd` and `fat-cli`.
- [Golang](https://golang.org/) 1.11 or later. The latest official release of
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
You should now see the `fatd` and `fat-cli` binaries in the current directory.

## Installing

TODO

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



## Default RPC Endpoint

`http://localhost:8078/v1`

## [RPC Reference](RPC.md)



## Troubleshooting
TODO
