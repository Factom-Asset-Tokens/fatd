module github.com/Factom-Asset-Tokens/fatd

go 1.12

require (
	crawshaw.io/sqlite v0.1.3-0.20190520153332-66f853b01dfb
	github.com/AdamSLevy/go-merkle v0.0.0-20190611101253-ca33344a884d
	github.com/AdamSLevy/jsonrpc2/v11 v11.3.2
	github.com/Factom-Asset-Tokens/base58 v0.0.0-20181227014902-61655c4dd885
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/gocraft/dbr v0.0.0-20190131145710-48a049970bd2
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jinzhu/gorm v1.9.4
	github.com/jinzhu/now v1.0.0 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/posener/complete v1.2.1
	github.com/rs/cors v1.6.0
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.1-0.20190311161405-34c6fa2dc709
	golang.org/x/crypto v0.0.0-20190426145343-a29dc8fdc734
)

replace github.com/gocraft/dbr => github.com/AdamSLevy/dbr v0.0.0-20190429075658-5db28ac75cea

replace github.com/spf13/pflag v1.0.3 => github.com/AdamSLevy/pflag v1.0.4
