package state

import (
	_ "bitbucket.org/canonical-ledgers/fatd/db"
	_ "bitbucket.org/canonical-ledgers/fatd/factom"
	_ "bitbucket.org/canonical-ledgers/fatd/flag"
)

func Start() error {
	setupLogger()
	return nil
}
