![](https://png.icons8.com/ios-glyphs/200/5ECCDD/octahedron.png)![](https://png.icons8.com/color/64/3498db/golang.png)

# fatd - Factom Asset Token Daemon

A daemon written in Golang that maintains the current state of Factom Asset Tokens(FAT) tokens.

Provides a standard RPC API to access FAT data.



## Installing

Installing & running fatd requires [Golang 1.1](https://golang.org/dl/)

```bash
git clone https://github.com/Factom-Asset-Tokens/fatd.git
cd fatd
go get
```



## Running

From the command line:

```bash
go run main.go
```



### Run As A Background Service



## Updating

```bash
git pull
go run main.go
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