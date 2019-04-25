module github.com/Factom-Asset-Tokens/fatd

go 1.12

require (
	cloud.google.com/go v0.30.0 // indirect
	github.com/AdamSLevy/jsonrpc2/v11 v11.3.1
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/DATA-DOG/go-sqlmock v1.3.0 // indirect
	github.com/Factom-Asset-Tokens/base58 v0.0.0-20181227014902-61655c4dd885
	github.com/denisenkom/go-mssqldb v0.0.0-20180901172138-1eb28afdf9b6 // indirect
	github.com/erikstmartin/go-testdb v0.0.0-20160219214506-8d10e4a1bae5 // indirect
	github.com/gocraft/dbr v0.0.0-20190131145710-48a049970bd2
	github.com/gofrs/uuid v3.2.0+incompatible // indirect
	github.com/google/go-cmp v0.2.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/jinzhu/now v0.0.0-20180511015916-ed742868f2ae // indirect
	github.com/jmoiron/sqlx v1.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/posener/complete v1.1.2
	github.com/rs/cors v1.3.0
	github.com/sirupsen/logrus v1.1.1
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190325154230-a5d413f7728c
	golang.org/x/sys v0.0.0-20190322080309-f49334f85ddc // indirect
	google.golang.org/appengine v1.2.0 // indirect
)

replace github.com/gocraft/dbr => github.com/AdamSLevy/dbr v0.0.0-20190204220409-06403a96499f

replace github.com/spf13/pflag v1.0.3 => github.com/AdamSLevy/pflag v1.0.4
