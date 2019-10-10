# MIT License
#
# Copyright 2018 Canonical Ledgers, LLC
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

all: fatd fat-cli

race: fatd.race fat-cli.race

distribution: fatd-distribution fat-cli-distribution

fatd-distribution: fatd.mac fatd.exe fatd-linux
fat-cli-distribution: fat-cli.mac fat-cli.exe fat-cli-linux

REVISION     = $(shell ./revision)

export GOFLAGS
GOFLAGS = -gcflags=all=-trimpath=${PWD} -asmflags=all=-trimpath=${PWD}

GO_LDFLAGS   = -extldflags=$(LDFLAGS) -X github.com/Factom-Asset-Tokens/fatd
FATD_LDFLAGS = "$(GO_LDFLAGS)/flag.Revision=$(REVISION)"
CLI_LDFLAGS  = "$(GO_LDFLAGS)/cli/cmd.Revision=$(REVISION)"

DEPSRC = go.mod go.sum
SRC = $(DEPSRC) $(filter-out %_test.go,$(wildcard *.go */*.go */*/*.go))

FATDSRC=$(filter-out cli/%,$(SRC)) $(GENSRC)
fatd: $(FATDSRC)
	go build -trimpath -ldflags=$(FATD_LDFLAGS) ./

CLISRC=$(filter-out main.go engine/% state/% flag/%,$(SRC)) $(GENSRC)
fat-cli: $(CLISRC)
	go build -trimpath -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli


fatd.race: $(FATDSRC)
	go build -trimpath -race -ldflags=$(FATD_LDFLAGS) -o fatd.race ./

fat-cli.race: $(CLISRC)
	go build -trimpath -race -ldflags=$(CLI_LDFLAGS) -o fat-cli.race ./cli


fatd.mac: $(FATDSRC)
	env GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags=$(FATD_LDFLAGS) -o fatd.mac ./

fatd.exe: $(FATDSRC)
	env GOOS=windows GOARCH=amd64 go build -trimpath -ldflags=$(FATD_LDFLAGS) -o fatd.exe ./

fatd-linux: $(FATDSRC)
	env GOOS=linux GOARCH=amd64 go build -trimpath -ldflags=$(FATD_LDFLAGS) ./

fat-cli.mac: $(CLISRC)
	env GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags=$(CLI_LDFLAGS) -o fat-cli.mac ./

fat-cli.exe: $(CLISRC)
	env GOOS=windows GOARCH=amd64 go build -trimpath -ldflags=$(CLI_LDFLAGS) -o fat-cli.exe ./

fat-cli-linux: $(CLISRC)
	env GOOS=linux GOARCH=amd64 go build -trimpath -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli

.PHONY: clean clean-gen purge-db unpurge-db

clean:
	rm -f ./fatd ./fatd.mac ./fatd.exe ./fat-cli ./fat-cli.mac ./fat-cli.exe ./fatd.race ./fat-cli.race

clean-gen:
	rm -f $(GENSRC)

DATE = $(shell date -Ins)
purge-db:
	mv ./fatd.db /tmp/fatd.db.save-$(DATE)

PURGED_DB = $(shell ls /tmp/fatd.db.save-* -d | tail -n 1)
unpurge-db:
	cp -aTn $(PURGED_DB) ./fatd.db

