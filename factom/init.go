package factom

import (
	"github.com/AdamSLevy/factom"
	_log "github.com/Factom-Asset-Tokens/fatd/log"
)

var log _log.Log

// Init sets up the package specific logger and should be called at the
// beginning of any program using this package.
func Init() {
	log = _log.New("factom")
}

// GetHeights returns a struct of Factom Blockchain Heights.
var GetHeights = factom.GetHeights

// RpcConfig is a pointer to the RPC settings.
var RpcConfig = factom.RpcConfig
