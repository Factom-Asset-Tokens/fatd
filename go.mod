module github.com/Factom-Asset-Tokens/fatd

go 1.13

require (
	crawshaw.io/sqlite v0.1.3-0.20190520153332-66f853b01dfb
	github.com/AdamSLevy/jsonrpc2/v12 v12.0.2-0.20191015223217-9181d6ac9347
	github.com/AdamSLevy/sqlitechangeset v0.0.0-20190925183646-3ddb70fb709d
	github.com/Factom-Asset-Tokens/factom v0.0.0-20191114223543-dab6c2e7759c
	github.com/goji/httpauth v0.0.0-20160601135302-2da839ab0f4d
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nightlyone/lockfile v0.0.0-20180618180623-0ad87eef1443
	github.com/posener/complete v1.2.1
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
)

replace github.com/spf13/pflag v1.0.3 => github.com/AdamSLevy/pflag v1.0.4

replace crawshaw.io/sqlite => github.com/AdamSLevy/sqlite v0.1.3-0.20191014215059-b98bb18889de

//replace github.com/Factom-Asset-Tokens/factom => ../factom

//replace crawshaw.io/sqlite => /home/aslevy/repos/go-modules/AdamSLevy/sqlite

//replace github.com/AdamSLevy/jsonrpc2/v12 => /home/aslevy/repos/go-modules/AdamSLevy/jsonrpc2
