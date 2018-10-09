package srv

import (
	_log "bitbucket.org/canonical-ledgers/fatd/log"
)

var (
	log _log.Log
)

func Start() error {
	log = _log.New("srv")
	return nil
}

func Stop() error {
	return nil
}
