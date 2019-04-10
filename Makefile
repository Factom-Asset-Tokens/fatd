all: fatd fat-cli

race: fatd.race fat-cli.race

distribution: fatd-distribution fat-cli-distribution

fatd-distribution: fatd.app fatd.exe fatd-linux
fat-cli-distribution: fat-cli.app fat-cli.exe fat-cli-linux

REVISION     = $(shell ./revision)
FATD_LDFLAGS = "-X github.com/Factom-Asset-Tokens/fatd/flag.Revision=$(REVISION)"
CLI_LDFLAGS  = "-X main.Revision=$(REVISION)"

DEPSRC = go.mod go.sum
SRC = $(DEPSRC) $(filter-out %_test.go,$(wildcard *.go */*.go */*/*.go))

fatd-race: $(FATDSRC)
	go build -race -ldflags=$(FATD_LDFLAGS) ./

fat-cli-race: $(CLISRC)
	go build -race -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli

GENSRC=factom/idkey_gen.go factom/idkey_gen_test.go

FATDSRC=$(filter-out cli/%,$(SRC)) $(GENSRC)
fatd: $(FATDSRC)
	go build -ldflags=$(FATD_LDFLAGS) ./

CLISRC=$(filter-out main.go engine/% state/% flag/%,$(SRC)) $(GENSRC)
fat-cli: $(CLISRC)
	go build -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli


fatd.app: $(FATDSRC)
	env GOOS=darwin GOARCH=amd64 go build -ldflags=$(FATD_LDFLAGS) -o fatd.app ./

fatd.exe: $(FATDSRC)
	env GOOS=windows GOARCH=amd64 go build -ldflags=$(FATD_LDFLAGS) -o fatd.exe ./

fatd-linux: $(FATDSRC)
	env GOOS=linux GOARCH=amd64 go build -ldflags=$(FATD_LDFLAGS) ./

fat-cli.app: $(CLISRC)
	env GOOS=darwin GOARCH=amd64 go build -ldflags=$(CLI_LDFLAGS) -o fat-cli.app ./

fat-cli.exe: $(CLISRC)
	env GOOS=windows GOARCH=amd64 go build -ldflags=$(CLI_LDFLAGS) -o fat-cli.exe ./

fat-cli-linux: $(CLISRC)
	env GOOS=linux GOARCH=amd64 go build -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli

$(GENSRC): factom/gen.go  factom/genmain.go $(wildcard factom/*.tmpl)
	go generate ./factom


.PHONY: clean purge-db unpurge-db

clean:
	rm -f ./fatd ./fatd.app ./fatd.exe ./fat-cli ./fat-cli.app ./fat-cli.exe $(GENSRC)

DATE = $(shell date -Ins)
purge-db:
	mv ./fatd.db /tmp/fatd.db.save-$(DATE)

PURGED_DB = $(shell ls /tmp/fatd.db.save-* -d | tail -n 1)
unpurge-db:
	cp -aTn $(PURGED_DB) ./fatd.db

