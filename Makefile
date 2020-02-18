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

.PHONY: all fatd fat-cli fatd.race fat-cli.race clean

all: fatd fat-cli

race: fatd.race fat-cli.race

distribution: fatd-distribution fat-cli-distribution

fatd-distribution: fatd.mac fatd.exe fatd-linux
fat-cli-distribution: fat-cli.mac fat-cli.exe fat-cli-linux

REVISION     := $(shell sh -c "./revision")

GO_LDFLAGS   = -extldflags=$(LDFLAGS)
FATD_LDFLAGS = "$(GO_LDFLAGS) -X github.com/Factom-Asset-Tokens/fatd/internal/flag.Revision=$(REVISION)"
CLI_LDFLAGS  = "$(GO_LDFLAGS) -X main.Revision=$(REVISION)"

fatd:
	go build -trimpath -ldflags=$(FATD_LDFLAGS) ./

fat-cli:
	go build -trimpath -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli


fatd.race:
	go build -race -trimpath -ldflags=$(FATD_LDFLAGS) ./

fat-cli.race:
	go build -race -trimpath -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli


fatd.mac:
	env GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags=$(FATD_LDFLAGS) -o fatd.mac ./

fatd.exe:
	env GOOS=windows GOARCH=amd64 go build -trimpath -ldflags=$(FATD_LDFLAGS) -o fatd.exe ./

fatd-linux:
	env GOOS=linux GOARCH=amd64 go build -trimpath -ldflags=$(FATD_LDFLAGS) ./

fat-cli.mac:
	env GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags=$(CLI_LDFLAGS) -o fat-cli.mac ./

fat-cli.exe:
	env GOOS=windows GOARCH=amd64 go build -trimpath -ldflags=$(CLI_LDFLAGS) -o fat-cli.exe ./

fat-cli-linux:
	env GOOS=linux GOARCH=amd64 go build -trimpath -ldflags=$(CLI_LDFLAGS) -o fat-cli ./cli

clean:
	rm -f ./fatd ./fat-cli ./fatd.{mac,exe,race} ./fat-cli.{mac,exe,race}
