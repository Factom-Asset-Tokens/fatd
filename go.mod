module github.com/Factom-Asset-Tokens/fatd

go 1.13

require (
	crawshaw.io/sqlite v0.2.5
	github.com/AdamSLevy/jsonrpc2/v14 v14.0.0
	github.com/AdamSLevy/sqlbuilder v0.0.0-20191126201320-5b1948d48973
	github.com/AdamSLevy/sqlitechangeset v0.0.0-20191210201651-f95453d87aff
	github.com/Factom-Asset-Tokens/base58 v0.0.0-20191118025050-4fa02e92ec20 // indirect
	github.com/Factom-Asset-Tokens/factom v0.0.0-20200218003939-f06d3b5f187a
	github.com/goji/httpauth v0.0.0-20160601135302-2da839ab0f4d
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nightlyone/lockfile v0.0.0-20180618180623-0ad87eef1443
	github.com/posener/complete v1.2.1
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.5.0
	github.com/stretchr/testify v1.4.0
	github.com/subchen/go-trylock/v2 v2.0.0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
)

// Fixes a small annoyance when displaying defaults for custom Vars.
replace github.com/spf13/pflag v1.0.5 => github.com/AdamSLevy/pflag v1.0.6-0.20191204180553-73c85c9446e1

//replace github.com/Factom-Asset-Tokens/factom => ../factom
