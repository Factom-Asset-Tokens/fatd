![](https://png.icons8.com/ios-glyphs/200/5ECCDD/octahedron.png)![](https://png.icons8.com/color/64/3498db/golang.png)

# fatd - Factom Asset Token Daemon

A daemon written in Golang that maintains the current state of Factom Asset
Tokens (FAT) tokens.

Provides a standard RPC API to access FAT data.



## Building

Installing & running fatd requires [Golang 1.11](https://golang.org/dl/) or later.

```bash
$ git clone https://github.com/Factom-Asset-Tokens/fatd.git
$ cd fatd
$ go build
$ ./fatd
```



## Installing

The `fatd` binary can be run from anywhere in your system. If `$GOPATH/bin` is
in your `PATH` then you can use `go install` from inside the build directory to
move `fatd` to that location and then run `fatd`.

### Systemd Service

TODO: Later there will be a systemd service to run fatd and further
instructions on how to set that up should go here.



## Running

From the command line:

```bash
$ fatd
```

TODO: Add example of some common flags that people need to use.



### Run As A Background Service

TODO: Sort out this section with the section in Installing above.



## Updating

```bash
git pull
go build
```



## Flags

- `-flag` - Value



## Config

```json
{
    "configparam":"value"
}
```



## RPC Endpoint

### `http://localhost:8078/v0`



## [RPC Reference](https://github.com/Factom-Asset-Tokens/FAT/blob/FATIP-300-FAT-RPC-API-Standard/fatips/300.md)
