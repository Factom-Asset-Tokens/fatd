package factom

import (
	"bitbucket.org/canonical-ledgers/fatd/flag"

	"github.com/sirupsen/logrus"
)

func init() {
	setupLogger()
}

var log *logrus.Entry

func setupLogger() {
	_log := logrus.New()
	_log.Formatter = &logrus.TextFormatter{ForceColors: true,
		DisableTimestamp:       true,
		DisableLevelTruncation: true}
	if flag.LogDebug {
		_log.SetLevel(logrus.DebugLevel)
	}
	log = _log.WithField("pkg", "factom")
}
