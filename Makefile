all: fatd fat-cli

race: fatd-race fat-cli-race

distribution: fatd-distribution fat-cli-distribution

REVISION     = $(shell ./revision)
FATD_LDFLAGS = "-X github.com/Factom-Asset-Tokens/fatd/flag.Revision=$(REVISION)"
CLI_LDFLAGS  = "-X main.Revision=$(REVISION)"

SRC = go.mod go.sum $(wildcard *.go */*.go */*/*.go)

fatd: $(SRC)
	go build -ldflags=$(FATD_LDFLAGS) ./

fatd-distribution: $(SRC)
	env GOOS=linux GOARCH=amd64 go build -ldflags=$(FATD_LDFLAGS) ./ && env GOOS=windows GOARCH=amd64 go build -ldflags=$(FATD_LDFLAGS) -o fatd.exe ./ && env GOOS=darwin GOARCH=amd64 go build -ldflags=$(FATD_LDFLAGS) -o fatd.app ./

fat-cli: $(SRC)
	go build -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli

fat-cli-distribution: $(SRC)
	env GOOS=linux GOARCH=amd64 go build -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli && env GOOS=windows GOARCH=amd64 go build -ldflags=$(CLI_LDFLAGS) -o fat-cli.exe ./cli && env GOOS=darwin GOARCH=amd64 go build -ldflags=$(CLI_LDFLAGS) -o fat-cli.app ./cli

fatd-race: $(SRC)
	go build -race -ldflags=$(FATD_LDFLAGS) ./

fat-cli-race: $(SRC)
	go build -race -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli

.PHONY: clean purge-db unpurge-db

clean:
	rm -f ./fatd ./fatd.app ./fatd.exe ./fat-cli ./fat-cli.app ./fat-cli.exe

DATE = $(shell date -Ins)
purge-db:
	mv ./fatd.db /tmp/fatd.db.save-$(DATE)

PURGED_DB = $(shell ls /tmp/fatd.db.save-* -d | tail -n 1)
unpurge-db:
	cp -aTn $(PURGED_DB) ./fatd.db

