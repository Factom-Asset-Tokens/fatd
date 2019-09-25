module github.com/Factom-Asset-Tokens/fatd

go 1.13

require (
	crawshaw.io/sqlite v0.1.3-0.20190520153332-66f853b01dfb
	github.com/AdamSLevy/jsonrpc2/v11 v11.3.2
	github.com/AdamSLevy/sqlitechangeset v0.0.0-20190925183646-3ddb70fb709d
	github.com/Factom-Asset-Tokens/factom v0.0.0-20190911201853-7b283996f02a
	github.com/goji/httpauth v0.0.0-20160601135302-2da839ab0f4d
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nightlyone/lockfile v0.0.0-20180618180623-0ad87eef1443
	github.com/posener/complete v1.2.1
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	golang.org/x/sys v0.0.0-20190911201528-7ad0cfa0b7b5 // indirect
)

replace github.com/spf13/pflag v1.0.3 => github.com/AdamSLevy/pflag v1.0.4

replace crawshaw.io/sqlite => github.com/AdamSLevy/sqlite v0.1.3-0.20190925021041-5f14912819e9
