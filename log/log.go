package log

import (
	"bitbucket.org/canonical-ledgers/fatd/flag"

	"github.com/sirupsen/logrus"
)

type Log struct {
	*logrus.Entry
}

func New(pkg string) Log {
	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{ForceColors: true,
		DisableTimestamp:       true,
		DisableLevelTruncation: true}
	if flag.LogDebug {
		log.SetLevel(logrus.DebugLevel)
	}
	return Log{Entry: log.WithField("pkg", pkg)}
}
