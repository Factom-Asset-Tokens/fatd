all: fatd fat-cli

dev: fatd-dev fat-cli-dev

REVISION = $$(./revision)
LDFLAGS	 = "-X main.Revision=$(REVISION)"

CLI_SRC = $(wildcard cli/*.go)
FATD_SRC := $(filter-out $(CLI_SRC), $(wildcard *.go */*.go */*/*.go))

fatd: $(FATD_SRC)
	go build -ldflags=$(LDFLAGS) ./

fat-cli: $(CLI_SRC)
	go build -ldflags=$(LDFLAGS) -o fat-cli ./cli

fatd-dev: $(FATD_SRC)
	go build -ldflags=$(LDFLAGS) -race ./

fat-cli-dev: $(CLI_SRC)
	go build -ldflags=$(LDFLAGS) -race -o fat-cli ./cli

.PHONY: clean purge unpurge

clean:
	rm -f ./fatd ./fat-cli

DATE = $$(date -Ins)
purge: clean
	mv ./fatd.db /tmp/fatd.db.save-$(DATE)

unpurge:
	cp -aTn $$(ls /tmp/fatd.db.save-* -d | tail -n 1) ./fatd.db

