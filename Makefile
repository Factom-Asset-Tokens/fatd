all: fatd fat-cli

race: fatd-race fat-cli-race

REVISION     = $(shell ./revision)
FATD_LDFLAGS = "-X github.com/Factom-Asset-Tokens/fatd/flag.Revision=$(REVISION)"
CLI_LDFLAGS  = "-X main.Revision=$(REVISION)"

SRC = go.mod go.sum $(wildcard *.go */*.go */*/*.go)

fatd: $(SRC)
	go build -ldflags=$(FATD_LDFLAGS) ./

fat-cli: $(SRC)
	go build -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli

fatd-race: $(SRC)
	go build -race -ldflags=$(FATD_LDFLAGS) ./

fat-cli-race: $(SRC)
	go build -race -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli

.PHONY: clean purge-db unpurge-db

clean:
	rm -f ./fatd ./fat-cli

DATE = $(shell date -Ins)
purge-db:
	mv ./fatd.db /tmp/fatd.db.save-$(DATE)

PURGED_DB = $(shell ls /tmp/fatd.db.save-* -d | tail -n 1)
unpurge-db:
	cp -aTn $(PURGED_DB) ./fatd.db

