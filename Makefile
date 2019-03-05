all: fatd fat-cli

dev: fatd-dev fat-cli-dev

clean:
	rm -f ./fatd ./fat-cli

DATE = $$(date -Iseconds)

purge: clean
	mv ./fatd.db /tmp/fatd.db.save-$(DATE)

fatd:
	go build ./

fat-cli:
	go build -o fat-cli ./cli

fatd-dev:
	go build -race ./

fat-cli-dev:
	go build -race -o fat-cli ./cli

.PHONY: clean purge
