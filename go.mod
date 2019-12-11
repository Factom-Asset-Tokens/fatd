module github.com/Factom-Asset-Tokens/fatd

go 1.13

require (
	crawshaw.io/sqlite v0.2.1-0.20191210213818-c9d1693acfe9
	github.com/AdamSLevy/jsonrpc2/v13 v13.0.1
	github.com/AdamSLevy/sqlbuilder v0.0.0-20191126201320-5b1948d48973
	github.com/AdamSLevy/sqlitechangeset v0.0.0-20191210201651-f95453d87aff
	github.com/Factom-Asset-Tokens/factom v0.0.0-20191207221317-3707342f4ae9
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
	github.com/wasmerio/go-ext-wasm v0.0.0-20191206132826-225d01fcd22c
)

replace (
	// Fixes a small annoyance when displaying defaults for custom Vars.
	github.com/spf13/pflag v1.0.5 => github.com/AdamSLevy/pflag v1.0.6-0.20191204180553-73c85c9446e1

	// Uses the LLVM backend and exposes metering. Also fixes a bug with
	// InstanceContext.Data.
	//github.com/wasmerio/go-ext-wasm => github.com/AdamSLevy/go-ext-wasm v0.0.0-20191212063545-ec82ca05412e
	github.com/wasmerio/go-ext-wasm => github.com/AdamSLevy/go-ext-wasm v0.0.0-20191212234502-d66004a8582c
)

//replace github.com/Factom-Asset-Tokens/factom => ../factom

//replace crawshaw.io/sqlite => /home/aslevy/repos/go-modules/AdamSLevy/sqlite

//replace github.com/AdamSLevy/jsonrpc2/v12 => /home/aslevy/repos/go-modules/AdamSLevy/jsonrpc2

//replace github.com/wasmerio/go-ext-wasm => /home/aslevy/repos/go-ext-wasm
